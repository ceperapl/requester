package main

import (
	"log"

	"github.com/ceperapl/requester/cmd"
)

func main() {
	if err := cmd.RunServer(); err != nil {
		log.Fatal(err)
	}
}
