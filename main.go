package main

import (
	"fmt"
	"flag"
	"os"
	"path/filepath"
	"log"
	"github.com/pkg/xattr"
	"io"
	"encoding/hex"
	"strings"
	"crypto"
	"hash"
)

// Import the various supported hash libraries

import _ "golang.org/x/crypto/md4"
import _ "crypto/md5"
import _ "crypto/sha1"
import _ "crypto/sha256"
import _ "crypto/sha512"
import _ "golang.org/x/crypto/ripemd160"
import _ "golang.org/x/crypto/sha3"
import _ "golang.org/x/crypto/blake2b"

type integrity_fileCard struct {
	FileInfo 		*os.FileInfo
	fullpath		string
	checksum		string
	digest_type		crypto.Hash
	digest_name     string
}



var config *Config = NewConfig()

const xattribute_name = "integrity."




func integ_getChecksumRaw (path string, digest_name string) (string, error) {
	var err error
	var data []byte
	if data, err = xattr.Get(path, xattribute_name + digest_name ); err != nil {
		return "", err
	}
	return string(data[:len(data)]), nil
}

func integ_getChecksum (currentFile *integrity_fileCard) (error) {
	var err error
	currentFile.digest_name = config.DigestName
	if currentFile.checksum, err = integ_getChecksumRaw(currentFile.fullpath, currentFile.digest_name); err != nil {
		return err
	}
	return nil
}

func integ_hashGetName(hash crypto.Hash) (string, error) {
	switch hash {
	case crypto.MD4:
		return "md4",nil
	case crypto.MD5:
		return "md5",nil
	case crypto.SHA1:
		return "sha1",nil
	case crypto.SHA256:
		return "sha256",nil
	case crypto.SHA512:
		return "sha512",nil
	default:
		return "",fmt.Errorf("Unknown hash type: %d", hash)
	}
}

func integ_hashGetHashType (hashname string) (crypto.Hash, error) {
	switch hashname {
	case "md5":
		return crypto.MD5,nil
	case "sha1":
		return crypto.SHA1,nil
	case "sha256":
		return crypto.SHA256,nil
	case "sha512":
		return crypto.SHA512,nil
	default:
		return 0,fmt.Errorf("integ_hashGetHashType: Unknown hash type: %s", hashname)
	}
}

func integ_removeChecksum (currentFile *integrity_fileCard) (error) {
	var err error
	if err = xattr.Remove(currentFile.fullpath, xattribute_name + currentFile.digest_name); err != nil {
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

	var hashObj crypto.Hash = config.DigestHash
	if !hashObj.Available() {
		return fmt.Errorf("integ_generateChecksum: hash object [%s] not compiled in!", hashObj)
	}
	var hashFunc hash.Hash = hashObj.New()
	if _, err := io.Copy(hashFunc, fileHandle); err != nil {
		return err
	}
	currentFile.checksum = hex.EncodeToString(hashFunc.Sum(nil))
	return nil
}

func integ_addChecksum (currentFile *integrity_fileCard) (error) {
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

func integ_writeChecksum (currentFile *integrity_fileCard) (error) {
	var err error
	if err = integ_generateChecksum(currentFile); err != nil {
		return err
	}
	checksumBytes := []byte(currentFile.checksum)
	if err = xattr.Set(currentFile.fullpath, xattribute_name + config.DigestName, checksumBytes); err != nil {
		return err
	}
	return nil
}

func integ_confirmChecksum (currentFile *integrity_fileCard, testChecksum string) (error) {
	var err error
	var xtattrbChecksum string
	if xtattrbChecksum, err = integ_getChecksumRaw(currentFile.fullpath, config.DigestName); err != nil {
		return err
	}
	if testChecksum != xtattrbChecksum {
		fmt.Fprintf(os.Stderr, "%s : Calculated checksum and filesystem read checksum differ!\n",currentFile.fullpath)
		fmt.Fprintf(os.Stderr, " ├── xatr; [%s]\n", xtattrbChecksum)
		fmt.Fprintf(os.Stderr, " └── calc; [%s]\n", testChecksum)
		//ToDo: Define new error
		return err
	}
	currentFile.digest_name  = config.DigestName
	return nil
}

func integ_checkChecksum (currentFile *integrity_fileCard) (error) {
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
					var file_display string
					if config.Action == "list_trim" {
						// ToDO: change this to use the pointer in the currentFile struct
						file_display = fileinfo.Name()
					} else {
						file_display = path
					}
					if config.Verbose {
 						fmt.Printf("%s : %s : %s\n", file_display, currentFile.digest_name, currentFile.checksum)
					} else {
						fmt.Printf("%s : %s\n", file_display, currentFile.checksum)
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
						fmt.Printf("%s : %s : %s : added\n", path, currentFile.digest_name, currentFile.checksum)
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
