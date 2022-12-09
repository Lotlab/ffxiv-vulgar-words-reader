package main

import (
	"fmt"
	"io"
	"os"

	"github.com/lotlab/ffxiv-vulgar-words-reader/pkg"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("%s: <dict>\n", os.Args[0])
		return
	}

	file := os.Args[1]
	fs, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	bytes, err := io.ReadAll(fs)
	if err != nil {
		panic(err)
	}

	dict, err := pkg.NewDict(bytes)
	if err != nil {
		panic(err)
	}

	strs, err := dict.DumpDict()
	if err != nil {
		panic(err)
	}
	for _, v := range strs {
		fmt.Println(v)
	}
}
