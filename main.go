package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/quocquann/download_image_book/types"
)

func crawlImageUrl() (chan types.Job, error) {
	res, err := http.Get("https://gacxepbookstore.vn/all-books?page=5")
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Status code error")
	}

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)

	if err != nil {
		return nil, err
	}

	bookItems := doc.Find(".product-loop-1.product-base")
	numBookItem := bookItems.Length()
	jobs := make(chan types.Job, numBookItem)
	bookItems.Each(func(i int, s *goquery.Selection) {
		url := s.Find(".product-thumbnail>a.image_link.display_flex>img").AttrOr("data-lazyload", "")
		fileName := s.Find("h3.product-name a").AttrOr("href", "")
		jobs <- types.Job{Url: "https:" + url, FileName: fileName + ".jpg"}
	})

	close(jobs)

	return jobs, nil
}

func downLoad(jobs chan types.Job, wg *sync.WaitGroup) {
	for job := range jobs {
		res, err := http.Get(job.Url)
		if err != nil {
			return
		}

		if res.StatusCode != 200 {
			return
		}

		defer res.Body.Close()

		file, err := os.Create("./images" + job.FileName)
		if err != nil {
			return
		}
		_, err = io.Copy(file, res.Body)
		if err != nil {
			return
		}
	}

	wg.Done()
}

func main() {
	jobs, err := crawlImageUrl()
	if err != nil {
		log.Fatal(err)
	}

	const numWorker = 20
	wg := sync.WaitGroup{}
	for i := 0; i < numWorker; i++ {
		wg.Add(1)
		go downLoad(jobs, &wg)
	}

	wg.Wait()
}
