package hyperone

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	openapi "github.com/hyperonecom/h1-client-go"
)

type stepCreateVMFromDisk struct {
	vmID string
}

func (s *stepCreateVMFromDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*openapi.APIClient)
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)
	sshKey := state.Get("ssh_public_key").(string)
	chrootDiskUri := state.Get("chroot_disk_uri").(*string)

	ui.Say("Creating VM from disk...")

	options := openapi.ComputeProjectVmCreate{
		Name:    config.VmName,
		Service: config.VmType,
		Credential: []openapi.ComputeProjectVmCreateCredential{
			{
				Type:  "ssh",
				Value: sshKey,
			},
		},
	}
	options.SetStart(false)

	refreshToken(state) //TODO move to h1-client-go
	vm, _, err := client.
		ComputeProjectVmApi.
		ComputeProjectVmCreate(ctx, config.Project, config.Location).
		ComputeProjectVmCreate(options).
		Execute()

	if err != nil {
		err := fmt.Errorf("error creating VM from disk: %s", formatOpenAPIError(err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	refreshToken(state) //TODO move to h1-client-go
	_, _, err = client.
		ComputeProjectVmApi.
		ComputeProjectVmDiskCreate(ctx, config.Project, config.Location, vm.Id).
		ComputeProjectVmDiskCreate(openapi.ComputeProjectVmDiskCreate{Disk: *chrootDiskUri}).
		Execute()

	if err != nil {
		err := fmt.Errorf("error creating VM from disk, attaching: %s", formatOpenAPIError(err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.vmID = vm.Id
	state.Put("vm_id", vm.Id)
	state.Put("vm_uri", vm.Uri)

	return multistep.ActionContinue
}

func (s *stepCreateVMFromDisk) Cleanup(state multistep.StateBag) {
	if s.vmID == "" {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)

	ui.Say(fmt.Sprintf("Deleting VM %s (from chroot disk)...", s.vmID))
	err := deleteVMWithDisks(context.Background(), state, s.vmID)
	if err != nil {
		ui.Error(err.Error())
	}
}
