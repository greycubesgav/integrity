package main

import (
	"fmt"
	"os"
	"crypto"
	"github.com/pborman/getopt/v2"
)

const integrity_version = "0.1.10"
const integrity_website = "https://www.example.com"

var digestTypes = map[string]crypto.Hash {
	"md4" : crypto.MD4,
	"md5" : crypto.MD5,
	"sha1" : crypto.SHA1,
	"sha224" : crypto.SHA224,
	"sha256" : crypto.SHA256,
	"sha384" : crypto.SHA384,
	"sha512" : crypto.SHA512,
	"md5sha1" : crypto.MD5SHA1,
	"ripemd160" : crypto.RIPEMD160,
	"sha3_224" : crypto.SHA3_224,
	"sha3_256" : crypto.SHA3_256,
	"sha3_384" : crypto.SHA3_384,
	"sha3_512" : crypto.SHA3_512,
	"sha512_224" : crypto.SHA512_224,
	"sha512_256" : crypto.SHA512_256,
	"blake2s_256" : crypto.BLAKE2s_256,
	"blake2b_256" : crypto.BLAKE2b_256,
	"blake2b_384" : crypto.BLAKE2b_384,
	"blake2b_512" : crypto.BLAKE2b_512,
}


type Config struct {
	ShowHelp				bool
	//TestChecksum			bool
	Verbose					bool
	DigestHash				crypto.Hash
	DigestName				string
	Action					string
	DisplayFormat           string
	Action_Add				bool
	Action_Delete			bool
	Action_List				bool
	Action_Transform		bool
	//Option_List_sha1sum		bool
	//Option_List_md5sum		bool
	Action_Check			bool
	Option_Force			bool
	Option_ShortPaths		bool
	Option_Recursive		bool
	Option_AllDigests		bool
	Option_DefaultDigest	bool
}

func NewConfig() *Config {
	var c *Config = &Config{
		ShowHelp:				false,
		Action_Check:			false,
		Action_Add:				false,
		Action_Delete:			false,
		Action_List:			false,
		Action_Transform:		false,
		//Option_List_sha1sum:	false,
		//Option_List_md5sum:		false,
		Option_Force:			false,
		Option_ShortPaths:		false,
		Option_Recursive:		false,
		Option_AllDigests:		false,
		Option_DefaultDigest:	false,
		//TestChecksum:			false,
		Verbose:				false,
		DigestHash:				crypto.SHA1,
		DigestName:				"",
		DisplayFormat:          "",
		Action:					"check",
	}
	c.ParseCmdlineOpt()
	return c
}

func (c *Config) ParseCmdlineOpt() {

	getopt.FlagLong(&c.ShowHelp, "help", 'h', "show this help")
	getopt.FlagLong(&c.Action_Check, "check", 'c', "check the checksum of the file matches the one stored in the extended attributes")
	getopt.FlagLong(&c.Action_Add, "add", 'a', "calculate the checksum of the file and add it to the extended attributes")
	getopt.FlagLong(&c.Action_Delete, "delete", 'd', "delete a checksum stored for a file")
	getopt.FlagLong(&c.Action_List, "list", 'l', "list the checksum stored for a file")
	getopt.FlagLong(&c.Action_Transform, "transform", 't', "transform an old extended attribute value name to the current format")

	//getopt.FlagLong(&c.Option_List_sha1sum, "sha1sum", 0, "list the checksum stored for a file as per the output of sha1sum, note this does not exclude the use of other digest formats!")
	//getopt.FlagLong(&c.Option_List_md5sum,   "md5sum", 0, "list the checksum stored for a file as per the output of md5sum, note this does not exclude the use of other digest formats!")

	getopt.FlagLong(&c.Option_AllDigests, "all", 'x', "include all digests, not just the default digest. Only applies to --delete and --list options")

	getopt.FlagLong(&c.Option_Force, "force", 'f', "force the calculation and writing of a checksum even if one already exists (default behaviour is to skip files with checksums already stored)")

	getopt.FlagLong(&c.Verbose, "verbose", 'v', "print more verbose messages")

	getopt.FlagLong(&c.DigestName, "digest", 0, "set the digest method (see help for list of digest types available)")

	getopt.FlagLong(&c.Option_ShortPaths, "short-paths", 's', "show only file name when showing file names, useful for generating sha1sum files")

	getopt.FlagLong(&c.Option_Recursive, "recursive", 'r', "recurse into sub-directories")

	getopt.FlagLong(&c.DisplayFormat, "display-format", 0, "set the output display format (sha1sum, md5sum). Note: this only sets the output format, any digest type can be displayed!")

	getopt.Parse()

	if c.ShowHelp {
		printHelp()
		os.Exit(0)
	}

	if getopt.NArgs() == 0 {
		fmt.Printf("integrity version %s\n", integrity_version)
		fmt.Printf("Web site: %s\n", integrity_website)
		getopt.Usage()
		fmt.Fprint(os.Stderr, "Error : no arguments given\n")
		os.Exit(1)
	}

	// If we haven't been passed a digest name
	// Try and get it from the environment
	// If this doesn't work, set it to sha1
	if c.DigestName == "" {
		var envDigest string
		envDigest = os.Getenv("I_DIGEST")

		if envDigest != "" {
			c.DigestName = envDigest
		} else {
			c.DigestName = "sha1"
			c.Option_DefaultDigest = true
		}
	}

	if c.DigestHash = digestTypes[c.DigestName]; c.DigestHash == 0 {
		fmt.Fprintf(os.Stderr, "Error : unknown hash type '%s'\n", c.DigestName)
		os.Exit(2)
	}

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
}

