package ecs

import (
	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

type alicloudSSHHelper interface {
}

// SSHHost returns a function that can be given to the SSH communicator
func SSHHost(e alicloudSSHHelper) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		ipAddress := state.Get("instance_ip").(string)
		return ipAddress, nil
	}
}
