package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/duongvanha/fanaticsCrawler/logger"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/joho/godotenv/autoload"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var dbConnection *gorm.DB
var l = sync.RWMutex{}

func init() {
	db, err := gorm.Open("postgres", os.Getenv("DATABASE_URL"))

	if err != nil {
		fmt.Printf("failed to connect database : %v", err)
		panic("failed to connect database")
	}

	dbConnection = db

	db.AutoMigrate(&Size{}, &Product{})
}

type Model struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Breadcrumb struct {
	Model
	Level int
	Text  string `gorm:"PRIMARY_KEY"`
}
type Size struct {
	Model
	Text string `gorm:"PRIMARY_KEY"`
}

type Product struct {
	Model
	Id          int `gorm:"PRIMARY_KEY"`
	Breadcrumbs string
	Sizes       []Size `gorm:"many2many:product_size;"`
	Detail      string
	Description string
}

func getFailRetry(url string, numberRetry int, errBefore error) (doc *goquery.Document, err error) {
	timeStart := time.Now()

	defer func() {
		logger.BkLog.Infof("Get url %v took %v", url, time.Since(timeStart))
	}()

	if numberRetry <= -1 {
		return nil, errBefore
	}
	res, err := http.Get(url)

	if err != nil {
		logger.BkLog.Warn("Error get url : %v with error : %v", url, err)
		return getFailRetry(url, numberRetry-1, err)
	}
	defer res.Body.Close()

	doc, err = goquery.NewDocumentFromReader(res.Body)

	if err != nil {
		logger.BkLog.Warn("Error read document in url : %s", url)
		return getFailRetry(url, numberRetry-1, err)
	}

	if res.StatusCode != 200 {
		logger.BkLog.Warn("status code error: %d %s %s", res.StatusCode, res.Status, url)
		return getFailRetry(url, numberRetry-1, err)
	}

	return doc, err
}

func containsString(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

func getUrlPage(url string) string {
	return "https://www.fanatics.com" + url
}

func CrawlerPageProduct(url string, isFree bool) {
	document, err := getFailRetry(url, 3, nil)

	if err != nil {
		logger.BkLog.Errorf("Can not get page url: %v, err %v", err)
		return
	}

	var breadcrumbs []string
	var sizes []Size

	document.Find(`.breadcrumbs-container li[typeof="ListItem"]`).Each(func(i int, selection *goquery.Selection) {
		breadcrumbs = append(breadcrumbs, selection.Text())
	})

	regex, _ := regexp.Compile("Product ID: (\\d+)")

	res := regex.FindStringSubmatch(breadcrumbs[len(breadcrumbs)-1])

	document.Find(".size-selector-list .size-selector-button").Each(func(i int, selection *goquery.Selection) {
		sizes = append(sizes, Size{Text: selection.Text()})
	})

	detail := document.Find(".product-details").Text()

	description := document.Find(".description-box-content div").Text()

	breadcrumbsJson, _ := json.Marshal(breadcrumbs)

	product := &Product{
		Breadcrumbs: string(breadcrumbsJson),
		Sizes:       sizes,
		Description: description,
		Detail:      detail,
	}

	if len(res) > 0 {
		i64, _ := strconv.ParseInt(res[0], 10, 32)
		product.Id = int(i64)
	}
	l.Lock()
	if product.Id > 0 {
		dbConnection.FirstOrCreate(&product, &Product{Id: product.Id})
	} else {
		dbConnection.Create(&product)
	}
	l.Unlock()

}

func CrawlerPageCollection(maxRoutine int, url string) {
	document, err := getFailRetry(url, 3, nil)

	if err != nil {
		logger.BkLog.Errorf("Can not get page url: %v, err %v", err)
		return
	}

	href, exist := document.Find(`div.side-nav-facet-items.featuredDepartmentsBoxes a[href*="jerseys"]`).Attr("href")

	if !exist {
		logger.BkLog.Errorf("href jerseys nil %v", url)
		return
	}

	document, err = getFailRetry(getUrlPage(href), 3, nil)

	if err != nil {
		logger.BkLog.Errorf("Can not get page url: %v, err %v", err)
		return
	}

	type dataProduct struct {
		url         string
		existJersey bool
	}

	var urls []dataProduct

	document.Find("div.product-card").Each(func(i int, selection *goquery.Selection) {
		href, exist := selection.Find("div.product-image-container > a").Attr("href")
		_, existJersey := selection.Find("span.jersey-assurance-message").Attr("href")

		if !exist {
			logger.BkLog.Errorf("warning href is index %v, page %v", i, getUrlPage(href))
			return
		}

		urls = append(urls, dataProduct{
			url:         getUrlPage(href),
			existJersey: existJersey,
		})

	})

	maxItem := len(urls)

	var ch = make(chan dataProduct, maxItem)
	var wg sync.WaitGroup

	wg.Add(maxItem)
	for i := 0; i < maxRoutine; i++ {
		go func() {
			for {
				data, ok := <-ch
				if !ok {
					wg.Done()
					return
				}
				go CrawlerPageProduct(data.url, data.existJersey)
			}
		}()
	}

	for _, url := range urls {
		ch <- url
	}

	close(ch)
	wg.Wait()

}

func startRunCrawlerPage(maxRoutine int) {

	document, err := getFailRetry("https://www.fanatics.com/", 3, nil)

	if err != nil {
		logger.BkLog.Errorf("Can not get page %v", err)
	}

	menuUsed := []string{"nfl", "mlb", "nba", "nhl"}

	urls := []string{}

	document.Find("body header .top-nav-component li").Each(func(i int, s *goquery.Selection) {
		a := s.Find("a.top-nav-item-link").Text()
		if !containsString(menuUsed, a) {
			return
		}

		s.Find(".dropdown-column > a").Each(func(i int, selection *goquery.Selection) {

			href, exits := selection.Attr("href")

			if !exits {
				return
			}

			urls = append(urls, getUrlPage(href))
		})
	})

	maxItem := len(urls)

	var ch = make(chan string, maxItem)
	var wg sync.WaitGroup

	wg.Add(maxItem)
	for i := 0; i < maxRoutine; i++ {
		go func() {
			for {
				url, ok := <-ch
				if !ok {
					wg.Done()
					return
				}
				go CrawlerPageCollection(10, url)
			}
		}()
	}

	for _, url := range urls {
		ch <- url
	}

	close(ch)
	wg.Wait()

}

func main() {
	http.HandleFunc("/crawler", func(w http.ResponseWriter, r *http.Request) {
		startRunCrawlerPage(10)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "hello word")
	})

	_ = http.ListenAndServe(":"+os.Getenv("PORT"), nil)

}
