package main

import (
	"os"

	"github.com/greycubesgav/integrity"
)

func main() {
	status := integrity.Run()
	os.Exit(status)
}
