package testpkg

import (
	"log"
	"os"
)

func badPanic() {
	panic("test panic") // want "avoid using panic"
}

func badLogFatal() {
	log.Fatal("test fatal") // want "log.Fatal should not be used outside main.main"
}

func badLogFatalf() {
	log.Fatalf("test fatalf: %v", 123) // want "log.Fatalf should not be used outside main.main"
}

func badOsExit() {
	os.Exit(1) // want "os.Exit should not be used outside main.main"
}

func goodFunction() {
	println("This is fine")
}
