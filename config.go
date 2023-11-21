package main

import (
	"crypto"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pborman/getopt/v2"
	"github.com/sirupsen/logrus"
)

const integrity_website = "https://github.com/greycubesgav/integrity"
const xattribute_name = "integrity"
const env_name_prefix = "INTEGRITY"

var digestTypes = map[string]crypto.Hash{
	"md4":         crypto.MD4,
	"md5":         crypto.MD5,
	"sha1":        crypto.SHA1,
	"sha224":      crypto.SHA224,
	"sha256":      crypto.SHA256,
	"sha384":      crypto.SHA384,
	"sha512":      crypto.SHA512,
	"md5sha1":     crypto.MD5SHA1,
	"ripemd160":   crypto.RIPEMD160,
	"sha3_224":    crypto.SHA3_224,
	"sha3_256":    crypto.SHA3_256,
	"sha3_384":    crypto.SHA3_384,
	"sha3_512":    crypto.SHA3_512,
	"sha512_224":  crypto.SHA512_224,
	"sha512_256":  crypto.SHA512_256,
	"blake2s_256": crypto.BLAKE2s_256,
	"blake2b_256": crypto.BLAKE2b_256,
	"blake2b_384": crypto.BLAKE2b_384,
	"blake2b_512": crypto.BLAKE2b_512,
}

type Config struct {
	ShowHelp    bool
	ShowVersion bool
	ShowInfo    bool
	//TestChecksum				bool
	Verbose          bool
	DigestHash       crypto.Hash
	DigestName       string
	Action           string
	DisplayFormat    string
	Action_Add       bool
	Action_Delete    bool
	Action_List      bool
	Action_Transform bool
	//Option_List_sha1sum		bool
	//Option_List_md5sum		bool
	Action_Check         bool
	Option_Force         bool
	Option_ShortPaths    bool
	Option_Recursive     bool
	Option_AllDigests    bool
	Option_DefaultDigest bool
	xattribute_fullname  string
	logLevelName         string
	logObject            *logrus.Logger
}

func NewConfig() *Config {
	var c *Config = &Config{
		ShowHelp:         false,
		ShowVersion:      false,
		ShowInfo:      	  false,
		Action_Check:     false,
		Action_Add:       false,
		Action_Delete:    false,
		Action_List:      false,
		Action_Transform: false,
		//Option_List_sha1sum:	false,
		//Option_List_md5sum:		false,
		Option_Force:         false,
		Option_ShortPaths:    false,
		Option_Recursive:     false,
		Option_AllDigests:    false,
		Option_DefaultDigest: false,
		//TestChecksum:			false,
		Verbose:             false,
		DigestHash:          crypto.SHA1,
		DigestName:          "sha1",
		DisplayFormat:       "",
		Action:              "check",
		xattribute_fullname: "",
		logLevelName:        "info",
		logObject:           logrus.New(),
	}
	c.ParseCmdlineOpt()
	return c
}

