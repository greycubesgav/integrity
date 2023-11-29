package integrity_test

import (
	"os"
	"testing"

	"github.com/greycubesgav/integrity"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"integrity": integrity.Run,
	}))
}

func TestIntegrity(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}
