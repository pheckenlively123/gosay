package main

import (
	"errors"
	"flag"
)

type cmdopt struct {
	message string
	animal  string
	list    bool
}

func main() {

	opt, err := getopt()
	if err != nil {
		panic("error processing command line arguments: " + err.Error())
	}

}

func getopt() (cmdopt, error) {
	message := flag.String("message", "", "Message for the gopher to say.")
	animal := flag.String("animal", "", "Alternate animal.")
	list := flag.Bool("list", false, "List available animals.")

	flag.Parse()

	if *message == "" && !*list {
		return cmdopt{}, errors.New("-message is a required parameter")
	}

	return cmdopt{
		message: *message,
		animal:  *animal,
		list:    *list,
	}, nil
}
