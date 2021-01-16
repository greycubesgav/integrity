package main

import (
	"crypto"
	_ "crypto/md5"
	_ "crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/pborman/getopt/v2"
	"github.com/pkg/xattr"
	_ "golang.org/x/crypto/blake2b"
	_ "golang.org/x/crypto/md4"
	_ "golang.org/x/crypto/ripemd160"
	_ "golang.org/x/crypto/sha3"
	"hash"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type integrity_fileCard struct {
	FileInfo    *os.FileInfo
	fullpath    string
	checksum    string
	digest_type crypto.Hash
	digest_name string
}

// ToDo Add option to skip mac files http://www.westwind.com/reference/OS-X/invisibles.html
// ToDo change errors to summarise at end like rsync - some errors occured
// ToDo check all errors goto stderr all normal messages go to stdout

var config *Config = NewConfig()

func integ_testChecksumStored(currentFile *integrity_fileCard) (bool, error) {
	var err error
	if _, err = xattr.Get(currentFile.fullpath, config.xattribute_fullname); err != nil {
		var errorString string
		errorString = err.Error()
		if strings.Contains(errorString, "attribute not found") || strings.Contains(errorString, "no data available") {
			// We got an error with attribute not found so simply return false and no error
			return false, nil
		} else {
			// We got a different error so return false and the error
			return false, err
		}
	}
	// We must have an attribute stored
	return true, nil
}

func integ_swapXattrib(currentFile *integrity_fileCard) error {
	// ToDo: add new custom error for cases where none of the old names were found
	//  Outout != RENAMED => SKIPPED
	var err error
	var data []byte
	var found bool = false

	attributeNames := []string{"user.integ.sha1", "integ.sha1", "user.integrity.sha1"}

	for _, oldAttribute := range attributeNames {
		if runtime.GOOS == "linux" {
			oldAttribute = "user." + oldAttribute
		}
		data, err = xattr.Get(currentFile.fullpath, oldAttribute)
		if err != nil {
			var errorString string
			errorString = err.Error()
			if strings.Contains(errorString, "attribute not found") {
				if config.Verbose {
					fmt.Printf("%s : Didn't find old attribute: %s\n", currentFile.fullpath, oldAttribute)
				}
			} else {
				// We got a different error looking for the attribute
				return err
			}

		} else {
			// We must have found an old attribute
			found = true

			if config.Verbose {
				fmt.Printf("%s : Found old attribute [%s] : Setting new attribute: [%s]\n", currentFile.fullpath, oldAttribute, config.xattribute_fullname)
			}

			if err = xattr.Set(currentFile.fullpath, config.xattribute_fullname, data); err != nil {
				return err
			}

			if err = xattr.Remove(currentFile.fullpath, oldAttribute); err != nil {
				return err
			}
		}
	}
	if !found {
		// We've not found any of the old attributes
		err = errors.New("No old attributes found")
		return err
	}
	return nil
}

func integ_getChecksumRaw(path string, digest_name string) (string, error) {
	var err error
	var data []byte
	if data, err = xattr.Get(path, config.xattribute_fullname); err != nil {
		return "", err
	}
	return string(data[:len(data)]), nil
}

func integ_getChecksum(currentFile *integrity_fileCard) error {
	var err error
	currentFile.digest_name = config.DigestName
	if currentFile.checksum, err = integ_getChecksumRaw(currentFile.fullpath, currentFile.digest_name); err != nil {
		return err
	}
	return nil
}

// integ_removeChecksum tries to remove a defined checksum attribute
// if we get an error because the attribute didn't exist we suppress the error and simple return false
// if we get an other type of error we pass it back
// otherwise we assume all is well and return true
// this allows the outer code to determine if we actually removed an attribute or not
func integ_removeChecksum(currentFile *integrity_fileCard) (bool, error) {
	var err error
	if err = xattr.Remove(currentFile.fullpath, config.xattribute_fullname); err != nil {
		var errorString string
		errorString = err.Error()
		if strings.Contains(errorString, "attribute not found") {
			// We got an error with attribute not found so simply return false and no error
			return false, nil
		} else {
			// We got a different error so return false and the error
			return false, err
		}
	}
	// We must have removed the attribute
	return true, nil
}

func integ_generateChecksum(currentFile *integrity_fileCard) error {
	var err error

	fileHandle, err := os.Open(currentFile.fullpath)
	if err != nil {
		return err
	}
	defer fileHandle.Close()

	config.logObject.Debugf("integ_generateChecksum config.DigestName:%s\n", config.DigestName)

	if config.DigestName == "oshash" {
		currentFile.checksum, err = OSHashFromFilePath(currentFile.fullpath)
		if err != nil {
			return err
		}
	} else if config.DigestName == "phash" {
		currentFile.checksum, err = integrityPhashFromFile(currentFile.fullpath)
		if err != nil {
			return err
		}
	} else {
		var hashObj crypto.Hash = config.DigestHash
		if !hashObj.Available() {
			return fmt.Errorf("integ_generateChecksum: hash object [%s] not compiled in!", hashObj)
		}
		var hashFunc hash.Hash = hashObj.New()
		if _, err := io.Copy(hashFunc, fileHandle); err != nil {
			return err
		}
		currentFile.checksum = hex.EncodeToString(hashFunc.Sum(nil))
	}
	config.logObject.Debugf("integ_generateChecksum currentFile.checksum:%s\n", currentFile.checksum)
	return nil
}

func integ_addChecksum(currentFile *integrity_fileCard) error {
	var err error
	// Write a new checksum to the file
	if err = integ_writeChecksum(currentFile); err != nil {
		return err
	}
	// Confirm that the checksum written to the xatrib when read back matches the one in memory
	if err = integ_confirmChecksum(currentFile, currentFile.checksum); err != nil {
		return err
	}
	return nil
}

func integ_writeChecksum(currentFile *integrity_fileCard) error {
	var err error
	if err = integ_generateChecksum(currentFile); err != nil {
		return err
	}
	checksumBytes := []byte(currentFile.checksum)
	if err = xattr.Set(currentFile.fullpath, config.xattribute_fullname, checksumBytes); err != nil {
		return err
	}
	return nil
}

func integ_confirmChecksum(currentFile *integrity_fileCard, testChecksum string) error {
	var err error
	var xtattrbChecksum string
	if xtattrbChecksum, err = integ_getChecksumRaw(currentFile.fullpath, config.DigestName); err != nil {
		return err
	}
	if testChecksum != xtattrbChecksum {
		fmt.Fprintf(os.Stderr, "%s : Calculated checksum and filesystem read checksum differ!\n", currentFile.fullpath)
		fmt.Fprintf(os.Stderr, " ├── xatr; [%s]\n", xtattrbChecksum)
		fmt.Fprintf(os.Stderr, " └── calc; [%s]\n", testChecksum)
		//ToDo: Define new error
		return err
	}
	currentFile.digest_name = config.DigestName
	return nil
}

func integ_checkChecksum(currentFile *integrity_fileCard) error {
	var err error
	var xtattrbChecksum string

	if err = integ_generateChecksum(currentFile); err != nil {
		return err
	}
	if xtattrbChecksum, err = integ_getChecksumRaw(currentFile.fullpath, config.DigestName); err != nil {
		return err
	}
	//xtattrbChecksum= "tiger";
	if currentFile.checksum != xtattrbChecksum {
		return fmt.Errorf("%s : Calculated checksum and filesystem read checksum differ!\n ├── xatr; [%s]\n └── calc; [%s]", currentFile.fullpath, xtattrbChecksum, currentFile.checksum)
	}

	if err = integ_confirmChecksum(currentFile, currentFile.checksum); err != nil {
		return err
	}
	return nil
}

func integ_printChecksum(err error, currentFile *integrity_fileCard, fileDisplayPath string) error {
	// Pass in the fileDisplayPath so we only need to generate it once outside this function

	if err = integ_getChecksum(currentFile); err != nil {
		var errorString string
		errorString = err.Error()
		if strings.Contains(errorString, "attribute not found") {
			if config.Verbose {
				fmt.Printf("%s : %s : [no checksum stored in %s]\n", fileDisplayPath, config.DigestName, config.xattribute_fullname)
			}
		} else {
			fmt.Printf("%s : Error : %s\n", fileDisplayPath, err.Error())
			return err
		}
	} else {
		if config.DisplayFormat == "sha1sum" {
			if strings.HasPrefix(currentFile.digest_name, "sha") {
				fmt.Printf("%s *%s\n", currentFile.checksum, fileDisplayPath)
			}
		} else if config.DisplayFormat == "md5sum" {
			if strings.HasPrefix(currentFile.digest_name, "md5") {
				fmt.Printf("%s  %s\n", currentFile.checksum, fileDisplayPath)
			}
		} else if config.Verbose || config.Option_AllDigests {
			fmt.Printf("%s : %s : %s\n", fileDisplayPath, config.DigestName, currentFile.checksum)
		} else {
			fmt.Printf("%s : %s\n", fileDisplayPath, currentFile.checksum)
		}
	}
	return nil
}

func integ_generatefileDisplayPath(currentFile *integrity_fileCard) string {
	if config.Option_ShortPaths {
		var fileInfo os.FileInfo
		fileInfo = *currentFile.FileInfo
		return fileInfo.Name()
	} else {
		return currentFile.fullpath
	}
}

func handle_path(path string, fileinfo os.FileInfo, err error) error {

	// ToDo: Refactor output to use common print function
	//       Something that takes the file, the message, and whether its an error type or not?
	//       Polymorphic to add error on end if we're printing error?

	if !fileinfo.IsDir() {
		var currentFile integrity_fileCard
		currentFile.FileInfo = &fileinfo
		currentFile.fullpath = path

		// Generate the display path here as most options will need it
		var fileDisplayPath string
		fileDisplayPath = integ_generatefileDisplayPath(&currentFile)

		// Generate a list of digests to work on here to prevent very similar code blocks for 1 hash and multiple hashes
		var digestList map[string]crypto.Hash
		digestList = make(map[string]crypto.Hash)
		if config.Option_AllDigests {
			digestList = digestTypes
		} else {
			digestList[config.DigestName] = config.DigestHash
		}

		switch config.Action {
		case "list":
			for hashType := range digestList {
				config.DigestName = hashType
				config.DigestHash = digestTypes[hashType]
				if err = integ_printChecksum(err, &currentFile, fileDisplayPath); err != nil {
					// Only continue as the function would have printed any error already
					continue
				}
			}
		case "delete":
			for hashType := range digestList {
				config.DigestName = hashType
				config.DigestHash = digestTypes[hashType]
				var hadAttribute bool
				hadAttribute, err = integ_removeChecksum(&currentFile)
				if err != nil {
					if config.Verbose {
						fmt.Printf("%s : %s : Error removing checksum: %s\n", fileDisplayPath, config.DigestName, err.Error())
					}
				} else if !hadAttribute {
					if config.Verbose {
						fmt.Printf("%s : %s : no attribute\n", fileDisplayPath, config.DigestName)
					}
				} else {
					if config.Verbose {
						fmt.Printf("%s : %s : removed\n", fileDisplayPath, config.DigestName)
					}
				}
			}

		case "add":
			if !config.Option_Force {
				var haveDigestStored bool
				haveDigestStored, err = integ_testChecksumStored(&currentFile)
				if err != nil {
					if config.Verbose {
						fmt.Printf("%s : FAILED : Error testing for stored checksum; %s\n", fileDisplayPath, err.Error())
					} else {
						fmt.Printf("%s : FAILED\n", fileDisplayPath)
					}
					return nil
				} else if haveDigestStored {
					fmt.Printf("%s : %s : skipped\n", fileDisplayPath, config.DigestName)
					return nil
				}
			}

			// If we've reached here we must want to add the checksum
			if err = integ_addChecksum(&currentFile); err != nil {
				if config.Verbose {
					fmt.Printf("%s : %s : Error : Error adding checksum; %s\n", fileDisplayPath, config.DigestName, err.Error())
				} else {
					fmt.Printf("%s : FAILED\n", fileDisplayPath)
				}
			} else {
				if config.Verbose {
					fmt.Printf("%s : %s : %s : added\n", fileDisplayPath, currentFile.digest_name, currentFile.checksum)
				} else {
					fmt.Printf("%s : %s : added\n", fileDisplayPath, currentFile.digest_name)
				}
			}

		case "check":
			var haveDigestStored bool
			haveDigestStored, err = integ_testChecksumStored(&currentFile)
			if haveDigestStored {
				if err = integ_checkChecksum(&currentFile); err != nil {
					if config.Verbose {
						fmt.Printf("Error checking checksum; %s\n", err.Error())
					} else {
						fmt.Printf("%s : FAILED\n", fileDisplayPath)
					}
				} else {
					if config.Verbose {
						fmt.Printf("%s : %s : %s : PASSED\n", fileDisplayPath, currentFile.digest_name, currentFile.checksum)
					} else {
						fmt.Printf("%s : %s : PASSED\n", fileDisplayPath, currentFile.digest_name)
					}
				}
			} else {
				if config.Verbose {
					fmt.Printf("%s : %s : no checksum, skipped\n", fileDisplayPath, config.DigestName)
				}
				return nil
			}
		case "transform":
			if err = integ_swapXattrib(&currentFile); err != nil {
				errorString := err.Error()
				if strings.Contains(errorString, "No old attributes found") {
					fmt.Printf("%s : %s : SKIPPED\n", fileDisplayPath, config.xattribute_fullname)
				} else {
					if config.Verbose {
						fmt.Printf("Error rename checksum; %s\n", err.Error())
					} else {
						fmt.Printf("%s : FAILED\n", fileDisplayPath)
					}
				}
			} else {
				fmt.Printf("%s : %s : RENAMED\n", fileDisplayPath, config.xattribute_fullname)
			}
		default:
			fmt.Fprintf(os.Stderr, "Error : Unknown action \"%s\"\n", config.Action)
			os.Exit(2)
		}
	}
	return nil
}

func integrityLogf(fmt_ string, args ...interface{}) {
	programCounter, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(programCounter)
	prefix := fmt.Sprintf("[%s:%s %d] %s", file, fn.Name(), line, fmt_)
	fmt.Printf(prefix, args...)
	//fmt.Println()
}

func main() {

	for _, path := range getopt.Args() {

		pathStat, err := os.Stat(path)
		// If we can stat the given file
		if err != nil {
			fmt.Fprintf(os.Stderr, "[%s] Error : there was an issue reading from this file\n└── %s\n", path, err)
			continue
		}

		if pathStat.IsDir() {
			if config.Option_Recursive {
				// Walk the directory structure
				err := filepath.Walk(path, handle_path)
				if err != nil {
					config.logObject.Fatal(err)
				}
			} else {
				if config.Verbose {
					fmt.Printf("%s : skipping directory\n", path)
				}
			}
		} else {
			path_fileinfo, err := os.Stat(path)
			handle_path(path, path_fileinfo, err)
		}
	}

}
