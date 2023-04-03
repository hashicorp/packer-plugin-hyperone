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

func convertTags(list map[string]string) []openapi.Tag {
	result := make([]openapi.Tag, len(list))

	index := 0
	for key, value := range list {
		result[index] = openapi.Tag{Key: key, Value: value}
		index++
	}
	return result
}

type stepCreateVM struct {
	vmID string
}

const (
	chrootDiskName = "packer-chroot-disk"
)

func (s *stepCreateVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*openapi.APIClient)
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)
	sshKey := state.Get("ssh_public_key").(string)

	ui.Say("Creating VM...")

	netAdapter := pickNetAdapter(config)

	sshKeys := append([]string{sshKey}, config.SSHKeys...)

	credential := []openapi.ComputeProjectVmCreateCredential{}

	for _, key := range sshKeys {
		credential = append(credential, openapi.ComputeProjectVmCreateCredential{
			Type:  "ssh",
			Value: key,
		})
	}

	disks := []openapi.ComputeProjectVmCreateDisk{
		{
			Name:    config.DiskName,
			Service: config.DiskType,
			Size:    config.DiskSize,
		},
	}

	if config.ChrootDisk {
		disks = append(disks, openapi.ComputeProjectVmCreateDisk{
			Service: config.ChrootDiskType,
			Size:    config.ChrootDiskSize,
			Name:    chrootDiskName,
		})
	}

	options := openapi.ComputeProjectVmCreate{
		Name:         config.VmName,
		Service:      config.VmType,
		Image:        &config.SourceImage,
		Credential:   credential,
		Disk:         disks,
		Netadp:       []openapi.ComputeProjectVmCreateNetadp{netAdapter},
		UserMetadata: &config.UserData,
		Tag:          convertTags(config.VmTags),
		Username:     &config.Comm.SSHUsername,
	}

	refreshToken(state) //TODO move to h1-client-go
	vm, _, err := client.
		ComputeProjectVmApi.
		ComputeProjectVmCreate(ctx, config.Project, config.Location).
		ComputeProjectVmCreate(options).
		Execute()

	if err != nil {
		err := fmt.Errorf("error creating VM: %s", formatOpenAPIError(err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.vmID = vm.Id
	state.Put("vm_id", vm.Id)
	state.Put("vm_uri", vm.Uri)
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", vm.Id)

	refreshToken(state) //TODO move to h1-client-go
	disks2, _, err := client.
		ComputeProjectVmApi.
		ComputeProjectVmDiskList(ctx, config.Project, config.Location, vm.Id).
		Execute()

	if err != nil {
		err := fmt.Errorf("error listing disks: %s", formatOpenAPIError(err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	for _, disk := range disks2 {
		if disk.Name == chrootDiskName {
			state.Put("chroot_disk_id", disk.Id)
			state.Put("chroot_disk_uri", disk.Uri)
			break
		}
	}

	refreshToken(state) //TODO move to h1-client-go
	netadp, _, err := client.
		NetworkingProjectNetadpApi.
		NetworkingProjectNetadpList(ctx, config.Project, config.Location).
		AssignedId(vm.Id).
		Execute()

	if err != nil {
		err := fmt.Errorf("error listing netadp: %s", formatOpenAPIError(err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(netadp) < 1 {
		err := fmt.Errorf("no network adapters found")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	refreshToken(state) //TODO move to h1-client-go
	publicIP, err := associatePublicIP(ctx, config, client, netadp[0])
	if err != nil {
		err := fmt.Errorf("error associating IP: %s", formatOpenAPIError(err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("public_ip", publicIP)

	return multistep.ActionContinue
}

func pickNetAdapter(config *Config) openapi.ComputeProjectVmCreateNetadp {
	ret := openapi.ComputeProjectVmCreateNetadp{}

	if config.Network == "" {
		if config.PublicIP != "" {
			ret.Ip = []string{config.PublicIP}
		}
	} else {
		if config.PrivateIP != "" {
			ret.Ip = []string{config.PrivateIP}
		}

		ret.Network = config.Network
	}

	return ret
}

func associatePublicIP(ctx context.Context, config *Config, client *openapi.APIClient, netadp openapi.Netadp) (string, error) {

	ips, _, err := client.
		NetworkingProjectIpApi.
		NetworkingProjectIpList(ctx, config.Project, config.Location).
		AssociatedNetadp(netadp.Id).
		Execute()

	if err != nil {
		return "", err
	}

	if config.Network == "" || config.PublicIP == "" {
		// Public IP belongs to attached net adapter
		return *ips[0].Address, nil
	}

	var privateIP string
	if config.PrivateIP == "" {
		privateIP = ips[0].Id
	} else {
		privateIP = config.PrivateIP
	}

	ip, _, err := client.
		NetworkingProjectIpApi.
		NetworkingProjectIpAssociate(ctx, config.Project, config.Location, config.PublicIP).
		NetworkingProjectIpAssociate(*openapi.NewNetworkingProjectIpAssociate(privateIP)).
		Execute()

	if err != nil {
		return "", err
	}

	return *ip.Address, nil
}

func (s *stepCreateVM) Cleanup(state multistep.StateBag) {
	if s.vmID == "" {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)

	ui.Say(fmt.Sprintf("Deleting VM %s...", s.vmID))
	err := deleteVMWithDisks(context.Background(), state, s.vmID)
	if err != nil {
		ui.Error(err.Error())
	}
}

func deleteVMWithDisks(ctx context.Context, state multistep.StateBag, vmID string) error {
	client := state.Get("client").(*openapi.APIClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	refreshToken(state) //TODO move to h1-client-go
	disks, _, err := client.
		ComputeProjectVmApi.
		ComputeProjectVmDiskList(ctx, config.Project, config.Location, vmID).
		Execute()

	if err != nil {
		return fmt.Errorf("error listing disks: %s", formatOpenAPIError(err))
	}

	refreshToken(state) //TODO move to h1-client-go
	_, _, err = client.
		ComputeProjectVmApi.
		ComputeProjectVmDelete(ctx, config.Project, config.Location, vmID).
		Execute()

	if err != nil {
		return fmt.Errorf("error deleting server '%s' - please delete it manually: %s", vmID, formatOpenAPIError(err))
	}

	for _, disk := range disks {
		ui.Say(fmt.Sprintf("Deleting Disk %s...", disk.Id))
		refreshToken(state) //TODO move to h1-client-go
		_, _, err = client.
			StorageProjectDiskApi.
			StorageProjectDiskDelete(ctx, config.Project, config.Location, disk.Id).
			Execute()

		if err != nil {
			return fmt.Errorf("error deleting disk '%s' - please delete it manually: %s", disk.Id, formatOpenAPIError(err))
		}
	}

	return nil
}
