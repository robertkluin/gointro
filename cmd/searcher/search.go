package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type State struct {
	Source string
	Match  bool
	Count  int
}

func search(word, source string, out chan<- State) {
	response, err := http.Get(source)
	if err != nil {
		log.Printf("Failed to get source, %s, with error: %s", source, err)
		return
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Failed to load source, %s, with error: %s", source, err)
		return
	}

	lower_content := strings.ToLower(string(contents))
	match_count := strings.Count(lower_content, word)

	out <- State{source, true, match_count}
}

func do_sync_search(word string) {
	sources := []string{
		"https://news.ycombinator.com/",
		"http://slashdot.org/",
		"http://www.reddit.com/",
	}

	result_dump := make(chan State, 1)

	for _, source := range sources {
		search(word, source, result_dump)
		result := <-result_dump
		if result.Match {
			log.Printf("Term found %d times in %s", result.Count, result.Source)
		} else {
			log.Printf("Term NOT found in: %s", result)
		}
	}
}

func do_concurrent_search(word string) {
	sources := []string{
		"https://news.ycombinator.com/",
		"http://slashdot.org/",
		"http://www.reddit.com/",
	}

	result_dump := make(chan State, 1)

	for _, source := range sources {
		go search(word, source, result_dump)
	}

	result_count := 0
	for {
		select {
		case result := <-result_dump:
			result_count++
			if result.Match {
				log.Printf("Term found %d times in %s", result.Count, result.Source)
			} else {
				log.Printf("Term NOT found in: %s", result)
			}

			if result_count == len(sources) {
				return
			}
		}
	}
}

func main() {
	sync_mode := flag.Bool("sync", false, "Run in sync mode.")

	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatal("usage: search [sync] word_to_find")
		return
	}

	word := flag.Arg(0)

	if *sync_mode {
		do_sync_search(word)
	} else {
		do_concurrent_search(word)
	}
}
