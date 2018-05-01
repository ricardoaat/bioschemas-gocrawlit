# BIOSCHEMAS.ORG GO CRAWL IT!

Minimal web crawler. Extracts microdata and JSON-LD metadata with schema.org.

### ToDo

- [x] Crawl website
- [x] URL by command line parameters
- [x] Extracts JSON-LD
- [x] Extracts microdata
- [ ] JSON-LD schema.org check
- [ ] Better file output

## How to use it:

Run it like this: 
```
./bioschemas-gocrawlit -u "https://identifiers.org"
```

## Help

Use the -h parameter to get info about the command tool.

```./bioschemas-gocrawlit -h```


## Building binaries

To create a binary for your current SO use:
```make build```

To create a binary for windows, macos and linux SO use:
```make build-all```

The binaries would be found on build/ path.


