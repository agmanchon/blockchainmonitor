package main

import (
	"github.com/agmanchon/blockchainmonitor/cmd"
	"log"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Printf("Error executing cmd %v", err)
		os.Exit(1)
	}
}
