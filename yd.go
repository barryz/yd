package main

import (
	"flag"
	"fmt"
	"os"
)

func init() {
	flag.Usage = usage
	flag.Parse()
}

var (
	word = flag.String("w", "", "the word will translating")
	anki = flag.Bool("anki", false, "whether import result to anki")
)

func usage() {
	fmt.Fprintf(os.Stderr, "yd is translate command line program\n")
	fmt.Fprintf(os.Stderr, "Usage: yd [options]\n")
	fmt.Fprintf(os.Stderr, "Options: \n")
	fmt.Fprintf(os.Stderr, "-w       the word will translating\n")
	fmt.Fprintf(os.Stderr, "-anki    whether import result to anki\n")
	os.Exit(0)
}

func exitOnErr(err error) {
	fmt.Println(err)
	os.Exit(2)
}

func main() {
	if *word == "" {
		exitOnErr(fmt.Errorf("you must speicify a word to translate"))
	}

	ydCli := NewYouDaoAPIClient(YouDaoAPI)
	resp, err := ydCli.Translate(*word)
	if err != nil {
		exitOnErr(err)
	}

	// print result first
	fmt.Println(resp)

	if *anki {
		deck, err := GetDeck()
		if err != nil {
			exitOnErr(err)
		}
		// anki note
		note := &AnkiNoteMeta{
			Deck:     deck,
			Front:    resp.Word(),
			Back:     resp.AnkiBackContent(),
			AudioURL: resp.USSpeechLink(),
		}

		err = NewAnkiClient(AnkiConnectAPI).AddNote(note)
		if err != nil {
			exitOnErr(err)
		}
	}
}
