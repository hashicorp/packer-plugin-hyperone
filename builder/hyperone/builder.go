// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hyperone

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	openapi "github.com/hyperonecom/h1-client-go"
	credentials "github.com/hyperonecom/h1-credentials-helper-go"
	"github.com/hyperonecom/h1-credentials-helper-go/providers"
)

const BuilderID = "hyperone.builder"

type Builder struct {
	config   Config
	runner   multistep.Runner
	client   *openapi.APIClient
	provider providers.TokenAuthProvider
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	warnings, errs := b.config.Prepare(raws...)
	if errs != nil {
		return nil, warnings, errs
	}

	cfg := openapi.NewConfiguration()

	var err error

	b.provider, err = credentials.GetPassportCredentialsHelper("") // empty string means that the library should look for passport file in ~/.h1/passport.json
	// if you have this file in different location you can pass it to this function

	if err != nil {
		return nil, nil, err
	}

	if b.config.APIURL != "" {
		cfg.Servers[0].URL = b.config.APIURL
	}

	prefer := fmt.Sprintf("respond-async,wait=%d", int(b.config.StateTimeout.Seconds()))
	cfg.AddDefaultHeader("Prefer", prefer)

	b.client = openapi.NewAPIClient(cfg)

	return nil, nil, nil
}

type wrappedCommandTemplate struct {
	Command string
}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
	wrappedCommand := func(command string) (string, error) {
		ictx := b.config.ctx
		ictx.Data = &wrappedCommandTemplate{Command: command}
		return interpolate.Render(b.config.ChrootCommandWrapper, &ictx)
	}

	state := &multistep.BasicStateBag{}
	state.Put("provider", b.provider)
	state.Put("config", &b.config)
	state.Put("client", b.client)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("wrappedCommand", CommandWrapper(wrappedCommand))

	steps := []multistep.Step{
		&stepCreateSSHKey{},
		&stepCreateVM{},
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      getPublicIP,
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
	}

	if b.config.ChrootDisk {
		steps = append(steps,
			&stepPrepareDevice{},
			&stepPreMountCommands{},
			&stepMountChroot{},
			&stepPostMountCommands{},
			&stepMountExtra{},
			&stepCopyFiles{},
			&stepChrootProvision{},
			&stepStopVM{},
			&stepDetachDisk{},
			&stepCreateVMFromDisk{},
			&stepCreateImage{},
		)
	} else {
		steps = append(steps,
			&commonsteps.StepProvision{},
			&stepStopVM{},
			&stepCreateImage{},
		)
	}

	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	artifact := &Artifact{
		imageID:   state.Get("image_id").(string),
		imageName: state.Get("image_name").(string),
		state:     state,
		StateData: map[string]interface{}{"generated_data": state.Get("generated_data")},
	}

	return artifact, nil
}
