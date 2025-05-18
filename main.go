package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

type Page struct {
	Title string `xml:"title"`
	Text  string `xml:"revision>text"`
}

type Result struct {
	Title string
	Links []string
}

func main() {
	file, err := os.Open("wiki.xml")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	outFile, err := os.Create("wiki_graph.ndjson")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()
	writer := bufio.NewWriter(outFile)
	defer writer.Flush()

	decoder := xml.NewDecoder(file)
	linkRegex := regexp.MustCompile(`\[\[([^|\]\[]+)(?:\|.*?)?\]\]`)

	pageChan := make(chan Page, 1000)
	resultChan := make(chan Result, 1000)

	// Writer goroutine, adds the page to links map to the output file
	var writerWg sync.WaitGroup
	writerWg.Add(1)
	go func() {
		defer writerWg.Done()
		enc := json.NewEncoder(writer)
		count := 0
		for result := range resultChan {
			if len(result.Links) == 0 {
				continue
			}
			record := map[string][]string{result.Title: result.Links}
			if err := enc.Encode(record); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", result.Title, err)
			}
			count++
			if count%100000 == 0 {
				fmt.Printf("Wrote %d pages...\n", count)
			}
		}
	}()

	// Worker goroutines to extract links, uses all your threads
	numWorkers := runtime.NumCPU()
	var workerWg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for page := range pageChan {
				if page.Title == "" || strings.HasPrefix(page.Title, "File:") || strings.HasPrefix(page.Title, "Category:") {
					continue
				}
				matches := linkRegex.FindAllStringSubmatch(page.Text, -1)
				links := make([]string, 0, len(matches))
				for _, match := range matches {
					linked := strings.TrimSpace(match[1])
					if linked != "" && !strings.Contains(linked, ":") {
						links = append(links, linked)
					}
				}
				resultChan <- Result{Title: page.Title, Links: links}
			}
		}()
	}

	// XML parsing: read <page> elements
	go func() {
		for {
			token, err := decoder.Token()
			if err == io.EOF {
				break
			} else if err != nil {
				panic(err)
			}

			switch se := token.(type) {
			case xml.StartElement:
				if se.Name.Local == "page" {
					var page Page
					err := decoder.DecodeElement(&page, &se)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Failed to decode page: %v\n", err)
						return
					}
					pageChan <- page
				}
			}
		}
		close(pageChan)
	}()

	workerWg.Wait()
	close(resultChan)
	writerWg.Wait()

	fmt.Println("Finished writing wiki_graph.ndjson")
}
