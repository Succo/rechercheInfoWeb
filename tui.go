package main

import (
	"fmt"

	"github.com/peterh/liner"
)

func dynamicSearch(cacm, cs276 *Search) {
	// library used for the search prompt
	line := liner.NewLiner()
	defer line.Close()

	fmt.Println("Ctrl-C to exit the program")
	line.SetCtrlCAborts(true)
	corpus, err := line.Prompt("What corpus do you want to search? 'cacm'[1] or 'cs276'[2]? ")
	if err == liner.ErrPromptAborted {
		fmt.Println("Aborting prompt")
		return
	} else if err != nil {
		return
	}
	var search *Search
	if corpus == "cacm" || corpus == "1" {
		search = cacm
	} else if corpus == "cs276" || corpus == "2" {
		search = cs276
	} else {
		fmt.Println("Error unsuported option")
		return
	}
	for {
		if searched, err := line.Prompt("Searched keyword? "); err == nil {
			if searched == "q" || searched == ":q" {
				return
			}
			results := search.Search(searched)
			fmt.Printf("Found %d result\n", len(results))
			for _, result := range results {
				fmt.Println(result)
			}
		} else if err == liner.ErrPromptAborted {
			return
		}
	}
}
