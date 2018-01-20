package main

import (
	"flag"
	"fmt"
	"os"
	"crypto"
)


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
	Action_Add				bool
	Action_Delete			bool
	Action_List				bool
	Option_List_sha1sum		bool
	Option_List_md5sum		bool
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
		Option_List_sha1sum:	false,
		Option_List_md5sum:		false,
		Option_Force:			false,
		Option_ShortPaths:		false,
		Option_Recursive:		false,
		Option_AllDigests:		false,
		Option_DefaultDigest:	false,
		//TestChecksum:			false,
		Verbose:				false,
		DigestHash:				crypto.SHA1,
		DigestName:				"",
		Action:					"check",
	}
	c.ParseCmdlineOpt()
	return c
}

func (c *Config) ParseCmdlineOpt() {

	flag.BoolVar(&c.ShowHelp,"h",    c.ShowHelp, "show this help")
	flag.BoolVar(&c.ShowHelp,"help", c.ShowHelp, "show this help")

	flag.BoolVar(&c.Action_Check, "c", 	    c.Action_Check,	"check the checksum of the file matches the one stored in the extended attributes")
	flag.BoolVar(&c.Action_Check, "check", 	c.Action_Check, "check the checksum of the file matches the one stored in the extended attributes")

	flag.BoolVar(&c.Action_Add, "a",   c.Action_Add, "calculate the checksum of the file and add it to the extended attributes")
	flag.BoolVar(&c.Action_Add, "add", c.Action_Add, "calculate the checksum of the file and add it to the extended attributes")

	flag.BoolVar(&c.Action_Delete, "d",      c.Action_Delete, "delete a stored checksum")
	flag.BoolVar(&c.Action_Delete, "delete", c.Action_Delete, "delete a stored checksum")

	flag.BoolVar(&c.Action_List, "l",    c.Action_List, "list the checksum stored for a file")
	flag.BoolVar(&c.Action_List, "list", c.Action_List, "list the checksum stored for a file")

	flag.BoolVar(&c.Option_List_sha1sum, "sha1sum", c.Option_List_sha1sum, "list the checksum stored for a file as per sha1sum, not this does not exclude the use of other digest formats!")

	flag.BoolVar(&c.Option_List_md5sum, "md5sum", c.Option_List_md5sum, "list the checksum stored for a file as per md5sum, not this does not exclude the use of other digest formats!")

	flag.BoolVar(&c.Option_AllDigests, "x",   c.Option_Force,"include all digests, not just the default digest. Only applies to -d --delete and -l --list options")
	flag.BoolVar(&c.Option_AllDigests, "all", c.Option_Force,"include all digests, not just the default digest. Only applies to -d --delete and -l --list options")

	flag.BoolVar(&c.Option_Force, "f",     c.Option_Force,"force the writing of a checksum even if it already exists (default behaviour is to skip files with checksums already stored")
	flag.BoolVar(&c.Option_Force, "force", c.Option_Force,"force the writing of a checksum even if it already exists (default behaviour is to skip files with checksums already stored")

	flag.BoolVar(&c.Verbose, "v",       c.Verbose, "print more verbose messages")
	flag.BoolVar(&c.Verbose, "verbose", c.Verbose, "print more verbose messages")

	flag.StringVar(&c.DigestName, "digest", c.DigestName, "set the digest method (see help for list of available options")

	flag.BoolVar(&c.Option_ShortPaths, "s",           c.Option_ShortPaths,"show only file name when showing file names, useful for generating sha1sum files")
	flag.BoolVar(&c.Option_ShortPaths, "short-paths", c.Option_ShortPaths,"show only file name when showing file names, useful for generating sha1sum files")

	flag.BoolVar(&c.Option_Recursive, "r",         c.Option_Recursive,"recurse into directories")
	flag.BoolVar(&c.Option_Recursive, "recursive", c.Option_Recursive,"recurse into directories")

	//flag.BoolVar(&c.TestChecksum, "t", c.TestChecksum, "When adding a new checksum test if the existing checksum 'looks' correct, skip if it does")

	flag.Parse()

	if flag.NArg() == 0 || c.ShowHelp {
		printHelp()
		if flag.NFlag() == 0 {
			fmt.Fprint(os.Stderr, "Error : no arguments given\n")
			os.Exit(1)
		} else {
			os.Exit(0)
		}
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

	if c.Action_Check {
		c.Action = "check"
	} else if c.Action_Delete {
		c.Action = "delete"
	} else if c.Action_Add {
		c.Action = "add"
	} else if c.Action_List {
		c.Action = "list"
	}
}

func printHelp() {
	fmt.Printf("Usage: integrity [OPTIONS] FILE|PATH\n")
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println(`Examples:
  Check a files integrity checksum
   integrity myfile.jpg

  Adding integrity data to a file
    integrity -a data_01.dat

  Add integrity data to a file, skip if the file already has integrity data
    integrity -a -s data_01.dat

  Checking the integrity of a file with integrity data
    integrity -c data_01.dat

  Checking the integrity of a file with integrity data verbosely
    integrity -c -v data_01.dat

  Listing integrity data as shasum command output, with full filepath
    integrity -l data_01.dat

  Listing integrity data as shasum command output, with only filename
    integrity -m data_01.dat

  Using shasum to check the integrity of a list of files (osx)
    integrity -l data_01.dat | shasum -c

  Recursively add integrity data to all files within a directory structure
	linux:
    find directory -type f -print0  | xargs -0 integrity -a

    osx:
    find directory -type f -not -name '.DS_Store' -print0 | xargs -0 integrity -a

  Recursively list the checksums as shasum output (osx)
    find directory -type f -print0  | xargs -0 integrity -l

  Locate duplicate files within a directory structure (osx)
    integrity_dupes directory

  Transfering a file to a remote machine maintaining integrity metadata
    rsync -X data_01.dat remote_server:/destination/

Info:
  When copying files, extended attributes should be preserved to ensure
  integrity data is copied.
  e.g.
	rsync -X source destination
    cp -p source destination

  The digest can be set through an environment variable I_DIGEST. This allows for a you to set your prefered digest
  method without needing to set it on the command line each time.

  For example:
	I_DIGEST='sha256' integrity -a file.dat


  Design Choices

    * By default this util is meant to be quite quiet. i.e. when adding trying to add a checksum to a file with one
	  stored already, the app will simply skip over the file and continue. This is because the util is meant to be ran
	  over large numbers of data files which may or may not already have checksum data so output is kept to a minimum.

	  For example:

	     integrity -a -r directory/
	     Add the default checksum data too all files, 'added' will be shown for all files which had checksum data added.
		 Nothing will be shown for the others.

      Add the -v flag to see more verbose output.

	* The util is designed to do "sensible" things with basic options

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

  Supported Checksum Digests

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