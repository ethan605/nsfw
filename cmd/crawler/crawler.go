package main

import (
	"fmt"
	"nsfw/internal/crawler"
	"os"
)

func main() {
	fmt.Println("Start crawling...")

	source, _ := os.Open("./configs/crawler/seeds/instagram.json")
	crawler.CrawlInstagram(source)
}
