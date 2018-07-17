package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func ExampleScrape() {
	// Request the HTML page.
	res, err := http.Get("http://studygolang.com/topics")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	nodes := doc.Find(".topics .topic")

	for i := nodes.Length() - 1; i >= 0; i-- {
		node := nodes.Get(i)
		title := goquery.NewDocumentFromNode(node).Find(".title a").Text()
		fmt.Printf("Topic %d: %s\n", i+1, title)
	}
}

func main() {
	ExampleScrape()
}
