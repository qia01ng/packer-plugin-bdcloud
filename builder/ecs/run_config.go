//go:generate packer-sdc struct-markdown

package ecs

import (
	"errors"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
)

type RunConfig struct {
	AvailableZone string `mapstructure:"available_zone" required:"true"`
	InstanceSpec  string `mapstructure:"instance_spec" required:"true"`
	RootDiskType  string `mapstructure:"root_disk_type" required:"true"`
	RootDiskSize  int    `mapstructure:"root_disk_size" required:"true"`
	MountBns      string `mapstructure:"mount_bns" required:"true"`

	// Communicator settings
	Comm communicator.Config `mapstructure:",squash"`
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	if c.Comm.SSHKeyPairName == "" && c.Comm.SSHTemporaryKeyPairName == "" &&
		c.Comm.SSHPrivateKeyFile == "" && c.Comm.SSHPassword == "" && c.Comm.WinRMPassword == "" {

		c.Comm.SSHTemporaryKeyPairName = fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
	}

	// Validation
	errs := c.Comm.Prepare(ctx)
	if c.InstanceSpec == "" {
		errs = append(errs, errors.New("An instance_spec must be specified"))
	}

	return errs
}
