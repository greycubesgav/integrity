package integrity

import (
	_ "crypto/md5"
	_ "crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pborman/getopt/v2"
	"github.com/pkg/xattr"
	_ "golang.org/x/crypto/blake2b"
	_ "golang.org/x/crypto/blake2s"
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
				case 0, 1:
					// Don't print anything we're 'quiet'
				case 2:
					displayFileMessageNoDigest(currentFile.fullpath, fmt.Sprintf("old attribute not found : %s", oldAttribute))
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
				displayFileMessageNoDigest(currentFile.fullpath, fmt.Sprintf("Found old attribute [%s] : Setting new attribute: [%s]", oldAttribute, config.xattribute_fullname))
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
		hashObj := config.digestList[config.DigestName]
		if !hashObj.Available() {
			config.logObject.Debugf("integ_generateChecksum !hashObj.Available():%s\n", config.DigestName)
			return fmt.Errorf("integ_generateChecksum: hash object [%s] not supported", hashObj)
		}
		hashFunc := hashObj.New()
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
		return fmt.Errorf("calculated checksum and filesystem read checksum differ!\n ├── stored [%s]\n └── calc'd [%s]", xtattrbChecksum, currentFile.checksum)
	}
	currentFile.digest_name = config.DigestName
	return nil
}

func integ_checkChecksum(currentFile *integrity_fileCard) error {
	var err error
	// Generate the checksum from the file contents
	// Stores the generated checksum in currentFile.checksum
	if err = integ_generateChecksum(currentFile); err != nil {
		return err
	}
	// Check the checksum using the current file and the checksum just generated previously
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
				displayFileMessage(fileDisplayPath, "[none]")
			case 2:
				displayFileMessage(fileDisplayPath, fmt.Sprintf("[no checksum stored in %s]", config.xattribute_fullname))
			}

		} else {
			switch config.VerboseLevel {
			case 0, 1:
				// Always output errors if we're 'quiet'
				displayFileErrorMessage(fileDisplayPath, "FAILED")
			case 2:
				displayFileErrorMessage(fileDisplayPath, fmt.Sprintf("FAILED : Error reading checksum : %s", err.Error()))
			}
			return err
		}
	} else {
		displayFileMessage(fileDisplayPath, currentFile.checksum)
	}
	return nil
}

func displayFileMessageNoDigest(fileDisplayPath string, message string) {
	fmt.Fprintf(os.Stdout, "%s : %s\n", fileDisplayPath, message)
}

func displayFileErrorMessageNoDigest(fileDisplayPath string, message string) {
	fmt.Fprintf(os.Stderr, "%s : %s\n", fileDisplayPath, message)
}

func displayFileMessage(fileDisplayPath string, message string) {
	if config.DisplayFormat == "sha1sum" && strings.HasPrefix(config.DigestName, "sha") {
		fmt.Printf("%s *%s\n", message, fileDisplayPath)
	} else if config.DisplayFormat == "md5sum" && strings.HasPrefix(config.DigestName, "md5") {
		fmt.Printf("%s  %s\n", message, fileDisplayPath)
	} else {
		fmt.Fprintf(os.Stdout, "%s : %s : %s\n", fileDisplayPath, config.DigestName, message)
	}
}