func (c *Config) ParseCmdlineOpt() {
	// Get the digest from the name of the command
	// e.g. integrity.sha1, integrity.md5

	cmdName := filepath.Base(os.Args[0])
	cmdHash := strings.Split(cmdName, ".")
	if len(cmdHash) > 1 {
		c.DigestName = cmdHash[1]
	}

	getopt.FlagLong(&c.ShowHelp, "help", 'h', "show this help")

	getopt.FlagLong(&c.ShowVersion, "version", 0, "show version")
	getopt.FlagLong(&c.ShowInfo, "info", 0, "show information")

	getopt.FlagLong(&c.Action_Check, "check", 'c', "check the checksum of the file matches the one stored in the extended attributes [default]")
	getopt.FlagLong(&c.Action_Add, "add", 'a', "calculate the checksum of the file and add it to the extended attributes")
	getopt.FlagLong(&c.Action_Delete, "delete", 'd', "delete a checksum stored for a file")
	getopt.FlagLong(&c.Action_List, "list", 'l', "list the checksum stored for a file")
	getopt.FlagLong(&c.Action_Transform, "fix-old", 0, "fix an old extended attribute value name to the current format")

	//getopt.FlagLong(&c.Option_List_sha1sum, "sha1sum", 0, "list the checksum stored for a file as per the output of sha1sum, note this does not exclude the use of other digest formats!")
	//getopt.FlagLong(&c.Option_List_md5sum,   "md5sum", 0, "list the checksum stored for a file as per the output of md5sum, note this does not exclude the use of other digest formats!")

	getopt.FlagLong(&c.Option_AllDigests, "all", 'x', "include all digests, not just the default digest. Only applies to --delete and --list options")

	getopt.FlagLong(&c.Option_Force, "force", 'f', "force the calculation and writing of a checksum even if one already exists (default behaviour is to skip files with checksums already stored)")

	getopt.FlagLong(&c.Verbose, "verbose", 'v', "set verbose output.")
	getopt.FlagLong(&c.logLevelName, "loglevel", 0, "set the logging level. One of: panic, fatal, error, warn, info, debug, trace.")

	getopt.FlagLong(&c.DigestName, "digest", 0, "set the digest method (see help for list of digest types available)")

	getopt.FlagLong(&c.Option_ShortPaths, "short-paths", 's', "show only file name when showing file names, useful for generating sha1sum files")

	getopt.FlagLong(&c.Option_Recursive, "recursive", 'r', "recurse into sub-directories")

	getopt.FlagLong(&c.DisplayFormat, "display-format", 0, "set the output display format (sha1sum, md5sum). Note: this only shows any checkfiles ")

	getopt.Parse()

	if c.logLevelName == "trace" {
		c.logObject.SetLevel(logrus.TraceLevel)
	} else if c.logLevelName == "debug" {
		c.logObject.SetLevel(logrus.DebugLevel)
	} else if c.logLevelName == "info" {
		c.logObject.SetLevel(logrus.InfoLevel)
	} else if c.logLevelName == "warn" {
		c.logObject.SetLevel(logrus.WarnLevel)
	} else if c.logLevelName == "fatal" {
		c.logObject.SetLevel(logrus.FatalLevel)
	} else if c.logLevelName == "panic" {
		c.logObject.SetLevel(logrus.PanicLevel)
	} else {
		c.logObject.SetLevel(logrus.InfoLevel)
	}

	c.logObject.Debugf("LogObjectlevel : [%s]\n", c.logObject.Level)

	if c.ShowHelp {
		printHelp()
		os.Exit(0)
	}

	// If we haven't been passed a digest name
	// Try and get it from the environment
	// If this doesn't work, set it to sha1
	if c.DigestName == "" {
		envDigest := os.Getenv(env_name_prefix +"_DIGEST")

		if envDigest != "" {
			c.DigestName = envDigest
		} else {
			c.DigestName = "sha1"
			c.Option_DefaultDigest = true
		}
	}

	if c.DigestName != "oshash" && c.DigestName != "phash" {
		if c.DigestHash = digestTypes[c.DigestName]; c.DigestHash == 0 {
			fmt.Fprintf(os.Stderr, "Error : unknown hash type '%s'\n", c.DigestName)
			c.logObject.Fatalf("Error : unknown hash type '%s'\n", c.DigestName)
			os.Exit(2)
		}
	}

	c.logObject.Debugf("DigestName: '%s'\n", c.DigestName)

	// Check the current OS and create the full xattribute name from the os, const and digest
	switch runtime.GOOS {
		case "darwin", "freebsd":
			c.xattribute_fullname = fmt.Sprintf("%s.%s", xattribute_name, c.DigestName)
		case "linux":
			c.xattribute_fullname = fmt.Sprintf("user.%s.%s", xattribute_name, c.DigestName)
		default:
			c.logObject.Fatalf("Error: non-supported OS type '%s'\n", runtime.GOOS)
			c.logObject.Fatalf("Supported OS types 'darwin, freebsd, linux'\n")
			os.Exit(3)
	}

	c.logObject.Debugf("c.xattribute_fullname: '%s'\n", c.xattribute_fullname)

	if c.DisplayFormat != "" && c.DisplayFormat != "sha1sum" && c.DisplayFormat != "md5sum" {
		fmt.Fprintf(os.Stderr, "Error : unknown display format '%s'\n Should be one of: sha1sum, md5sum\n", c.DisplayFormat)
		os.Exit(3)
	}

	if c.Action_Check {
		c.Action = "check"
	} else if c.Action_Delete {
		c.Action = "delete"
	} else if c.Action_Add {
		c.Action = "add"
	} else if c.Action_List {
		c.Action = "list"
	} else if c.Action_Transform {
		c.Action = "transform"
	}

	// Show the version of the app and exit
	if c.ShowVersion {
		fmt.Printf("%s\n", integrity_version)
		os.Exit(0)
	}

	if c.ShowInfo {
		fmt.Printf("integrity version: %s\n", integrity_version)
		fmt.Printf("integrity attribute: %s\n", c.xattribute_fullname)
		fmt.Printf("runtime environment: %s\n", runtime.GOOS)
		os.Exit(0)
	}

	if getopt.NArgs() == 0 {
		fmt.Printf("integrity version %s\n", integrity_version)
		fmt.Printf("Web site: %s\n", integrity_website)
		getopt.Usage()
		fmt.Fprint(os.Stderr, "Error : no arguments given\n")
		os.Exit(1)
	}

}

