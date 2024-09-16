package main

import (
	"log"

	"github.com/JackalLabs/mulberry/cmd"
)

func main() {
	root := cmd.RootCMD()

	err := root.Execute()
	if err != nil {
		log.Fatal(err)
		return
	}
}
