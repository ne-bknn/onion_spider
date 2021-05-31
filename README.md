# Onion Spider
[![MIT License](https://img.shields.io/apm/l/atomic-design-ui.svg?)](https://github.com/tterb/atomic-design-ui/blob/master/LICENSEs)

That's a self-hosted link collector and search engine on top of it. It is built in Go using [Colly](https://github.com/gocolly/colly) to do heavy lifting, Redis as a task queue and [MeiliSearch](https://github.com/meilisearch/MeiliSearch) as a search engine.


## Deployment
 
Deploy redis and meilisearch by running

```bash
  docker-compose up
```

Build parser by running
```bash
  cd parser
  go build 
```

Launch parser with
```bash
  ./parser http://entrypoint.onion/
```

You'll get status on stdout and you can view and search results on https://localhost:7700


