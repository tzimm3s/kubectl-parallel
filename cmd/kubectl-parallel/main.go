package main

import (
	"fmt"
	"os"
)

func main() {
	rootCmd := NewRootCommand()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
