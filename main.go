package main

import (
	"fethcher/cmd"
	"fmt"
	"os"
)

func main() {
	if err := cmd.Start(); err != nil {
		fmt.Printf("server run into an error: %s", err)
		os.Exit(1)
	}
}
