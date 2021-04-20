package hyperone

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

func TestAccBuilder_basic(t *testing.T) {
	testCase := &acctest.PluginTestCase{
		Name:     "hyperone_basic_test",
		Setup: func() error {
			if v := os.Getenv("HYPERONE_TOKEN"); v == "" {
				return fmt.Errorf("HYPERONE_TOKEN must be set for acceptance tests")
			}
			return nil
		},
		Template: testBuilderAccBasic,
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
	//builderT.Test(t, builderT.TestCase{
	//	PreCheck: func() { testAccPreCheck(t) },
	//	Builder:  &Builder{},
	//	Template: testBuilderAccChroot,
	//})
	testCase := &acctest.PluginTestCase{
		Name:     "hyperone_chroot_test",
		Setup: func() error {
			if v := os.Getenv("HYPERONE_TOKEN"); v == "" {
				return fmt.Errorf("HYPERONE_TOKEN must be set for acceptance tests")
			}
			return nil
		},
		Template: testBuilderAccChroot,
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

const testBuilderAccBasic = `
{
	"builders": [{
		"type": "hyperone",
		"vm_type": "a1.nano",
		"source_image": "ubuntu",
		"disk_size": 10,
		"image_tags": {
			"key":"value"
		},
		"vm_tags": {
			"key_vm":"value_vm"
		}
	}]
}
`

const testBuilderAccChroot = `
{
	"builders": [{
		"type": "hyperone",
		"source_image": "ubuntu",
		"disk_size": 10,
		"vm_type": "a1.nano",
		"chroot_disk": true,
		"chroot_command_wrapper": "sudo {{.Command}}",
		"pre_mount_commands": [
			"parted {{.Device}} mklabel msdos mkpart primary 1M 100% set 1 boot on print",
			"mkfs.ext4 {{.Device}}1"
		],
		"post_mount_commands": [
			"apt-get update",
			"apt-get install debootstrap",
			"debootstrap --arch amd64 bionic {{.MountPath}}"
		]
	}]
}
`
