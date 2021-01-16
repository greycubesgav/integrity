package main

import (
	"fmt"
	"github.com/corona10/goimagehash"
	"image/jpeg"
	"os"
)

func integrityPhashFromFile(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return "", err
	}

	fileSize := int64(fi.Size())

	if fileSize == 0 {
		return "", nil
	}

	img1, _ := jpeg.Decode(f)
	pHash, _ := goimagehash.PerceptionHash(img1)
	return fmt.Sprintf("%016x", pHash.GetHash()), nil
}
