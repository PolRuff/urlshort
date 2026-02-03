package notmain

import (
	"log"
	"os"
)

func helper() {
	panic("forbidden")     // want "use of built-in function panic is forbidden"
	log.Fatal("forbidden") // want "call to log.Fatal is forbidden outside main.main"
	os.Exit(1)             // want "call to os.Exit is forbidden outside main.main"
}
