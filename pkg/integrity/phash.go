package integrity

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"github.com/corona10/goimagehash"
)

func integrityPhashFromFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	// Add a function to defer to ensure any issue closing the file is reported
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing file: %s\n", err)
		}
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}
	fileSize := int64(fileInfo.Size())
	if fileSize == 0 {
		return "", fmt.Errorf("filesize is zero")
	}

	// Limit the reader to only the necessary bytes
	limitedReader := io.LimitReader(file, 512) // 512 bytes should be enough for most formats

	// Use image.DecodeConfig to only decode the configuration (which includes format)
	_, _, err = image.DecodeConfig(limitedReader)
	if err != nil {
		return "", err
	}

	// Check the image type
	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		return "", err
	}

	// Reset the file pointer
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	// Decode the entire image
	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}

	pHash, err := goimagehash.PerceptionHash(img)
	if err != nil {
		return "", err
	} else {
		return fmt.Sprintf("%016x", pHash.GetHash()), nil
	}

}
