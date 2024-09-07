package integrity

import (
	"crypto"
	_ "crypto/md5"
	_ "crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/pborman/getopt/v2"
	"github.com/pkg/xattr"
	_ "golang.org/x/crypto/blake2b"
	_ "golang.org/x/crypto/sha3"
)

type integrity_fileCard struct {
	FileInfo    *os.FileInfo
	fullpath    string
	checksum    string
	digest_name string
}

// ToDo Add option to skip mac files http://www.westwind.com/reference/OS-X/invisibles.html
// ToDo change errors to summarise at end like rsync - some errors occured
// ToDo check all errors goto stderr all normal messages go to stdout

var config *Config = nil

func integ_testChecksumStored(currentFile *integrity_fileCard) (bool, error) {
	var err error
	if _, err = xattr.Get(currentFile.fullpath, config.xattribute_fullname); err != nil {
		var errorString string = err.Error()
		if strings.Contains(errorString, "attribute not found") || strings.Contains(errorString, "no data available") {
			// We got an error with attribute not found (darwin) or no data available (linux) so simply return false and no error
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
			var errorString string = err.Error()
			if strings.Contains(errorString, "attribute not found") || strings.Contains(errorString, "no data available") {
				switch config.VerboseLevel {
				case 0:
					// Don't print anything we're 'quiet'
				case 1, 2:
					fmt.Printf("%s : Didn't find old attribute: %s\n", currentFile.fullpath, oldAttribute)
				}
			} else {
				// We got a different error looking for the attribute
				return err
			}

		} else {
			// We must have found an old attribute
			found = true

			switch config.VerboseLevel {
			case 0, 1:
				// Don't print anything we're 'quiet'
			case 2:
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
		err = errors.New("no old attributes found")
		return err
	}
	return nil
}

func integ_getChecksumRaw(path string) (string, error) {
	var err error
	var data []byte
	if data, err = xattr.Get(path, config.xattribute_fullname); err != nil {
		return "", err
	}
	return string(data), nil
}

func integ_getChecksum(currentFile *integrity_fileCard) error {
	var err error
	currentFile.digest_name = config.DigestName
	if currentFile.checksum, err = integ_getChecksumRaw(currentFile.fullpath); err != nil {
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
		var errorString string = err.Error()
		if strings.Contains(errorString, "attribute not found") || strings.Contains(errorString, "no data available") {
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
		currentFile.checksum, err = oshashFromFilePath(currentFile.fullpath)
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
			return fmt.Errorf("integ_generateChecksum: hash object [%s] not supported", config.DigestHash)
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
	if xtattrbChecksum, err = integ_getChecksumRaw(currentFile.fullpath); err != nil {
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
	if xtattrbChecksum, err = integ_getChecksumRaw(currentFile.fullpath); err != nil {
		return err
	}
	if currentFile.checksum != xtattrbChecksum {
		return fmt.Errorf("calculated checksum and filesystem read checksum differ!\n ├── stored [%s]\n └── calc'd [%s]", xtattrbChecksum, currentFile.checksum)
	}
	if err = integ_confirmChecksum(currentFile, currentFile.checksum); err != nil {
		return err
	}
	return nil
}

func integ_printChecksum(currentFile *integrity_fileCard, fileDisplayPath string) error {
	// Pass in the fileDisplayPath so we only need to generate it once outside this function
	var err error
	if err = integ_getChecksum(currentFile); err != nil {
		var errorString string = err.Error()
		// Two different errors can be returned depending on the OS
		if strings.Contains(errorString, "attribute not found") || strings.Contains(errorString, "no data available") {
			switch config.VerboseLevel {
			case 0:
				// Don't print anything we're 'quiet'
			case 1:
				fmt.Printf("%s : %s : [none]\n", fileDisplayPath, config.DigestName)
			case 2:
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
		} else {
			fmt.Printf("%s : %s : %s\n", fileDisplayPath, config.DigestName, currentFile.checksum)
		}

	}
	return nil
}

func integ_generatefileDisplayPath(currentFile *integrity_fileCard) string {
	if config.Option_ShortPaths {
		var fileInfo os.FileInfo = *currentFile.FileInfo
		return fileInfo.Name()
	} else {
		return currentFile.fullpath
	}
}

func handle_path(path string, fileinfo os.FileInfo, err error) error {

	if err != nil {
		// Handle the error and return it to stop walking
		fmt.Printf("Error walking the path %v: %v\n", path, err)
		return err
	}

	// ToDo: Refactor output to use common print function
	//       Something that takes the file, the message, and whether its an error type or not?
	//       Polymorphic to add error on end if we're printing error?

	if !fileinfo.IsDir() {
		var currentFile integrity_fileCard
		currentFile.FileInfo = &fileinfo
		currentFile.fullpath = path

		// Generate the display path here as most options will need it
		var fileDisplayPath string = integ_generatefileDisplayPath(&currentFile)

		//----------------------------------------------------------------------------
		// ToDo: this following 2 blocks are ran for every file, this could be optimised to run 1 during config setup
		// Generate a list of digests to work on here to prevent very similar code blocks for 1 hash and multiple hashes
		var digestList map[string]crypto.Hash
		digestList = make(map[string]crypto.Hash)
		if config.Option_AllDigests {
			digestList = digestTypes
		} else {
			digestList[config.DigestName] = config.DigestHash
		}
		// Sort the list of digestNames we're running against
		digestNames := make([]string, 0, len(digestList)) // The sorted list of digestNames we'll run against
		for digestName := range digestList {
			digestNames = append(digestNames, digestName)
		}
		// Add the two digests that don't come from crypto.Hash
		//digestNames = append(digestNames, "oshash")
		//digestNames = append(digestNames, "phash")
		sort.Strings(digestNames)
		//----------------------------------------------------------------------------

		switch config.Action {
		case "list":
			for _, digestName := range digestNames {
				config.DigestName = digestName
				config.DigestHash = digestTypes[digestName]
				config.xattribute_fullname = config.Xattribute_prefix + config.DigestName
				if err = integ_printChecksum(&currentFile, fileDisplayPath); err != nil {
					// Only continue as the function would have printed any error already
					continue
				}
			}

		case "delete":
			for hashType := range digestList {
				config.DigestName = hashType
				config.DigestHash = digestTypes[hashType]
				config.xattribute_fullname = config.Xattribute_prefix + config.DigestName
				var hadAttribute bool
				hadAttribute, err = integ_removeChecksum(&currentFile)
				if err != nil {
					switch config.VerboseLevel {
					case 0:
						// Don't print anything we're 'quiet'
						// Does this make sense here? Do we want to still print this error if we're 'quiet'?
					case 1, 2:
						fmt.Printf("%s : %s : Error removing checksum: %s\n", fileDisplayPath, config.DigestName, err.Error())
					}
				} else if !hadAttribute {
					switch config.VerboseLevel {
					case 0:
						// Don't print anything we're 'quiet'
						// Does this make sense here? Do we want to still print this error if we're 'quiet'?
					case 1, 2:
						fmt.Printf("%s : %s : no attribute\n", fileDisplayPath, config.DigestName)
					}
				} else {
					switch config.VerboseLevel {
					case 0:
						// Don't print anything we're 'quiet'
					case 1, 2:
						fmt.Printf("%s : %s : removed\n", fileDisplayPath, config.DigestName)
					}
				}
			}

		case "add":
			if !config.Option_Force {
				var haveDigestStored bool
				haveDigestStored, err = integ_testChecksumStored(&currentFile)
				if err != nil {
					switch config.VerboseLevel {
					case 0, 1:
						fmt.Printf("%s : FAILED\n", fileDisplayPath)
					case 2:
						fmt.Printf("%s : FAILED : Error testing for existing checksum; %s\n", fileDisplayPath, err.Error())
					}
					return nil
				} else if haveDigestStored {
					switch config.VerboseLevel {
					case 0:
						// Don't print anything we're 'quiet'
					case 1:
						fmt.Printf("%s : %s : skipped\n", fileDisplayPath, config.DigestName)
					case 2:
						fmt.Printf("%s : %s : We already have a checksum stored, skipped\n", fileDisplayPath, config.DigestName)
					}
					return nil
				}
			}

			// If we've reached here we must want to add the checksum
			if err = integ_addChecksum(&currentFile); err != nil {
				switch config.VerboseLevel {
				case 0:
					fmt.Printf("%s : FAILED\n", fileDisplayPath)
				case 1:
					fmt.Printf("%s : %s : FAILED\n", fileDisplayPath, config.DigestName)
				case 2:
					fmt.Printf("%s : %s : FAILED : Error adding checksum; %s\n", fileDisplayPath, config.DigestName, err.Error())
				}
			} else {
				switch config.VerboseLevel {
				case 0:
					// Don't print anything we're 'quiet'
				case 1:
					fmt.Printf("%s : %s : added\n", fileDisplayPath, currentFile.digest_name)
				case 2:
					fmt.Printf("%s : %s : %s : added\n", fileDisplayPath, currentFile.digest_name, currentFile.checksum)
				}
			}

		case "check":
			var haveDigestStored bool
			if haveDigestStored, err = integ_testChecksumStored(&currentFile); err != nil {
				fmt.Fprintf(os.Stderr, "%s : failed checking if checksum was stored : %s\n", fileDisplayPath, err)
				return nil
			} else {
				if haveDigestStored {
					if err = integ_checkChecksum(&currentFile); err != nil {
						switch config.VerboseLevel {
						case 0:
							fmt.Printf("%s : FAILED\n", fileDisplayPath)
						case 1:
							fmt.Printf("%s : %s : FAILED\n", fileDisplayPath, config.DigestName)
						case 2:
							fmt.Printf("%s : %s : FAILED : %s\n", config.DigestName, fileDisplayPath, err.Error())
						}
					} else {
						switch config.VerboseLevel {
						case 0:
							// Don't print anything we're 'quiet'
						case 1:
							fmt.Printf("%s : %s : PASSED\n", fileDisplayPath, currentFile.digest_name)
						case 2:
							fmt.Printf("%s : %s : %s : PASSED\n", fileDisplayPath, currentFile.digest_name, currentFile.checksum)
						}
					}
				} else {
					switch config.VerboseLevel {
					case 0:
						// Musing: is it an 'error' if we don't have a checksum?
						// Answer: "no" We have 2 states for no output during check,
						// The file has a checksum and it is correct or it doesn't have a checksum
						// The assumption here is if we are quiet and don't have a checksum the file
						// Isn't important enough to check
					case 1:
						fmt.Printf("%s : %s : No checksum\n", fileDisplayPath, config.DigestName)
					case 2:
						fmt.Printf("%s : %s : No checksum, skipped\n", fileDisplayPath, config.DigestName)
					}
					return nil
				}
			}

		case "transform":
			if err = integ_swapXattrib(&currentFile); err != nil {
				errorString := err.Error()
				if strings.Contains(errorString, "No old attributes found") {
					switch config.VerboseLevel {
					case 0:
						// Don't print anything we're 'quiet'
					case 1:
						fmt.Printf("%s : %s : SKIPPED\n", fileDisplayPath, currentFile.digest_name)
					case 2:
						fmt.Printf("%s : %s : SKIPPED : No old attributes found\n", fileDisplayPath, currentFile.digest_name)
					}
				} else {
					switch config.VerboseLevel {
					case 0:
						fmt.Printf("%s : ERROR\n", fileDisplayPath)
					case 1:
						fmt.Printf("%s : %s : ERROR : Error renaming checksum\n", fileDisplayPath, currentFile.digest_name)
					case 2:
						fmt.Printf("%s : %s : ERROR : Error renaming checksum : %s\n", fileDisplayPath, currentFile.digest_name, err.Error())
					}
				}
			} else {
				switch config.VerboseLevel {
				case 0:
					// Don't print anything we're 'quiet'
				case 1:
					fmt.Printf("%s : %s : RENAMED\n", fileDisplayPath, currentFile.digest_name)
				case 2:
					fmt.Printf("%s : %s : Renamed old integrity attribute\n", fileDisplayPath, config.xattribute_fullname)
				}
			}
		default:
			fmt.Fprintf(os.Stderr, "Error : Unknown action \"%s\"\n", config.Action)
			config.returnCode = 6
			return errors.New("unknown action")
		}
	}
	return nil
}

func Run() int {

	config = newConfig()

	config.logObject.Debugf("integrity.Run()\n")
	config.logObject.Debugf("config.returnCode: %d\n", config.returnCode)

	switch config.returnCode {
	case 0:
		// config.returnCode=0 reserved for success
	case 1:
		// config.returnCode=1 reserved for show help runs, we show output and then exit but it wasn't an error
		return 0
	default:
		return config.returnCode
	}

	for _, path := range getopt.Args() {

		pathStat, err := os.Stat(path)
		// If we can stat the given file
		if err != nil {
			var errorString string = err.Error()
			if strings.Contains(errorString, "no such file or directory") {
				fmt.Fprintf(os.Stderr, "%s : no such file or directory\n", path)
				config.returnCode = 10
				continue
			}
			fmt.Fprintf(os.Stderr, "%s : ERROR : %s\n", path, err)
			config.returnCode = 12
			continue
		}

		if pathStat.IsDir() {
			if config.Option_Recursive {
				// Walk the directory structure
				err := filepath.Walk(path, handle_path)
				if err != nil {
					//config.logObject.Fatal(err)
					config.logObject.Error(err)
					return 1
				}
			} else {
				switch config.VerboseLevel {
				case 0, 1:
					// Don't print anything we're 'quiet' / this is not an error
				case 2:
					fmt.Printf("%s : skipping directory\n", path)
				}
			}
		} else {
			path_fileinfo, err := os.Stat(path)
			handle_path(path, path_fileinfo, err)
		}
	}
	return config.returnCode
}
