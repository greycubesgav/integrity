package main

import (
	"fmt"
	"flag"
	"os"
	"path/filepath"
	"log"
	"github.com/pkg/xattr"
	"crypto/sha1"
	"io"
	"encoding/hex"
	"strings"
)


type integrity_fileCard struct {
	FileInfo 		*os.FileInfo
	fullpath		string
	checksum		string
	digest_type		string
}

var config *Config = NewConfig()

const prefix = "user."


func integ_getChecksumRaw (path string) (string, error) {
	var err error
	var data []byte
	if data, err = xattr.Get(path, "integ.sha1"); err != nil {
		return "", err
	}
	return string(data[:len(data)]), nil
}

func integ_getChecksum (currentFile *integrity_fileCard) (error) {
	var err error
	if currentFile.checksum, err = integ_getChecksumRaw(currentFile.fullpath); err != nil {
		return err
	}
	return nil
}

func integ_removeChecksum (currentFile *integrity_fileCard) (error) {
	var err error
	if err = xattr.Remove(currentFile.fullpath, "integ.sha1"); err != nil {
		return err
	}
	return nil
}

func integ_generateChecksum (currentFile *integrity_fileCard) (error) {
	var err error

	fileHandle, err := os.Open(currentFile.fullpath)
	if err != nil {
		return err
	}
	defer fileHandle.Close()

	hashGen := sha1.New()
	if _, err := io.Copy(hashGen, fileHandle); err != nil {
		return err
	}
	currentFile.checksum = hex.EncodeToString(hashGen.Sum(nil))
	return nil
}

func integ_writeChecksum (currentFile *integrity_fileCard) (error) {
	var err error
	if err = integ_generateChecksum(currentFile); err != nil {
		return err
	}
	checksumBytes := []byte(currentFile.checksum)
	if err = xattr.Set(currentFile.fullpath, "integ.sha1", checksumBytes); err != nil {
		return err
	}
	return nil
}

func integ_addChecksum (currentFile *integrity_fileCard) (error) {
	var err error
	if err = integ_writeChecksum(currentFile); err != nil {
		return err
	}
    if err = integ_confirmChecksum(currentFile, currentFile.checksum); err != nil {
		return err
	}
	return nil
}

func integ_confirmChecksum (currentFile *integrity_fileCard, testChecksum string) (error) {
	var err error
	var xtattrbChecksum string
	if xtattrbChecksum, err = integ_getChecksumRaw(currentFile.fullpath); err != nil {
		return err
	}
	if testChecksum != xtattrbChecksum {
		fmt.Fprintf(os.Stderr, "%s : Calculated checksum and filesystem read checksum differ!\n",currentFile.fullpath)
		fmt.Fprintf(os.Stderr, " ├── xatr; [%s]\n", xtattrbChecksum)
		fmt.Fprintf(os.Stderr, " └── calc; [%s]\n", testChecksum)
		//ToDo: Define new error
		return err
	}
	return nil
}

func integ_checkChecksum (currentFile *integrity_fileCard) (error) {
	var err error
	var xtattrbChecksum string

	if err = integ_generateChecksum(currentFile); err != nil {
		return err
	}
	if xtattrbChecksum, err = integ_getChecksumRaw(currentFile.fullpath); err != nil {
		return err
	}
	//xtattrbChecksum= "tiger";
	if currentFile.checksum != xtattrbChecksum {
		return fmt.Errorf("%s : Calculated checksum and filesystem read checksum differ!\n ├── xatr; [%s]\n └── calc; [%s]",currentFile.fullpath, xtattrbChecksum, currentFile.checksum)
	}

	if err = integ_confirmChecksum(currentFile, currentFile.checksum); err != nil {
		return err
	}
	return nil
}

func handle_path(path string, fileinfo os.FileInfo, err error) error {
	if ! fileinfo.IsDir() {
		var currentFile integrity_fileCard
		currentFile.FileInfo = &fileinfo
		currentFile.fullpath = path

		switch config.Action {
			case "list" , "list_trim":
				if err = integ_getChecksum(&currentFile); err != nil {
					var errorString string
					errorString = err.Error();
					if strings.Contains(errorString, "attribute not found") {
						if config.Verbose {
							fmt.Printf("%s : no checksum stored\n", path)
						}
					} else {
						fmt.Printf("Error; %s\n", err.Error());
					}
				} else {
					if config.Action == "list_trim" {
						// ToDO: change this to use the pointer in the struct
						fmt.Printf("%s : %s\n", fileinfo.Name() , currentFile.checksum)
					} else {
						fmt.Printf("%s : %s\n", path, currentFile.checksum)
					}
				}
			case "delete":
				if err = integ_removeChecksum(&currentFile); err != nil {
					if config.Verbose {
						fmt.Printf("Error removing checksum; %s\n", err.Error());
					}
				}
			case "write":
				if err = integ_addChecksum(&currentFile); err != nil {
					if config.Verbose {
						fmt.Printf("Error adding checksum; %s\n", err.Error());
					}
				} else {
					if config.Verbose {
						fmt.Printf("%s : %s : %s : added\n", path, "sha1", currentFile.checksum)
					} else {
						fmt.Printf("%s : added\n", path)
					}
				}
			case "check":
				if err = integ_checkChecksum(&currentFile); err != nil {
					if config.Verbose {
						fmt.Printf("Error checking checksum; %s\n", err.Error());
					} else {
						fmt.Printf("%s : FAILED\n", path)
					}
				} else {
					fmt.Printf("%s : PASSED\n", path)
				}
			default:
				fmt.Fprintf(os.Stderr, "Error : Unknown action \"%s\"\n",config.Action)
				os.Exit(2)
		}
	}
	return nil
}

func main() {

	for _, path := range flag.Args() {

		pathStat, err := os.Stat(path)
		// If we can stat the given file
		if err != nil {
			fmt.Fprintf(os.Stderr, "[%s] Error : there was an issue reading from this file\n└── %s\n", path, err)
			continue
		}

		if (pathStat.IsDir()) {
			// Walk the directory structure
			err := filepath.Walk(path, handle_path)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			path_fileinfo,err := os.Stat(path)
			handle_path(path, path_fileinfo, err)
		}
	}

}
