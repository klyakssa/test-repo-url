package main

import (
	"log"
	"os"
)

func main() {
	log.Fatal("allowed in main")
	log.Fatalf("allowed in main")
	os.Exit(0)

	panic("not allowed even in main") // want "avoid using panic"
}

func helper() {
	log.Fatal("not allowed") // want "log.Fatal should not be used outside main.main"
	os.Exit(1)               // want "os.Exit should not be used outside main.main"
}
