package integrity

import (
	"crypto"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/pborman/getopt/v2"
	"github.com/sirupsen/logrus"
)

//go:embed help.txt
var helpText string

//go:embed usage.txt
var usageText string

const integrity_website = "https://github.com/greycubesgav/integrity"
const xattribute_name = "integrity"
const env_name_prefix = "INTEGRITY"

var digestTypes = map[string]crypto.Hash{
	"md5":         crypto.MD5,
	"sha1":        crypto.SHA1,
	"sha224":      crypto.SHA224,
	"sha256":      crypto.SHA256,
	"sha384":      crypto.SHA384,
	"sha512":      crypto.SHA512,
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
	ShowHelp            bool
	ShowVersion         bool
	ShowInfo            bool
	Verbose             bool
	Quiet               bool
	VerboseLevel        int
	DigestHash          crypto.Hash
	DigestName          string
	Action              string
	DisplayFormat       string
	Action_Add          bool
	Action_Delete       bool
	Action_List         bool
	Action_Transform    bool
	Action_Check        bool
	Option_Force        bool
	Option_ShortPaths   bool
	Option_Recursive    bool
	Option_AllDigests   bool
	xattribute_fullname string
	xattribute_prefix   string
	logLevelName        string
	logObject           *logrus.Logger
	returnCode          int // used to store a return code for the cmd util
	digestList          map[string]crypto.Hash
	digestNames         []string
	binaryDigestName    string
}

func newConfig() *Config {
	var c *Config = &Config{
		ShowHelp:            false,
		ShowVersion:         false,
		ShowInfo:            false,
		Action_Check:        false,
		Action_Add:          false,
		Action_Delete:       false,
		Action_List:         false,
		Action_Transform:    false,
		Option_Force:        false,
		Option_ShortPaths:   false,
		Option_Recursive:    false,
		Option_AllDigests:   false,
		Verbose:             false,
		Quiet:               false,
		VerboseLevel:        1,
		DigestHash:          crypto.SHA1,
		DigestName:          "",
		DisplayFormat:       "",
		Action:              "check",
		xattribute_fullname: "",
		xattribute_prefix:   "",
		logLevelName:        "info",
		logObject:           logrus.New(),
		returnCode:          0,
		digestList:          make(map[string]crypto.Hash),
		digestNames:         make([]string, 0),
		binaryDigestName:    "",
	}
	c.parseCmdlineOpt()
	return c
}

