package main

import (
	"log"
	"os"
)

func helper() {
	// These are OK because they are not exit calls
	log.Print("info")
	os.Getpid()
}

func main() {
	// All exit calls are allowed in main.main
	log.Fatal("ok")
	os.Exit(0)
}
