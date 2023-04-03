// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hyperone

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepPrepareDevice struct{}

func (s *stepPrepareDevice) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)

	if config.ChrootDevice != "" {
		state.Put("device", config.ChrootDevice)
		return multistep.ActionContinue
	}

	// controllerNumber := state.Get("chroot_controller_number").(string)
	// controllerLocation := state.Get("chroot_controller_location").(int)
	chrootDiskID := state.Get("chroot_disk_id").(string)

	log.Println("Searching for available device...")

	cmd := fmt.Sprintf("readlink -f /dev/disk/by-id/scsi-*%s | uniq", chrootDiskID[12:])

	device, err := captureOutput(cmd, state)
	if err != nil {
		err := fmt.Errorf("error finding available device: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if device == "" {
		err := fmt.Errorf("device not found")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if strings.Contains(device, "\n") {
		err := fmt.Errorf("FIXME: multiple devices found")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Found device: %s", device))
	state.Put("device", device)
	return multistep.ActionContinue
}

func (s *stepPrepareDevice) Cleanup(state multistep.StateBag) {}
