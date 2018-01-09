package main

import (
	"flag"
	"fmt"
	"os"
	"crypto"
)

type Config struct {
	ShowHelp			bool
	TestChecksum		bool
	SkipFiles			bool
	Verbose				bool
	DigestHash			crypto.Hash
	DigestName			string
	Action				string
	Action_Write		bool
	Action_Delete		bool
	Action_List			bool
	Action_List_Trim	bool
	Action_Check		bool
}

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

func NewConfig() *Config {
	var c *Config = &Config{
		ShowHelp:				false,
		Action_Check:			false,
		Action_Write:			false,
		Action_Delete:			false,
		Action_List:			false,
		Action_List_Trim:		false,
		TestChecksum:			false,
		SkipFiles:				true,
		Verbose:				false,
		DigestHash:			    crypto.SHA1,
		DigestName:				"sha1",
		Action:					"check",
	}
	c.ParseCmdlineOpt()
	return c
}

func (c *Config) ParseCmdlineOpt() {

	flag.BoolVar(&c.ShowHelp,        "h", c.ShowHelp,        "Show help")
	flag.BoolVar(&c.Action_Check,    "c", c.Action_Check,    "Check the checksum of FILE")
	flag.BoolVar(&c.Action_Write,    "a", c.Action_Write,    "Add a new checksum to FILE")
	flag.BoolVar(&c.Action_Delete,   "d", c.Action_Delete,   "Remove the checksum from FILE")
	flag.BoolVar(&c.Action_List,     "l", c.Action_List,     "List files, with path, along with checksums as per a shasum output")
	flag.BoolVar(&c.Action_List_Trim,"f", c.Action_List_Trim,"List files, without paths, along with checksums as per a shasum output,")
	flag.BoolVar(&c.TestChecksum,    "t", c.TestChecksum,    "When adding a new checksum test if the existing checksum 'looks' correct, skip if it does")
	flag.BoolVar(&c.SkipFiles,       "s", c.SkipFiles,       "When adding new checksums skip if the file already has checksum data")
	flag.BoolVar(&c.Verbose,         "v", c.TestChecksum,    "Verbose messages")
	flag.StringVar(&c.DigestName, "digest", c.DigestName, "Set the digest method")
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

	if c.DigestHash = digestTypes[c.DigestName]; c.DigestHash == 0 {
		fmt.Fprintf(os.Stderr, "\nError : unknown hash type '%s'\n", c.DigestName)
		os.Exit(2)
	}

	if c.Action_Check {
		c.Action = "check"
	} else if c.Action_Delete {
		c.Action = "delete"
	} else if c.Action_Write {
		c.Action = "write"
	} else if c.Action_List {
		c.Action = "list"
	} else if c.Action_List_Trim {
		c.Action = "list_trim"
	}
}

func printHelp() {
	fmt.Printf("Usage: integrity [OPTIONS] FILE|PATH\n")
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println(`Examples:
  Check a files integrity checksum
    ${0##*/} myfile.jpg

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
  e.g. rsync -X source destination
       osx : cp -p source destination

  This script assumes opensll is available in your path.`)

}