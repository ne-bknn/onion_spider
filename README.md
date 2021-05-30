# Onion Spider
[![MIT License](https://img.shields.io/apm/l/atomic-design-ui.svg?)](https://github.com/tterb/atomic-design-ui/blob/master/LICENSEs)

That's a self-hosted link collector and search engine on top of it. It is built in Go, using Redis as task queue and MeiliSearch as search engine.




## Deployment

Fill `starting_list.txt` with domains of your choice.

Modify `config.ini` with values of your choice (how deep to go) 

Deploy by running

```bash
  docker-compose up
```

You'll get status on stdout and you can view and search results on https://localhost:7700
