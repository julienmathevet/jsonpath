package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/gdey/jsonpath"
)

func applyFilter(filter jsonpath.Applicator, data []byte) (interface{}, error) {

	var jsonbody = make(map[string]interface{})
	err := json.Unmarshal(data, &jsonbody)
	if err != nil {
		return nil, err
	}
	return filter.Apply(jsonbody)
}

// ./jsonpath "$.store" a.json
// cat a.json | ./jsonpath "$.store"
func main() {
	if len(os.Args) < 2 {
		// Only have the command name. Need to print out help text.
		log.Printf("Usage: %v '$['json']['path']['string']' a.json...\n", os.Args[0])
		os.Exit(1)
	}
	filter, err := jsonpath.Parse(os.Args[1])
	if err != nil {
		log.Printf("Usage: %v '$['json']['path']['string']' a.json...\n", os.Args[0])
		log.Printf("Got the following error Parsing %v : %v\n", os.Args[1], err)
		os.Exit(1)
	}
	if len(os.Args) == 2 {
		// use stdinput
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Printf("Usage: %v '$['json']['path']['string']' a.json...\n", os.Args[0])
			log.Printf("Got the following error Reading from STDIN  %v\n", err)
			os.Exit(1)
		}
		d, err := applyFilter(filter, data)
		if err != nil {
			log.Printf("Usage: %v '$['json']['path']['string']' a.json...\n", os.Args[0])
			log.Printf("Got the following error Applying filter  %v\n", err)
			os.Exit(1)
		}
		data, err = json.MarshalIndent(d, "", "   ")
		if err != nil {
			log.Printf("Usage: %v '$['json']['path']['string']' a.json...\n", os.Args[0])
			log.Printf("Got the following error Printing out json %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	}
	for _, filename := range os.Args[2:] {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Printf("Got the following error Reading from %v  %v\n", filename, err)
			continue
		}
		d, err := applyFilter(filter, data)
		if err != nil {
			log.Printf("Got the following error Applying filter for %v  %v\n", filename, err)
			continue
		}
		data, err = json.MarshalIndent(d, "", "   ")
		if err != nil {
			log.Printf("Got the following error Trying to print result for file %v  %v\n", filename, err)
			continue
		}
		fmt.Println(string(data))
	}

}
