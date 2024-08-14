package main

import (
	"os"

	"github.com/greycubesgav/integrity/pkg/integrity"
)

func main() {
	status := integrity.Run()
	os.Exit(status)
}
