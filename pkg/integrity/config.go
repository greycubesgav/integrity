package integrity

import (
	"crypto"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
	ShowHelp             bool
	ShowVersion          bool
	ShowInfo             bool
	Verbose              bool
	DigestHash           crypto.Hash
	DigestName           string
	Action               string
	DisplayFormat        string
	Action_Add           bool
	Action_Delete        bool
	Action_List          bool
	Action_Transform     bool
	Action_Check         bool
	Option_Force         bool
	Option_ShortPaths    bool
	Option_Recursive     bool
	Option_AllDigests    bool
	Option_DefaultDigest bool
	xattribute_fullname  string
	Xattribute_prefix    string
	logLevelName         string
	logObject            *logrus.Logger
	returnCode           int // used to store a return code for the cmd util
}

func newConfig() *Config {
	var c *Config = &Config{
		ShowHelp:             false,
		ShowVersion:          false,
		ShowInfo:             false,
		Action_Check:         false,
		Action_Add:           false,
		Action_Delete:        false,
		Action_List:          false,
		Action_Transform:     false,
		Option_Force:         false,
		Option_ShortPaths:    false,
		Option_Recursive:     false,
		Option_AllDigests:    false,
		Option_DefaultDigest: false,
		Verbose:              false,
		DigestHash:           crypto.SHA1,
		DigestName:           "",
		DisplayFormat:        "",
		Action:               "check",
		xattribute_fullname:  "",
		Xattribute_prefix:    "",
		logLevelName:         "info",
		logObject:            logrus.New(),
		returnCode:           0,
	}
	c.parseCmdlineOpt()
	return c
}

func (c *Config) parseCmdlineOpt() {

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

	//-----------------------------------------------------------------------------------------
	// Cover the help displays with exits first
	//-----------------------------------------------------------------------------------------
	// Show simple help, note return code '1' resvered for exit app but return '0' to terminal
	if c.ShowHelp {
		printHelp()
		c.returnCode = 1
		return
	}
	// Show the version of the app and exit
	if c.ShowVersion {
		fmt.Printf("%s\n", integrity_version)
		c.returnCode = 1
		return
	}

	//-----------------------------------------------------------------------------------------
	// Workout the digest we are using
	//-----------------------------------------------------------------------------------------
	if c.DigestName == "" {
		// Try and get the digest from the name of the command, this overrides any other setting
		// other than display formats sha1,md5 etc above
		// e.g. integrity.sha1, integrity.md5
		cmdName := filepath.Base(os.Args[0])
		cmdHash := strings.Split(cmdName, ".")
		if len(cmdHash) > 1 {
			c.DigestName = cmdHash[1]
		} else {
			// If we haven't been passed a digest name on the command line with --digest=
			// Try and get it from the environment variable
			envDigest := os.Getenv(env_name_prefix + "_DIGEST")
			if envDigest != "" {
				c.DigestName = envDigest
			} else {
				// If this doesn't work, set it to sha1 as default
				c.DigestName = "sha1"
				c.Option_DefaultDigest = true
			}
		}
	}
	// Check we know the digest type
	if c.DigestName != "oshash" && c.DigestName != "phash" {
		if c.DigestHash = digestTypes[c.DigestName]; c.DigestHash == 0 {
			fmt.Fprintf(os.Stderr, "Error : unknown hash type '%s'\n", c.DigestName)
			c.logObject.Errorf("Error : unknown hash type '%s'\n", c.DigestName)
			c.returnCode = 2
			return
		}
	}
	c.logObject.Debugf("c.DigestName: '%s'\n", c.DigestName)

	// Check the current OS and create the full xattribute name from the os, const and digest
	switch runtime.GOOS {
	case "darwin", "freebsd":
		c.xattribute_fullname = fmt.Sprintf("%s.%s", xattribute_name, c.DigestName)
		c.Xattribute_prefix = fmt.Sprintf("%s.", xattribute_name)
	case "linux":
		c.xattribute_fullname = fmt.Sprintf("user.%s.%s", xattribute_name, c.DigestName)
		c.Xattribute_prefix = fmt.Sprintf("user.%s.", xattribute_name)
	default:
		c.logObject.Fatalf("Error: non-supported OS type '%s'\n", runtime.GOOS)
		c.logObject.Fatalf("Supported OS types 'darwin, freebsd, linux'\n")
		c.returnCode = 3
		return
	}
	c.logObject.Debugf("c.xattribute_fullname: '%s'\n", c.xattribute_fullname)

	// Show internal info about the apps
	if c.ShowInfo {
		fmt.Printf("integrity version: %s\n", integrity_version)
		fmt.Printf("integrity attribute: %s\n", c.xattribute_fullname)
		fmt.Printf("runtime environment: %s\n", runtime.GOOS)
		c.returnCode = 1
		return
	}

	//	if c.DisplayFormat != "" && c.DisplayFormat != "sha1sum" && c.DisplayFormat != "md5sum" {
	//		fmt.Fprintf(os.Stderr, "Error : unknown display format '%s'\n Should be one of: sha1sum, md5sum\n", c.DisplayFormat)
	//		c.returnCode = 4
	//		return
	//	}

	//-----------------------------------------------------------------------------------------
	// Return error of no arguments are given
	//-----------------------------------------------------------------------------------------
	if getopt.NArgs() == 0 {
		fmt.Fprint(os.Stderr, "Error : no arguments given\n")
		getopt.Usage()
		c.returnCode = 5
		return
	}

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

	// Display actions list sha1 and md5, overwrite the action and digest type
	if c.DisplayFormat != "" {
		// If user has asked for display format, override action and digest type
		c.logObject.Debugf("c.DisplayFormat: '%s'\n", c.DisplayFormat)
		c.Action = "list"
		switch c.DisplayFormat {
		case "sha1sum":
			c.DigestName = "sha1"
		case "md5sum":
			c.DigestName = "md5"
		default:
			fmt.Fprintf(os.Stderr, "Error : unknown display format '%s'\n Should be one of: sha1sum, md5sum\n", c.DisplayFormat)
			c.returnCode = 4
			return
		}
	}

	c.logObject.Debugf("c.Action: '%s'\n", c.Action)

}

func printHelp() {
	fmt.Printf("integrity version %s\n", integrity_version)
	fmt.Printf("Web site: %s\n", integrity_website)
	fmt.Println(helpText)
	getopt.Usage()
	fmt.Println(usageText)
}
