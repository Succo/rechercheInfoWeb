package main

import (
	"fmt"

	"github.com/peterh/liner"
)

func dynamicSearch(cacm, cs276 *Search) {
	// library used for the search prompt
	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)
	corpus, err := line.Prompt("What corpus do you want to search? 'cacm' or 'cs276'")
	if err == liner.ErrPromptAborted {
		fmt.Println("Aborting prompt")
		return
	} else if err != nil {
		return
	}
	if corpus == "cacm" {
		fmt.Println("Working with cacm")
	} else if corpus == "cs276" {
		fmt.Println("Working with cs276")
	} else {
		fmt.Println("Error unsuported option")
	}
}
