# BIOSCHEMAS.ORG GO CRAWL IT!

Crawls and extracts bioschemas.org/schema.org JSON-LD and Microdata from a given website. The extracted information is stored on a JSON file and optionally can be stored on a Elasticsearch local service.


## How to use it:
---
Use example: 
```bash
./bioschemas-gocrawlit -p -u "https://www.ebi.ac.uk/biosamples/samples" -q -query "start"
./bioschemas-gocrawlit_mac_64 -q -u https://tess.elixir-europe.org/sitemaps/events.xml
./bioschemas-gocrawlit_mac_64 -u http://159.149.160.88/pscan_chip_dev/
```

A folder "bioschemas_gocrawlit_cache" will be created on the current path of execution; This folder contains crawled website information in order to prevent multiple download of pages. Is safe to delete this folder.


### Output

Scraped data will be stored in a json file named ```<website_host>_schema.json``` on the current program folder.


### Available commands

- **-p**: Stay on current path. i.e. When crawling a page like ```https://www.ebi.ac.uk/biosamples/samples``` and don't want it to crawl the whole website, e.g. ```https://www.ebi.ac.uk```.
- **-m**: Max number of recursion depth of visited URLs. Default infinity recursion. (The crawler does not revisit URLs)
- **-e**: Adds crawled data to an Elasticsearch (v6) service at http://127.0.0.1:9200.
- **-u**: Start page to start crawling.
- **-q**: Remove query section from the link URL found.
- **--query**: Use with **-q** so it follows only links that contain the query word provided, e.g., ```./bioschemas-gocrawlit_mac_64 -u https://tess.elixir-europe.org/events -q --page page```
- **-h**: Print Help and exit.


## Building binaries
----
To create a binary for your current SO use:
```bash
make build
```

To create a binary for windows, macos and linux SO use:
```bash
make build-all
```

The binaries would be placed under build/ path.


## Elasticsearch quick setup [DOCKER](https://www.docker.com/)
---
Steps for starting dockerized [elasticsearch](https://www.elastic.co/products/elasticsearch) and [kibana](https://www.elastic.co/products/kibana) locally. This requires [Docker](https://store.docker.com/search?type=edition&offering=community).

#### Create a custom network for your elastic-stack:

```docker network create elastic-stack```

#### Pull and run an elasticsearch image:

```docker run -it --network=elastic-stack -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" --name elasticsearch docker.elastic.co/elasticsearch/elasticsearch:6.2.4```
> Avoid changing the containers name since Kibana docker image points by default to `http://elasticsearch:9200`.

#### Pull and run an elasticsearch image:

```docker run --network=elastic-stack --rm -it -p 5601:5601 --name kibana docker.elastic.co/kibana/kibana:6.2.4```

> Remember the --rm flag will delete the container once it is stoped.


## ToDo

- [x] Crawl website
- [x] URL by command line parameters
- [x] JSON-LD Extraction
- [x] Microdata extraction
- [x] Better file output
- [x] Sitemap.xml Crawl option
- [x] Pagination option
- [x] Conecting to a flexible storage
- [ ] RDFa extraction support
- [x] Writing file as it scraps
