package main

import (
	"fmt"

	"github.com/alexbakker/blogen/commands"
)

func main() {
	if err := commands.RootCmd.Execute(); err != nil {
		fmt.Printf("error: %s\n", err.Error())
	}
}
