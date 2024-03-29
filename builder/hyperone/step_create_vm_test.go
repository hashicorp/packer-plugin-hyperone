// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hyperone

import (
	"testing"

	openapi "github.com/hyperonecom/h1-client-go"
	"github.com/stretchr/testify/assert"
)

func TestPickNetAdapter(t *testing.T) {
	cases := []struct {
		Name     string
		Config   Config
		Expected openapi.ComputeProjectVmCreateNetadp
	}{
		{
			Name: "no_network_public_ip",
			Config: Config{
				PublicIP: "some-public-ip",
			},
			Expected: openapi.ComputeProjectVmCreateNetadp{
				Ip: []string{"some-public-ip"},
			},
		},
		{
			Name: "no_network_private_ip",
			Config: Config{
				PrivateIP: "some-private-ip",
			},
			Expected: openapi.ComputeProjectVmCreateNetadp{},
		},
		{
			Name: "no_network_both_ip",
			Config: Config{
				PublicIP:  "some-public-ip",
				PrivateIP: "some-private-ip",
			},
			Expected: openapi.ComputeProjectVmCreateNetadp{
				Ip: []string{"some-public-ip"},
			},
		},
		{
			Name: "network_no_ip",
			Config: Config{
				Network: "some-network",
			},
			Expected: openapi.ComputeProjectVmCreateNetadp{
				Network: "some-network",
			},
		},
		{
			Name: "network_public_ip",
			Config: Config{
				Network:  "some-network",
				PublicIP: "some-public-ip",
			},
			Expected: openapi.ComputeProjectVmCreateNetadp{
				Network: "some-network",
			},
		},
		{
			Name: "network_private_ip",
			Config: Config{
				Network:   "some-network",
				PrivateIP: "some-private-ip",
			},
			Expected: openapi.ComputeProjectVmCreateNetadp{
				Network: "some-network",
				Ip:      []string{"some-private-ip"},
			},
		},
		{
			Name: "network_both_ip",
			Config: Config{
				Network:   "some-network",
				PublicIP:  "some-public-ip",
				PrivateIP: "some-private-ip",
			},
			Expected: openapi.ComputeProjectVmCreateNetadp{
				Network: "some-network",
				Ip:      []string{"some-private-ip"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			result := pickNetAdapter(&c.Config)
			assert.Equal(t, c.Expected, result)
		})
	}
}
