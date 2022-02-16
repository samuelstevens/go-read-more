package main

import (
	"container/heap"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const cacheDir = "/Users/samstevens/iCloud/reading-list-cache"
const feedFile = "/Users/samstevens/Development/personal-website/readinglist.xml"

// const LinkFile = "/Users/samstevens/iCloud/reading-list-links.txt"
const LinkFile = "reading-list-links.txt"
const SocketLimit = 16

type Html struct {
	title string
}

func getHtml(url url.URL) (Html, error) {
	panic("not implemented")

	client := &http.Client{}

	resp, err := client.Get("http://example.com")
	if err != nil {
		return Html{}, err
	}

	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)

	// TODO: Might have to do some work on decoding based on the content type.
	return Html{}, err
}

func parseArticleToRead(link string, secs int64) (ItemDesc, error) {
	if secs > 32503680000 {
		// If the date is newer than year 3000, it's probably milliseconds.
		return parseArticleToRead(link, secs/1000)
	}

	url, err := url.Parse(link)
	if err != nil {
		return ItemDesc{}, err
	}

	date := time.Unix(secs, 0)

	return ItemDesc{*url, date}, nil
}

func writeItem(item Item) error {
	panic("not implemented")
}

type ParseResult struct {
	desc ItemDesc
	err  error
}

func parse(path string) <-chan ParseResult {
	outc := make(chan ParseResult)

	go func() {
		// Read the lines from the link file
		contents, err := os.ReadFile(path)
		if err != nil {
			outc <- ParseResult{ItemDesc{}, err}
			close(outc)
			return
		}

		// Get all the non-empty lines
		for _, line := range strings.Split(string(contents), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				log.Printf("Skipping line because it is empty.")
				continue
			}

			// Split on the comma
			split := strings.Split(line, ",")
			if len(split) != 2 {
				outc <- ParseResult{ItemDesc{}, fmt.Errorf("Line '%s' has more %d elements.", line, len(split))}
				continue
			}

			// Parse the two strings on each line
			link := split[0]
			time, err := strconv.ParseInt(split[1], 10, 64)
			if err != nil {
				outc <- ParseResult{ItemDesc{}, err}
				continue
			}

			item, err := parseArticleToRead(link, time)
			if err != nil {
				outc <- ParseResult{ItemDesc{}, err}
				continue
			}

			outc <- ParseResult{item, nil}
		}
		close(outc)
	}()

	return outc
}

type DownloadResult struct {
	item Item
	err  error
}

func downloader(in <-chan ParseResult, out chan<- DownloadResult, errHandler func(error)) {
	for p := range in {
		if p.err != nil {
			errHandler(p.err)
			continue
		}

		item, err := p.desc.toItem()
		out <- DownloadResult{item, err}
	}
}

func download(in <-chan ParseResult, errHandler func(error)) <-chan DownloadResult {
	outc := make(chan DownloadResult)
	var wg sync.WaitGroup

	wg.Add(SocketLimit)
	for i := 0; i < SocketLimit; i++ {
		go func() {
			downloader(in, outc, errHandler)
			wg.Done()
		}()
	}

	// Wait for all go routines to finish downloading
	// before closing the out channel.
	go func() {
		wg.Wait()
		close(outc)
	}()

	return outc
}

func sorted(in <-chan DownloadResult) []Item {
	h := ItemHeap{}
	heap.Init(&h)

	for d := range in {
		if d.err != nil {
			log.Printf(d.err.Error())
			continue
		}

		heap.Push(&h, d.item)
	}

	return h
}

func logErr(err error) {
	log.Printf(err.Error())
}

func main() {
	// Parse the link file into an iterator of item descriptions.
	parsedc := parse(LinkFile)

	// As they arrive, get the content (from the network)
	itemc := download(parsedc, logErr)
	log.Println("Downloading...")

	// Insert them into a thread-safe sorted list using a heap.
	items := sorted(itemc)
	log.Println("Sorting...")

	fmt.Println(items)

	fmt.Println("Convert these items into an xml expression.")
}
