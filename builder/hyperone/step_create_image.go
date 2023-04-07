// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hyperone

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	openapi "github.com/hyperonecom/h1-client-go"
)

type stepCreateImage struct {
	imageID string
}

func (s *stepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*openapi.APIClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm_uri").(*string)

	ui.Say("Creating image...")

	options := openapi.StorageProjectImageCreate{
		Name:        config.ImageName,
		Vm:          vm,
		Service:     &config.ImageService,
		Description: &config.ImageDescription,
		Tag:         convertTags(config.ImageTags),
	}

	refreshToken(state) //TODO move to h1-client-go
	image, _, err := client.
		StorageProjectImageApi.
		StorageProjectImageCreate(ctx, config.Project, config.Location).
		StorageProjectImageCreate(options).
		Execute()

	if err != nil {
		err := fmt.Errorf("error creating image: %s", formatOpenAPIError(err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.imageID = image.Id

	state.Put("image_id", image.Id)
	state.Put("image_name", image.Name)

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {
	if s.imageID == "" {
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	client := state.Get("client").(*openapi.APIClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	refreshToken(state) //TODO move to h1-client-go
	_, _, err := client.
		StorageProjectImageApi.
		StorageProjectImageDelete(context.TODO(), config.Project, config.Location, s.imageID).
		Execute()

	if err != nil {
		ui.Error(fmt.Sprintf("error deleting image '%s' - consider deleting it manually: %s",
			s.imageID, formatOpenAPIError(err)))
	}
}
