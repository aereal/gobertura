package main

import (
	"os"

	"github.com/aereal/gobertura/internal/cli"
)

func main() {
	c := cli.NewStd()
	os.Exit(c.Run(os.Args[1:]))
}
