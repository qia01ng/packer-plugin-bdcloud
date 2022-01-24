//go:generate packer-sdc mapstructure-to-hcl2 -type Config

// The alicloud  contains a packersdk.Builder implementation that
// builds ecs images for alicloud.
package ecs

import (
	"context"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// The unique ID for this builder
const BuilderId = "baidu.bdcloud"

type Config struct {
	common.PackerConfig  `mapstructure:",squash"`
	AlicloudAccessConfig `mapstructure:",squash"`
	AlicloudImageConfig  `mapstructure:",squash"`
	RunConfig            `mapstructure:",squash"`

	ctx interpolate.Context
}

type Builder struct {
	config Config
	runner multistep.Runner
}

type InstanceNetWork string

const (
	ALICLOUD_DEFAULT_SHORT_TIMEOUT = 180
	ALICLOUD_DEFAULT_TIMEOUT       = 1800
	ALICLOUD_DEFAULT_LONG_TIMEOUT  = 3600
)

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	err := config.Decode(&b.config, &config.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)
	b.config.ctx.EnableEnv = true
	if err != nil {
		return nil, nil, err
	}

	// Accumulate any errors
	var errs *packersdk.MultiError
	errs = packersdk.MultiErrorAppend(errs, b.config.AlicloudAccessConfig.Prepare(&b.config.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, b.config.AlicloudImageConfig.Prepare(&b.config.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	packersdk.LogSecretFilter.Set(b.config.BceAccessKey, b.config.BceSecretKey)
	return nil, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {

	client, err := b.config.Client()
	if err != nil {
		return nil, err
	}

	rsClient, err := b.config.RsClient()
	if err != nil {
		return nil, err
	}
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("client", client)
	state.Put("rsClient", rsClient)
	state.Put("hook", hook)
	state.Put("ui", ui)
	var steps []multistep.Step

	// // Build the steps
	// steps = []multistep.Step{
	// 	&stepPreValidate{
	// 		AlicloudDestImageName: b.config.AlicloudImageName,
	// 		ForceDelete:           b.config.AlicloudImageForceDelete,
	// 	},
	// }

	// // Customer config image family skip image check
	// if b.config.AlicloudImageFamily != "" {
	// 	steps = append(steps,
	// 		&stepCheckAlicloudImageFamily{
	// 			ImageFamily: b.config.AlicloudImageFamily,
	// 		})
	// } else {
	// get source image id
	steps = append(steps,
		&stepCheckAlicloudSourceImage{
			SrcImageId:   b.config.SrcImageId,
			SrcImageType: b.config.SrcImageType,
			SrcImageName: b.config.SrcImageName,
		})
	// steps = append(steps,
	// 	&stepConfigAlicloudKeyPair{
	// 		Debug:        b.config.PackerDebug,
	// 		Comm:         &b.config.Comm,
	// 		DebugKeyPath: fmt.Sprintf("ecs_%s.pem", b.config.PackerBuildName),
	// 		RegionId:     b.config.AlicloudRegion,
	// 	})
	// if b.chooseNetworkType() == InstanceNetworkVpc {
	// 	steps = append(steps,
	// 		&stepConfigAlicloudVPC{
	// 			VpcId:     b.config.VpcId,
	// 			CidrBlock: b.config.CidrBlock,
	// 			VpcName:   b.config.VpcName,
	// 		},
	// 		&stepConfigAlicloudVSwitch{
	// 			VSwitchId:   b.config.VSwitchId,
	// 			ZoneId:      b.config.ZoneId,
	// 			CidrBlock:   b.config.CidrBlock,
	// 			VSwitchName: b.config.VSwitchName,
	// 		})
	// }
	// steps = append(steps,
	// 	&stepConfigAlicloudSecurityGroup{
	// 		SecurityGroupId:   b.config.SecurityGroupId,
	// 		SecurityGroupName: b.config.SecurityGroupId,
	// 		RegionId:          b.config.AlicloudRegion,
	// 	},
	// create instance
	steps = append(steps,
		&stepCreateAlicloudInstance{
			AvailableZone: b.config.AvailableZone,
			InstanceSpec:  b.config.InstanceSpec,
			RootDiskType:  b.config.RootDiskType,
			RootDiskSize:  b.config.RootDiskSize,
			MountBns:      b.config.MountBns,
		})
	// if b.chooseNetworkType() == InstanceNetworkVpc {
	// 	steps = append(steps, &stepConfigAlicloudEIP{
	// 		AssociatePublicIpAddress: b.config.AssociatePublicIpAddress,
	// 		RegionId:                 b.config.AlicloudRegion,
	// 		InternetChargeType:       b.config.InternetChargeType,
	// 		InternetMaxBandwidthOut:  b.config.InternetMaxBandwidthOut,
	// 		SSHPrivateIp:             b.config.SSHPrivateIp,
	// 	})
	// } else {
	// 	steps = append(steps, &stepConfigAlicloudPublicIP{
	// 		RegionId:     b.config.AlicloudRegion,
	// 		SSHPrivateIp: b.config.SSHPrivateIp,
	// 	})
	// }
	steps = append(steps,
		&communicator.StepConnect{
			Config:    &b.config.RunConfig.Comm,
			Host:      SSHHost(client),
			SSHConfig: b.config.RunConfig.Comm.SSHConfigFunc(),
		},
		&commonsteps.StepProvision{},
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.config.RunConfig.Comm,
		},
		// &stepStopAlicloudInstance{
		// 	ForceStop:   b.config.ForceStopInstance,
		// 	DisableStop: b.config.DisableStopInstance,
		// },
	)
	// 	&stepDeleteAlicloudImageSnapshots{
	// 		AlicloudImageForceDeleteSnapshots: b.config.AlicloudImageForceDeleteSnapshots,
	// 		AlicloudImageForceDelete:          b.config.AlicloudImageForceDelete,
	// 		AlicloudImageName:                 b.config.AlicloudImageName,
	// 		AlicloudImageDestinationRegions:   b.config.AlicloudImageConfig.AlicloudImageDestinationRegions,
	// 		AlicloudImageDestinationNames:     b.config.AlicloudImageConfig.AlicloudImageDestinationNames,
	// 	})

	// if b.config.AlicloudImageIgnoreDataDisks {
	// 	steps = append(steps, &stepCreateAlicloudSnapshot{
	// 		WaitSnapshotReadyTimeout: b.getSnapshotReadyTimeout(),
	// 	})
	// }

	steps = append(steps,
		&stepCreateAlicloudImage{},
	)
	// 	&stepCreateTags{
	// 		Tags: b.config.AlicloudImageTags,
	// 	},
	// 	&stepRegionCopyAlicloudImage{
	// 		AlicloudImageDestinationRegions: b.config.AlicloudImageDestinationRegions,
	// 		AlicloudImageDestinationNames:   b.config.AlicloudImageDestinationNames,
	// 		RegionId:                        b.config.AlicloudRegion,
	// 	},
	// 	&stepShareAlicloudImage{
	// 		AlicloudImageShareAccounts:   b.config.AlicloudImageShareAccounts,
	// 		AlicloudImageUNShareAccounts: b.config.AlicloudImageUNShareAccounts,
	// 		RegionId:                     b.config.AlicloudRegion,
	// 	})

	// Run!
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If there are no ECS images, then just return
	if _, ok := state.GetOk("alicloudimages"); !ok {
		return nil, nil
	}

	// Build the artifact and return it
	artifact := &Artifact{
		Region:         b.config.Region,
		ImageId:        state.Get("dest_image_id").(string),
		BuilderIdValue: BuilderId,
		Client:         client,
	}

	return artifact, nil
}

func (b *Builder) chooseNetworkType() InstanceNetWork {
	if b.isVpcNetRequired() {
		return InstanceNetworkVpc
	} else {
		return InstanceNetworkClassic
	}
}

func (b *Builder) isVpcNetRequired() bool {
	// UserData and KeyPair only works in VPC
	return b.isVpcSpecified() || b.isUserDataNeeded() || b.isKeyPairNeeded()
}

func (b *Builder) isVpcSpecified() bool {
	return false
	// return b.config.VpcId != "" || b.config.VSwitchId != ""
}

func (b *Builder) isUserDataNeeded() bool {
	// Public key setup requires userdata
	if b.config.RunConfig.Comm.SSHPrivateKeyFile != "" {
		return true
	}
	return false

	// return b.config.UserData != "" || b.config.UserDataFile != ""
}

func (b *Builder) isKeyPairNeeded() bool {
	return b.config.Comm.SSHKeyPairName != "" || b.config.Comm.SSHTemporaryKeyPairName != ""
}

func (b *Builder) getSnapshotReadyTimeout() int {
	// if b.config.WaitSnapshotReadyTimeout > 0 {
	// 	return b.config.WaitSnapshotReadyTimeout
	// }

	return ALICLOUD_DEFAULT_LONG_TIMEOUT
}