func printHelp() {
	fmt.Printf("integrity version %s\n", integrity_version)
	fmt.Printf("Web site: %s\n", integrity_website)
	fmt.Printf(`
integrity is a file checksum and verification tool capable of calculating
a number of different types of checksum and storing the result in the file's
extended attributes. This allows the file to be moved between directories
or copied to another machine while maintaining the checksum data

This checksum data can be used to verify the integrity of the file and ensure
it's contents are still valid.

The checksum data is also useful for efficiently finding duplicate files in
different directories'

Usage: integrity [OPTIONS] FILE|PATH

Options:`)
	getopt.Usage()
	fmt.Println(`
Usage Examples:

  Checking a file's integrity checksum
    integrity myfile.jpg
    > myfile.jpg : sha1 : PASSED

  Adding integrity data to a file, skip if the file already has integrity dat
    integrity -a data_01.dat
    > data01.dat : sha1 : added

  Adding integrity data to a file, forcing a recalcuation if file already has integrity data
    integrity -a -f data01.dat
    > data01.dat : sha1 : added

  Checking the integrity of a file
    integrity -c data_01.dat
    > data01.dat : sha1 : PASSED
    > data01.dat : FAILED # Todo: Fix output to include sha1

  Checking the integrity of a file with integrity data verbosely
    integrity -v  -c  data01.dat
    > data01.dat : sha1 : 65bb1872af65ed02db42f603c786f5ec7d392909 : PASSED

    integrity -c -v data_01_corrupt.dat
    > Error checking checksum; data_01_corrupt.dat : Calculated checksum and filesystem read checksum differ!
      ├── xatr; [3fc98aa337e328816416e179afc863a75ffb330a]
      └── calc; [32c48f2bca002218e7488d5d41bb9c82743a3392]

  Listing integrity data
    integrity -l data_01.dat # Todo: add digest to normal output
    > data01.dat : 65bb1872af65ed02db42f603c786f5ec7d392909

  Listing integrity data verbosely
    integrity -l -v data_01.dat
    > data01.dat : sha1 : 65bb1872af65ed02db42f603c786f5ec7d392909

  Listing integrity data as shasum command output
    integrity -l --display-format=sha1sum  data01.dat
    > 65bb1872af65ed02db42f603c786f5ec7d392909 *data01.dat

  Listing integrity data as md5sum command output
    integrity -l --display-format=md5sum  data01.dat #ToDo block display of non-md5sum digests
    > 65bb1872af65ed02db42f603c786f5ec7d392909  data01.dat

  Recursively add integrity data to all files within a directory structure
    integrity -a -r ~/data/

  Recursively list the checksums as shasum output
    integrity -l -r ~/data/

Further Information:

  When copying files, extended attributes should be preserved to ensure
  integrity data is copied.

  For example:
    rsync -X source destination
    cp -p source destination

  The digest can be set through an environment variable I_DIGEST. This allows for a you to set your prefered digest
  method without needing to set it on the command line each time.

  For example:
    I_DIGEST='sha256' integrity -a file.dat

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
	     Check a file's default checksum (sha1)

	     integrity -a file.dat
	     Add a checksum using the default digest (sha1), displaying no output if the the action is sucessful, skipping
	     if the file already has one stored

	     integrity -d file.dat
	     Remove the default checksum (sha1) data, skipping if there is none stored

         integrity -l file.dat
	     Display a file's default checksum (sha1) data, skipping if there is none stored

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
  `)

}