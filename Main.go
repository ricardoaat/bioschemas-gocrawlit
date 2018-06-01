package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/runner"
	"github.com/ricardoaat/bioschemas-gocrawlit/crawler"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

var (
	version   string
	buildDate string
	c         crawler.Crawler
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
	p := flag.Bool("p", false, "Stay on current path.")
	m := flag.Int("m", 0, "Max number of recursion depth of visited URLs")
	u := flag.String("u", "", "Url to crawl and extract markup")
	qr := flag.String("query", "", "Pagination query word")

	flag.Parse()

	logInit(*d)

	log.Info("--------------Init program--------------")
	log.Info(fmt.Sprintf("Version: %s Build Date: %s", version, buildDate))

	testChromed()
	return

	if *v {
		return
	}

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

	c = crawler.Crawler{
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

}

func testChromed() {
	var err error

	// create context
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create chrome instance
	c, err := chromedp.New(ctxt, chromedp.WithLog(log.Printf), chromedp.WithRunnerOptions(
		runner.Flag("headless", true),
		//runner.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.181 Safari/537.36"),
		runner.Flag("disable-gpu", true),
		runner.Flag("no-first-run", true),
		runner.Flag("no-default-browser-check", true),
	))

	if err != nil {
		log.Fatal(err)
	}

	// run task list
	var res string
	url := `http://bioschemas.org/bioschemas-uniprot-render/`
	err = c.Run(ctxt, text(&res, url))
	if err != nil {
		log.Fatal(err)
	}

	// shutdown chrome
	log.Warn("Shutting down CHROME")
	err = c.Shutdown(ctxt)
	if err != nil {
		log.Fatal(err)
	}

	log.Warn("Result ", res)
}

func text(res *string, url string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Sleep(1 * time.Second),
		//script[type=\"application/ld+json\"]
		chromedp.WaitReady(`data-loader`, chromedp.ByQuery),
		chromedp.InnerHTML(`html`, res, chromedp.ByQuery),
	}
}
