package main

import (
	"github.com/gocolly/colly"
	"github.com/gocolly/redisstorage"


	"github.com/go-redis/redis/v8"
	"github.com/meilisearch/meilisearch-go"

	"log"
	"fmt"
	"context"
	"regexp"
	"net/url"
	"strings"
	"time"
	"os"
)

type DB struct {
	conn *redis.Client
}

var ctx = context.Background()

func (d *DB) init() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",
		DB: 0,
	})

	d.conn = rdb
}

func (d *DB) add_domain(Domain string) {
	d.conn.SAdd(ctx, "onsp_domains", Domain)
}

func (d *DB) contains_domain(Domain string) (bool) {
	return d.conn.SIsMember(ctx, "onsp_domains", Domain).Val()
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Please provide entrypoint as first argument")
	}

	entrypoint := os.Args[1]

	var meiliClient = meilisearch.NewClient(meilisearch.Config{
		Host: "http://localhost:7700",
	})

	_, err := meiliClient.Indexes().Get("domains")
	if err != nil {
		_, err := meiliClient.Indexes().Create(meilisearch.CreateIndexRequest{
			UID: "domains",
			PrimaryKey: "id",
		})

		if err != nil {
			log.Print(err)
		}
	}

	cache := DB{}
	cache.init()

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"),
		colly.URLFilters(regexp.MustCompile(".*\\.onion.*")),
		colly.Async(true),
		colly.CacheDir("./cache"),
	)

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2})

	storage := &redisstorage.Storage{
		Address:  "localhost:6379",
		Password: "",
		DB:       0,
		Prefix:   "job01",
	}

	c.SetProxy("socks5://localhost:9050")

	err = c.SetStorage(storage)
	if err != nil {
		panic(err)
	}

	if err := storage.Clear(); err != nil {
		log.Fatal(err)
	}

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	c.OnHTML("title", func(e *colly.HTMLElement) {
		aurl := e.Request.AbsoluteURL(e.Request.URL.String())
		u, err := url.Parse(aurl)
		if err != nil {
			fmt.Printf("This URL caused url.Parse to panic: %s", u)
		}

		hostname := u.Hostname()

		var baseurl strings.Builder
		baseurl.WriteString("http://")
		baseurl.WriteString(hostname)

		id := strings.ReplaceAll(hostname, ".", "_")

		if !cache.contains_domain(hostname) {
			fmt.Printf("%s, %s, %s\n\n", id, hostname, e.Text)
			docs := []map[string]interface{}{
				{
					"id": id,
					"hostname": hostname,
					"title": e.Text,
				},
			}
			_, err := meiliClient.Documents("domains").AddOrUpdate(docs)

			c := 0

			for err != nil {
				fmt.Println("Retrying document insertion")
				c += 1
				if c > 5 {
					log.Fatal(err)
				}
				time.Sleep(3)
				_, err = meiliClient.Documents("domains").AddOrUpdate(docs)
			}

			fmt.Printf("Succeeded document insertion in %d attempts", c)

			cache.add_domain(hostname)
		}
	})

	c.Visit(entrypoint)
	c.Wait()
	fmt.Println("My work is done here")
	defer storage.Client.Close()
}
