package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"treaster/applications/poe_filter/lib"
)

func main() {
	inputFile := flag.String("input", "", "")
	outputFile := flag.String("output", "", "")
	flag.Parse()

	inputBytes, err := ioutil.ReadFile(*inputFile)
	if err != nil {
		fmt.Println("error reading file: ", err.Error())
		os.Exit(1)
	}

	output, err := lib.Compile(string(inputBytes))
	if err != nil {
		fmt.Println("error in compile: ", err.Error())
		os.Exit(1)
	}

	if *outputFile == "" {
		fmt.Println(output)
		return
	}

	err = ioutil.WriteFile(*outputFile, []byte(output), 0644)
	if err != nil {
		fmt.Println("error in writing file: ", err.Error())
		os.Exit(1)
	}

	return
}
