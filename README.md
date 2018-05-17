# BIOSCHEMAS.ORG GO CRAWL IT!

Minimal crawler and extractor of microdata and JSON-LD metadata.


## How to use it:

Use example: 
```
./bioschemas-gocrawlit -p -u "https://www.ebi.ac.uk/biosamples/samples"
./bioschemas-gocrawlit -q -u https://tess.elixir-europe.org/sitemaps/events.xml
./bioschemas-gocrawlit -u http://159.149.160.88/pscan_chip_dev/
```

A folder "bioschemas_gocrawlit_cache" will be created on the current path of execution;This folder contains crawled website information in order to prevent multiple download of pages. Is safe to delete this folder.

### Output

Scraped data will be stored in a json file named ```<website_host>_schema.json``` on the current program folder.


### Available commands

- **-p**: Stay on current path. i.e. When crawling a page like ```https://www.ebi.ac.uk/biosamples/samples``` and don't want it to crawl the whole website, e.g. ```https://www.ebi.ac.uk```.
- **-m**: Max number of recursion depth of visited URLs. Default infinity recursion. (The crawler does not revisit URLs)
- **-e**: Adds crawled data to an Elasticsearch (v6) service at http://127.0.0.1:9200.
- **-u**: Start page to start crawling.
- **-q**: Remove query section from the link URL found.
- **--query**: Use with **-q** so it follows only links that contain the query word provided, e.g., ```./bioschemas-gocrawlit -u https://tess.elixir-europe.org/events -q --page page```
- **-h**: Print Help and exit.


## Building binaries

To create a binary for your current SO use:
```make build```

To create a binary for windows, macos and linux SO use:
```make build-all```

The binaries would be placed under build/ path.


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