func displayFileErrorMessage(fileDisplayPath string, message string) {
	fmt.Fprintf(os.Stderr, "%s : %s : %s\n", fileDisplayPath, config.DigestName, message)
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
	config.logObject.Debugf("handle_path: '%s'\n", path)
	if err != nil {
		config.logObject.Debugf("handle_path: error '%s'\n", err)
		if strings.Contains(err.Error(), "permission denied") {
			switch config.VerboseLevel {
			case 0, 1:
				// Always output errors even if we're 'quiet'
				displayFileErrorMessageNoDigest(path, "skipped")
			case 2:
				displayFileErrorMessageNoDigest(path, fmt.Sprintf("skipped : %s", err.Error()))
			}
			return filepath.SkipDir
		} else {
			// Handle the error and return it to stop walking
			fmt.Fprintf(os.Stderr, "Error walking the path : %v : %v\n", path, err)
			return err
		}
	}

	config.logObject.Debugf("no errors contiuing\n")

	if !fileinfo.IsDir() {
		var currentFile integrity_fileCard
		currentFile.FileInfo = &fileinfo
		currentFile.fullpath = path

		// Generate the display path here as most options will need it
		var fileDisplayPath string = integ_generatefileDisplayPath(&currentFile)

		switch config.Action {
		case "list":
			for _, digestName := range config.digestNames {
				config.DigestName = digestName
				config.xattribute_fullname = config.xattribute_prefix + config.DigestName
				config.logObject.Debugf("list: '%s'\n", config.xattribute_fullname)
				if err = integ_printChecksum(&currentFile, fileDisplayPath); err != nil {
					// Only continue as the function would have printed any error already
					continue
				}
			}

		case "delete":
			for _, digestName := range config.digestNames {
				config.DigestName = digestName
				config.xattribute_fullname = config.xattribute_prefix + config.DigestName
				config.logObject.Debugf("delete: '%s'\n", config.xattribute_fullname)
				hadAttribute, err := integ_removeChecksum(&currentFile)
				if err != nil {
					switch config.VerboseLevel {
					case 0, 1:
						// Always output errors if we're 'quiet'
						displayFileErrorMessage(fileDisplayPath, "FAILED")
					case 2:
						displayFileErrorMessage(fileDisplayPath, fmt.Sprintf("FAILED : Error removing checksum : %s", err.Error()))
					}
				} else if !hadAttribute {
					switch config.VerboseLevel {
					case 0:
						// Don't print anything we're 'quiet'
						// Does this make sense here? Do we want to still print this error if we're 'quiet'?
					case 1:
						displayFileMessage(fileDisplayPath, "no attribute")
					case 2:
						displayFileMessage(fileDisplayPath, "no checksum attribute found")
					}
				} else {
					switch config.VerboseLevel {
					case 0:
						// Don't print anything we're 'quiet'
					case 1:
						displayFileMessage(fileDisplayPath, "removed")
					case 2:
						displayFileMessage(fileDisplayPath, "removed checksum attribute")
					}
				}
			}

		case "add":
			for _, digestName := range config.digestNames {
				config.DigestName = digestName
				config.xattribute_fullname = config.xattribute_prefix + config.DigestName
				config.logObject.Debugf("add: '%s'\n", config.xattribute_fullname)
				if !config.Option_Force {
					var haveDigestStored bool
					haveDigestStored, err = integ_testChecksumStored(&currentFile)
					if err != nil {
						switch config.VerboseLevel {
						case 0, 1:
							// Always output errors even if we're 'quiet'
							displayFileErrorMessage(fileDisplayPath, "FAILED")
						case 2:
							displayFileErrorMessage(fileDisplayPath, fmt.Sprintf("FAILED : Error testing for existing checksum : %s", err.Error()))
						}
						return nil
					} else if haveDigestStored {
						switch config.VerboseLevel {
						case 0:
							// Don't print anything we're 'quiet'
						case 1:
							displayFileMessage(fileDisplayPath, "skipped")
						case 2:
							displayFileMessage(fileDisplayPath, "skipped : We already have a checksum stored")
						}
						continue
					}
				}

				// If we've reached here we must want to add the checksum
				if err = integ_addChecksum(&currentFile); err != nil {
					switch config.VerboseLevel {
					case 0, 1:
						// Always output errors even if we're 'quiet'
						displayFileErrorMessage(fileDisplayPath, "FAILED")
					case 2:
						displayFileErrorMessage(fileDisplayPath, fmt.Sprintf("FAILED : Error adding checksum : %s", err.Error()))
					}
				} else {
					switch config.VerboseLevel {
					case 0:
						// Don't print anything we're 'quiet'
					case 1:
						displayFileMessage(fileDisplayPath, "added")
					case 2:
						displayFileMessage(fileDisplayPath, fmt.Sprintf("%s : added", currentFile.checksum))
					}
				}
			}

		case "check":
			for _, digestName := range config.digestNames {
				config.DigestName = digestName
				config.xattribute_fullname = config.xattribute_prefix + config.DigestName
				config.logObject.Debugf("check: '%s'\n", config.xattribute_fullname)
				var haveDigestStored bool
				if haveDigestStored, err = integ_testChecksumStored(&currentFile); err != nil {
					switch config.VerboseLevel {
					case 0, 1:
						// Always output errors even if we're 'quiet'
						displayFileErrorMessage(fileDisplayPath, "FAILED")
					case 2:
						displayFileErrorMessage(fileDisplayPath, fmt.Sprintf("FAILED : failed checking if checksum was stored : %s", err.Error()))
					}
					return nil
				} else {
					if haveDigestStored {
						if err = integ_checkChecksum(&currentFile); err != nil {
							switch config.VerboseLevel {
							case 0, 1:
								// Always output errors even if we're 'quiet'
								displayFileErrorMessage(fileDisplayPath, "FAILED")
							case 2:
								displayFileErrorMessage(fileDisplayPath, fmt.Sprintf("FAILED : %s", err.Error()))
							}
						} else {
							switch config.VerboseLevel {
							case 0:
								// Don't print anything we're 'quiet'
							case 1:
								displayFileMessage(fileDisplayPath, "PASSED")
							case 2:
								displayFileMessage(fileDisplayPath, fmt.Sprintf("%s : PASSED", currentFile.checksum))
							}
						}
					} else {
						switch config.VerboseLevel {
						case 0:
							// Musing: is it an 'error' if we don't have a checksum?
							// Answer: "no" We have 2 states for no output during check,
							// The file has a checksum and it is correct or it doesn't have a checksum
							// The assumption here is if we are quiet and don't have a checksum the file
							// isn't important enough to check
						case 1:
							displayFileMessage(fileDisplayPath, "no checksum")
						case 2:
							displayFileMessage(fileDisplayPath, "no checksum, skipped")
						}
						return nil
					}
				}
			}

		case "transform":
			if err = integ_swapXattrib(&currentFile); err != nil {
				errorString := err.Error()
				if strings.Contains(errorString, "no old attributes found") {
					switch config.VerboseLevel {
					case 0:
						// Don't print anything we're 'quiet'
					case 1:
						displayFileMessageNoDigest(fileDisplayPath, "skipped")
					case 2:
						displayFileMessageNoDigest(fileDisplayPath, "skipped : No old attributes found")
					}
				} else {
					switch config.VerboseLevel {
					case 0:
						// Always output errors even if we're 'quiet'
						displayFileErrorMessage(fileDisplayPath, "ERROR")
					case 1:
						displayFileErrorMessage(fileDisplayPath, "ERROR : Error renaming checksum")
					case 2:
						displayFileErrorMessage(fileDisplayPath, fmt.Sprintf("ERROR : Error renaming checksum : %s", err.Error()))
					}
				}
			} else {
				switch config.VerboseLevel {
				case 0:
					// Don't print anything we're 'quiet'
				case 1:
					displayFileMessage(fileDisplayPath, "RENAMED")
				case 2:
					displayFileMessage(fileDisplayPath, "RENAMED : Renamed any old integrity attributes")
				}
			}
		default:
			fmt.Fprintf(os.Stderr, "Error : Unknown action \"%s\"\n", config.Action)
			config.returnCode = 9 // Unknown action
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

		// ToDo: Consider how to deal with symlinks, should be follow them?
		config.logObject.Debugf("path: '%s'\n", path)
		path_fileinfo, err := os.Stat(path)
		// If we can stat the given file
		if err != nil {
			errorString := err.Error()
			if strings.Contains(errorString, "no such file or directory") {
				fmt.Fprintf(os.Stderr, "%s : no such file or directory\n", path)
				config.returnCode = 10 // No such file or directory
				continue
			}
			displayFileErrorMessageNoDigest(path, fmt.Sprintf("ERROR : %s", err.Error()))
			config.returnCode = 12 // Error stating file
			continue
		}

		if path_fileinfo.IsDir() {
			config.logObject.Debugf("path is directory: recurse? '%t'\n", config.Option_Recursive)
			if config.Option_Recursive {
				// Walk the directory structure
				err := filepath.Walk(path, handle_path)
				if err != nil {
					config.logObject.Debugf("Error from filepath.Walk: err(%s)", err.Error())
					return 1
				}
			} else {
				switch config.VerboseLevel {
				case 0, 1:
					// Don't print anything we're 'quiet' / this is not an error
				case 2:
					displayFileMessageNoDigest(path, "skipping directory")
				}
			}
		} else {
			handle_path(path, path_fileinfo, err)
		}
	}

	config.logObject.Debugf("config.returnCode: %d\n", config.returnCode)
	return config.returnCode
}
