package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

var ErrNotDownloaded = errors.New("Item not downloaded.")

type ItemDesc struct {
	url       url.URL
	dateAdded time.Time
}

func (i *ItemDesc) cacheFile() string {
	return filepath.Join(cacheDir, i.filename())
}

func (i *ItemDesc) filename() string {
	return i.url.Path
}

func (i *ItemDesc) cached() (Item, error) {
	if !i.downloaded() {
		return Item{}, ErrNotDownloaded
	}

	fd, err := os.Open(i.cacheFile())
	if err != nil {
		return Item{}, err
	}

	raw, err := io.ReadAll(fd)
	if err != nil {
		return Item{}, err
	}

	var item Item
	err = json.Unmarshal(raw, &item)
	return item, err
}

func (i *ItemDesc) downloaded() bool {
	_, err := os.Stat(i.cacheFile())
	return err != nil
}

func (i *ItemDesc) toItem() (Item, error) {
	if i.downloaded() {
		item, err := i.cached()
		if err == nil {
			return item, err
		}
	}

	// Download the content
	client := &http.Client{}

	resp, err := client.Get(i.url.String())
	if err != nil {
		return Item{}, err
	}

	defer resp.Body.Close()
	if value, ok := resp.Header["Content-Type"]; ok {
		if len(value) < 1 && value[0] != "text/html" {
			return Item{}, fmt.Errorf("Can't parse %s", value)
		}
	}
	raw, err := io.ReadAll(resp.Body)

	// Parse the content using readability.js
	content, err := parseContent(i.url, raw)
	if err != nil {
		return Item{}, err
	}

	fmt.Println("TODO: cache content")

	return Item{
		title:    content.Title,
		content:  content.Text,
		ItemDesc: *i,
	}, err
}

type Item struct {
	title   string
	content string
	ItemDesc
}

type ItemHeap []Item

func (h ItemHeap) Len() int               { return len(h) }
func (h ItemHeap) Less(i int, j int) bool { return h[i].dateAdded.After(h[j].dateAdded) }
func (h ItemHeap) Swap(i, j int)          { h[i], h[j] = h[j], h[i] }

func (h *ItemHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

func (h *ItemHeap) Push(x interface{}) {
	item := x.(Item)
	*h = append(*h, item)
}
