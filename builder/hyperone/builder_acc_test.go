package hyperone

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

func TestAccBuilder_basic(t *testing.T) {

	template, _ := os.ReadFile("./examples/basic.pkr.hcl")

	testCase := &acctest.PluginTestCase{
		Name: "hyperone_basic_test",
		Setup: func() error {
			if v := os.Getenv("HYPERONE_PASSPORT_FILE"); v == "" {
				return fmt.Errorf("HYPERONE_PASSPORT_FILE must be set for acceptance tests")
			}

			if v := os.Getenv("HYPERONE_PROJECT"); v == "" {
				return fmt.Errorf("HYPERONE_PROJECT must be set for acceptance tests")
			}

			return nil
		},
		Template: string(template),
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

func TestBuilderAcc_chroot(t *testing.T) {
	template, _ := os.ReadFile("./examples/chroot.pkr.hcl")

	testCase := &acctest.PluginTestCase{
		Name: "hyperone_chroot_test",
		Setup: func() error {
			if v := os.Getenv("HYPERONE_PASSPORT_FILE"); v == "" {
				return fmt.Errorf("HYPERONE_PASSPORT_FILE must be set for acceptance tests")
			}
			return nil
		},
		Template: string(template),
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}
