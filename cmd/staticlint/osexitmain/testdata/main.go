package main

import "os"

func main() {
	os.Exit(1) // want "found os.Exit in main function"
}