func printHelp() {
	fmt.Printf("integrity version %s\n", integrity_version)
	fmt.Printf("Web site: %s\n", integrity_website)
	fmt.Printf(`
integrity is a tool for calculating, storing and verifying checksums for files.
A number of different types of checksum are supported with the result stored in the file's
extended attributes. This allows the file to be moved between directories
or copied to another machine while retaining the checksum data along with the file.

This checksum data can be used to verify the integrity of the file at a later date and
ensure it's contents have not been changed or become corrupted.

The checksum data is also useful for efficiently finding duplicate files in
different directories.

Usage: integrity [OPTIONS] FILE|PATH

Options:`)
	getopt.Usage()
	fmt.Println(`
Usage Examples:

  Checking a file's integrity checksum
    integrity myfile.jpg
    > myfile.jpg : sha1 : PASSED

	integrity file_no_integrity_checksum.jpg
	> file_no_integrity_checksum.jpg : No checksums found

  Adding integrity data to a file, skip if the file already has integrity data
    integrity -a data_01.dat
    > data01.dat : sha1 : added

	integrity -a myfile.jpg
	> myfile.jpg : sha1 : exists, skipping

  Adding integrity data to a file, forcing a recalcuation if file already has integrity data
    integrity -a -f myfile.jpg
    > myfile.jpg : sha1 : added

  Checking the integrity of a list of files
    integrity *.jpg
    > myfile.jpg : sha1 : PASSED
    > wrong_checksum.jpg : sha1: FAILED

  Checking the integrity of a file with integrity data verbosely
    integrity -v myfile.jpg
    > myfile.jpg : sha1 : 65bb1872af65ed02db42f603c786f5ec7d392909 : PASSED

    integrity v wrong_checksum.jpg
    > Error checking sha1 checksum; wrong_checksum.jpg : Calculated checksum and filesystem read checksum differ!
      ├── calc; [32c48f2bca002218e7488d5d41bb9c82743a3392]
      └── disk; [3fc98aa337e328816416e179afc863a75ffb330a]

  List integrity data for default checksum, no verification
    integrity -l myfile.jpg
    > myfile.jpg : sha1 : 65bb1872af65ed02db42f603c786f5ec7d392909

  List integrity data for detault checksum, with current validity check
    integrity -l -c data_01.dat
    > data01.dat : sha1 : 65bb1872af65ed02db42f603c786f5ec7d392909 : PASSED

  Listing integrity data as shasum command output, note only shows any sha1 checksums
    integrity -l --display-format=sha1sum  data01.dat
    > 65bb1872af65ed02db42f603c786f5ec7d392909 *data01.dat

  Listing integrity data as md5sum command output, note only shows any md5 checksums
    integrity -l --display-format=md5sum  data01.dat
    > 65bb1872af65ed02db42f603c786f5ec7d392909  data01.dat

  List all checksums stored, not just the default/digest selected
    integrity -l -x data_01.dat
    > data_01.dat : md5 : 10c8d3e65b9243454b6f5f24e5f3197e
      data_01.dat : sha1 : ffccc1f78abcc5ac8b8434a5c4eeab75e64918ca

  Check all checksums stored
    integrity -c -x data_01.dat
    > data_01.dat : md5 : 10c8d3e65b9243454b6f5f24e5f3197e : PASSED
      data_01.dat : sha1 : ffccc1f78abcc5ac8b8434a5c4eeab75e64918ca : FAILED

  Check all checksums stored verbosely
    integrity -c -x -v data_01.dat
    > data_01.dat : md5 : 10c8d3e65b9243454b6f5f24e5f3197e : PASSED
    > Error checking sha1 checksum; "wrong_checksum.jpg" : Calculated checksum and filesystem read checksum differ!
      ├── calc; [32c48f2bca002218e7488d5d41bb9c82743a3392] : CALC
      └── disk; [3fc98aa337e328816416e179afc863a75ffb330a] : FAILED

  Remove the default digest's checksum data
    integrity -d data_01.dat
    > data_01.dat : sha1 : REMOVED

  Remove the all digests's checksum data
    integrity -d -x data_01.dat
    > data_01.dat : md5 : REMOVED
    > data_01.dat : sha1 : REMOVED

  Recursively add integrity data to all files within a directory structure
    integrity -a -r ~/data/

  Recursively list the checksum as shasum output
    integrity -l -r ~/data/

Further Information:

  When copying files across disks or machines extended attributes should be preserved to ensure
  the file's integrity data is also copied.

  Note: the destination filesystem must support extended attributes (see: Supported filesystems)

  For example:
    rsync -X source destination
    cp -p source destination

  The default digest can be set through an environment variable INTEGRITY_DIGEST. This allows for a you to set your prefered digest
  method without needing to set it on the command line each time.

  For example:
  	INTEGRITY_DIGEST='blake2s_256' integrity -a myfile.dat

Design Choices:

    * By default this utility is designed to quiet on output. i.e. when adding trying to add a checksum to a file with one
	  stored already, the app will simply skip over the file and continue. This is because the utility is meant to be run
	  over large numbers of data files which may or may not already have checksum data so output is kept to a minimum.

	  For example:

	     integrity -a -r directory/
	     Add the default checksum data too all files, 'added' will be shown for all files which had checksum data added.
		 Nothing will be shown for the others.

      Add the -v flag to see more verbose output.

	* The utility is designed to do "sensible" things with basic options

	  For example:

	     integrity -c file.dat
	     Check a file's default digest (sha1)
		 (The default digest type is sha1 unless overwritten by the environment variable INTEGRITY_DIGEST)

	     integrity -a file.dat
	     Add a checksum using the default digest (sha1)

	     integrity -d file.dat
	     Remove the default digest (sha1) data

         integrity -l file.dat
	     List the default digest (sha1) data

Supported Checksum Digest Algorithms:

    * md4
    * md5
    * sha1
    * sha224
    * sha256
    * sha384
    * sha512
    * md5sha1
    * ripemd160
    * sha3_224
    * sha3_256
    * sha3_384
    * sha3_512
    * sha512_224
    * sha512_256
    * blake2s_256
    * blake2b_256
    * blake2b_384
    * blake2b_512
    * oshash : media hashing algorithm as defined by opensubtitles (see: https://trac.opensubtitles.org/projects/opensubtitles/wiki/HashSourceCodes)
    * phash : perceptive image hash algorithm (Through https://github.com/corona10/goimagehash, see: https://www.hackerfactor.com/blog/index.php?/archives/432-Looks-Like-It.html)
  `)

}