func (c *Config) parseCmdlineOpt() {

	var userDigestString string

	// Set the potential command line options
	getopt.FlagLong(&c.ShowHelp, "help", 'h', "show this help")
	getopt.FlagLong(&c.ShowVersion, "version", 0, "show version")
	getopt.FlagLong(&c.ShowInfo, "info", 0, "show information")
	getopt.FlagLong(&c.Action_Check, "check", 'c', "check the checksum of the file matches the one stored in the extended attributes [default]")
	getopt.FlagLong(&c.Action_Add, "add", 'a', "calculate the checksum of the file and add it to the extended attributes")
	getopt.FlagLong(&c.Action_Delete, "delete", 'd', "delete a checksum stored for a file")
	getopt.FlagLong(&c.Action_List, "list", 'l', "list the checksum stored for a file")
	getopt.FlagLong(&c.Action_Transform, "fix-old", 0, "fix an old extended attribute value name to the current format")
	getopt.FlagLong(&c.Option_AllDigests, "all", 'x', "include all digests, not just the default digest. Only applies to --delete and --list options")
	getopt.FlagLong(&c.Option_Force, "force", 'f', "force the calculation and writing of a checksum even if one already exists (default behaviour is to skip files with checksums already stored)")
	getopt.FlagLong(&c.Verbose, "verbose", 'v', "output more information.")
	getopt.FlagLong(&c.Quiet, "quiet", 'q', "output less information.")
	getopt.FlagLong(&c.logLevelName, "loglevel", 0, "set the logging level. One of: panic, fatal, error, warn, info, debug, trace.")
	getopt.FlagLong(&userDigestString, "digest", 0, "set the digest method(s) as a comma separated list (see help for list of digest types available)")
	getopt.FlagLong(&c.Option_ShortPaths, "short-paths", 's', "show only file name when showing file names, useful for generating sha1sum files")
	getopt.FlagLong(&c.Option_Recursive, "recursive", 'r', "recurse into sub-directories")
	getopt.FlagLong(&c.DisplayFormat, "display-format", 0, "set the output display format (sha1sum, md5sum). Note: this only shows any checkfiles ")
	getopt.Parse()

	//-----------------------------------------------------------------------------------------
	// Cover the help displays with exits first
	//-----------------------------------------------------------------------------------------
	// Show simple help, note return code '1' resvered for exit app but return '0' to terminal
	if c.ShowHelp {
		printHelp()
		c.returnCode = 1 // Show help
		return
	}
	// Show the version of the app and exit
	if c.ShowVersion {
		fmt.Printf("%s\n", integrity_version)
		c.returnCode = 1 // Show version
		return
	}

	//-----------------------------------------------------------------------------------------
	// Return error of no arguments are given
	//-----------------------------------------------------------------------------------------
	if getopt.NArgs() == 0 && !c.ShowInfo {
		fmt.Fprint(os.Stderr, "Error : no arguments given\n")
		getopt.Usage()
		c.returnCode = 2 // No arguments
		return
	}

	//-----------------------------------------------------------------------------------------
	// Setup the logging level
	//-----------------------------------------------------------------------------------------
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

	//-----------------------------------------------------------------------------------------
	// Main Actions
	//-----------------------------------------------------------------------------------------
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
	c.logObject.Debugf("c.Action: '%s'\n", c.Action)

	//-----------------------------------------------------------------------------------------
	// Workout the digest we are using
	// Heirarchy
	// 1. binary name , e.g. integriy.sha1
	// └ 2. command line all digest option
	//   └ 3. command line option, e.g. --digest=sha256,sha512
	//     └ 4. environment variable, e.g. INTEGRITY_DIGEST='md5,sha256'
	// Note: the binary name overwrites any other options
	//-----------------------------------------------------------------------------------------
	// Output of this block is a list of potential digestNames, will be validated later
	if cmdHash := strings.Split(filepath.Base(os.Args[0]), "."); len(cmdHash) == 2 {
		// Try and get the digest from the name of the binary, e.g integriy.sha1, integrity.md5
		// this overrides any other digest setting, other than display formats sha1sum, md5sum etc
		c.digestNames = []string{cmdHash[1]}
		c.binaryDigestName = cmdHash[1]
	} else if c.Option_AllDigests {
		// Otherwise, if we've been asked to perform against all digest types
		for digestName := range digestTypes {
			c.digestNames = append(c.digestNames, digestName)
		}
		// Add the two digests that don't come from crypto.Hash
		c.digestNames = append(c.digestNames, "oshash")
		c.digestNames = append(c.digestNames, "phash")
	} else {
		// If we've not been given a string from the user, try and get it from the environment
		if userDigestString == "" {
			userDigestString = os.Getenv(env_name_prefix + "_DIGEST")
		}
		userDigestArray := strings.Split(userDigestString, ",")
		if len(userDigestArray) == 1 && userDigestArray[0] == "" {
			// If all this fails to set the digest, we'll default to sha1
			c.digestNames = []string{"sha1"}
		} else {
			c.digestNames = append(c.digestNames, userDigestArray...)
		}
	}

	// Check if the display format doesn't make the digest
	if c.DisplayFormat != "" {
		c.logObject.Debugf("c.DisplayFormat: '%s'\n", c.DisplayFormat)
		// Either we override the action to be list, or we error that the action is not list
		// if c.Action != "list" {
		// 	fmt.Fprintf(os.Stderr, "Error : Display format provided but not performing list '%s'\n Should be one of: sha1sum, md5sum\n", c.DisplayFormat)
		// 	c.returnCode = 8 // Display format provided but action not list
		// 	return
		// }
		// Override the action to be list as we've been given a display format
		c.Action = "list"
		// Validate the display format
		switch c.DisplayFormat {
		case "sha1sum":
			if c.binaryDigestName != "" && c.binaryDigestName != "sha1" {
				fmt.Fprintf(os.Stderr, "Error : asked for sha1sum output but not sha1 binary.\n")
				c.returnCode = 6 // sha1sum output but not .md5 binary
				return
			}
			c.digestNames = []string{"sha1"}
		case "md5sum":
			if c.binaryDigestName != "" && c.binaryDigestName != "md5" {
				fmt.Fprintf(os.Stderr, "Error : asked for md5sum output but not md5 binary.\n")
				c.returnCode = 7 // md5sum output but not .md5 binary
				return
			}
			c.digestNames = []string{"md5"}
		case "cksum":
			// We will output any checksum in this case, no need to force the digest
		default:
			fmt.Fprintf(os.Stderr, "Error : unknown display format '%s'\n Should be one of: sha1sum, md5sum\n", c.DisplayFormat)
			c.returnCode = 4 // Unknown display format
			return
		}
	}
	c.logObject.Debugf("c.digestNames: '%s'\n", c.digestNames)

	//-----------------------------------------------------------------------------------------
	// Check we know all the given digest names
	//-----------------------------------------------------------------------------------------
	for _, digestName := range c.digestNames {
		if digestName != "oshash" && digestName != "phash" {
			if digest, exists := digestTypes[digestName]; exists {
				c.digestList[digestName] = digest
			} else {
				fmt.Fprintf(os.Stderr, "Error : unknown digest type '%s'\n", digestName)
				c.returnCode = 5 // Unknown digest
				return
			}
		}
	}
	// Sort the file list to aid printing
	sort.Strings(c.digestNames)

	// Check the current OS and create the full xattribute name from the os, const and digest
	switch runtime.GOOS {
	case "darwin", "freebsd":
		c.xattribute_prefix = fmt.Sprintf("%s.", xattribute_name)
	case "linux":
		c.xattribute_prefix = fmt.Sprintf("user.%s.", xattribute_name)
	default:
		c.logObject.Fatalf("Error: non-supported OS type '%s'\n", runtime.GOOS)
		c.logObject.Fatalf("Supported OS types 'darwin, freebsd, linux'\n")
		c.returnCode = 3 // Unknown OS
		return
	}
	c.logObject.Debugf("c.xattribute_prefix: '%s'\n", c.xattribute_prefix)

	// Show internal info about the apps
	if c.ShowInfo {
		fmt.Printf("integrity version: %s\n", integrity_version)
		fmt.Printf("integrity attribute prefix: %s\n", c.xattribute_prefix)
		fmt.Printf("runtime environment: %s\n", runtime.GOOS)
		fmt.Printf("digest list: %s\n", c.digestNames)
		fmt.Printf("integrity verbose level: %d\n", c.VerboseLevel)
		c.returnCode = 1 // Show info
		return
	}

	if c.Quiet {
		c.VerboseLevel = 0
	} else if c.Verbose {
		c.VerboseLevel = 2
	} else {
		c.VerboseLevel = 1
	}

}

func printHelp() {
	fmt.Printf("integrity version %s\n", integrity_version)
	fmt.Printf("Web site: %s\n", integrity_website)
	fmt.Println(helpText)
	getopt.Usage()
	fmt.Println(usageText)
}
