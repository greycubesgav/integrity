package integrity_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/greycubesgav/integrity/pkg/integrity"
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
		Setup: func(env *testscript.Env) error {
			// Copy necessary binaries to the test environment
			err := exec.Command("cp", "-rf", "testdata/imgs/_MG_5859.JPG", env.WorkDir).Run()
			if err != nil {
				return err
			}
			err = exec.Command("cp", "-rf", "testdata/imgs/_MG_5860.heic", env.WorkDir).Run()
			if err != nil {
				return err
			}
			err = exec.Command("cp", "-rf", "testdata/imgs/_MG_5861.png", env.WorkDir).Run()
			if err != nil {
				return err
			}
			err = exec.Command("cp", "-rf", "testdata/imgs/_MG_5862.tiff", env.WorkDir).Run()
			if err != nil {
				return err
			}
			return nil
		},
	})
}
