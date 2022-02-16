package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
)

type Content struct {
	Byline  string `json:"byline"`
	Dir     string `json:"dir"`
	Excerpt string `json:"excerpt"`
	Lang    string `json:"lang"`
	Length  int    `json:"length"`
	Site    string `json:"site"`
	Text    string `json:"text"`
	Title   string `json:"title"`
	Error   string `json:"error"`
}

func parseContent(url url.URL, html []byte) (Content, error) {
	cmd := exec.Command("./make-readable", url.String())
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return Content{}, err
	}

	stdin.Write(html)
	stdin.Close()

	raw, err := cmd.Output()
	if err != nil {
		return Content{}, err
	}

	var content Content
	err = json.Unmarshal(raw, &content)
	if err != nil {

		return content, err
	}

	if content.Error != "" {
		return content, fmt.Errorf("Error: %s", content.Error)
	}

	fmt.Println("TODO: clean content by normalizing whitespace")
	return content, nil
}
