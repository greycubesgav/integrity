package integrity

import (
	"fmt"
	"image/jpeg"
	"os"

	"github.com/corona10/goimagehash"
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
		return "", fmt.Errorf("filesize is zero")
	}

	img1, err := jpeg.Decode(f)
	if err != nil {
		return "", err
	}
	pHash, err := goimagehash.PerceptionHash(img1)
	if err != nil {
		return "", err
	} else {
		return fmt.Sprintf("%016x", pHash.GetHash()), nil
	}

}
