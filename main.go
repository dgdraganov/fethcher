package main

import (
	"fethcher/cmd"
	"os"
)

func main() {
	if err := cmd.Start(); err != nil {
		os.Exit(1)
	}
}
