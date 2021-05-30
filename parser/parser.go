package main

import (
	"github.com/gocolly/colly"
	"github.com/gocolly/redisstorage"


	"github.com/go-redis/redis/v8"
	"github.com/meilisearch/meilisearch-go"

	"log"
	"fmt"
	"context"
	//"regexp"
	"net/url"
	"strings"
)

type DB struct {
	conn *redis.Client
}

var ctx = context.Background()

func (d *DB) init() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
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
	var meiliClient = meilisearch.NewClient(meilisearch.Config{
		Host: "http://search:7700",
		APIKey: "somerandomkey",
	})

	_, err := meiliClient.Indexes().Create(meilisearch.CreateIndexRequest{
		UID: "domains",
	})

	if err != nil {
		log.Print(err)
	}

	cache := DB{}
	cache.init()

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"),
		//colly.URLFilters(regexp.MustCompile(".*\\.onion$")),
		colly.Async(true),
		colly.CacheDir("./cache"),
	)

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 16})

	storage := &redisstorage.Storage{
		Address:  "redis:6379",
		Password: "",
		DB:       0,
		Prefix:   "job01",
	}

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

		if !cache.contains_domain(hostname) {
			fmt.Printf("%s: %s\n", baseurl.String(), e.Text)
			docs := []map[string]interface{}{
				{
					"domain": baseurl.String(),
					"title": e.Text,
				},
			}
			_, err := meiliClient.Documents("domains").AddOrUpdate(docs)

			if err != nil {
				log.Fatal(err)
			}

			cache.add_domain(hostname)
		}
	})

	c.Visit("http://httpbin.org/")
	c.Wait()

	defer storage.Client.Close()
}
