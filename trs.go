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
	word = flag.String("w", "", "word will translating")
)

func usage() {
	fmt.Fprintf(os.Stderr, "trs is translate command line program\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "trs [option]\n")
	flag.PrintDefaults()
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

	ydCli := NewYouDaoAPIClient(APIURL)
	resp, err := ydCli.Translate(*word)
	if err != nil {
		exitOnErr(err)
	}

	fmt.Println(resp)
}
