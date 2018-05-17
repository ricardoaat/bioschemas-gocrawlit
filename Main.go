package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/ricardoaat/bioschemas-gocrawlit/crawler"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

var (
	version   string
	buildDate string
)

func logInit(d bool) {

	logfile := "biocrawlit.log"
	fmt.Println("Loging to " + logfile)
	log.SetOutput(os.Stdout)
	if d {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.SetFormatter(&log.TextFormatter{})
	pathMap := lfshook.PathMap{
		log.DebugLevel: logfile,
		log.InfoLevel:  logfile,
		log.ErrorLevel: logfile,
		log.WarnLevel:  logfile,
		log.PanicLevel: logfile,
	}
	log.AddHook(lfshook.NewHook(
		pathMap,
		&log.JSONFormatter{},
	))
}

func main() {

	e := flag.Bool("e", false, "Connects to an elastisearch server on http://127.0.0.1:9200")
	d := flag.Bool("d", false, "Sets up the log level to debug")
	v := flag.Bool("v", false, "Returns the binary version and built date info")
	q := flag.Bool("q", false, "Skip queries on the URL.")
	u := flag.String("u", "", "Url to crawl and extract markup")
	m := flag.Int("m", 0, "Max number of recursion depth of visited URLs")
	p := flag.Bool("p", false, "Stay on current path.")
	qr := flag.String("query", "", "Pagination query word")

	flag.Parse()

	logInit(*d)

	log.Info("--------------Init program--------------")
	log.Info(fmt.Sprintf("Version: %s Build Date: %s", version, buildDate))

	if !*v {

		log.Info("URL to crawl ", *u)

		if *u == "" {
			log.Error("Empty URL")
		}

		baseURL, err := url.Parse(*u)
		if err != nil {
			log.Error("Error parsing URL ", err)
		}

		nq := baseURL
		nq.RawQuery = ""

		f := fmt.Sprintf("%s%sschema.json", baseURL.Host, strings.Replace(baseURL.Path, "/", "_", -1))

		filter := ""
		if *p {
			filter = fmt.Sprintf(`^%s://%s%s`, baseURL.Scheme, baseURL.Host, baseURL.Path)
		}

		var ad []string
		ad = append(ad, baseURL.Host)
		ad = append(ad, fmt.Sprintf("www.%s", baseURL.Host))

		c := crawler.Crawler{
			UseElastic:     *e,
			Index:          baseURL.Host,
			OutputFileName: f,
			BaseURL:        baseURL,
			SkipQueries:    *q,
			MaxDepth:       *m,
			AllowedDomains: ad,
			Filter:         filter,
			QueryWord:      *qr,
		}

		c.Init()

		if *e {
			if err := c.ElasticInit(); err != nil {
				log.Error("Error initializing elastic function ")
			}
		}

		c.Start()

		// err = c.ToJSONfile()
		// if err != nil {
		// 	log.Error("ToJSONfile error ", err)
		// }

	}
}
