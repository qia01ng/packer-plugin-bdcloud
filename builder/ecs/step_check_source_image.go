package ecs

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepCheckAlicloudSourceImage struct {
	SrcImageId   string
	SrcImageType string
	SrcImageName string
}

func (s *stepCheckAlicloudSourceImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ClientWrapper)
	// config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	if s.SrcImageId != "" {
		exist := client.CheckImageExist(s.SrcImageId)
		if !exist {
			return halt(state, fmt.Errorf("can not find image by id[%s]", s.SrcImageId), "")
		}
		state.Put("source_image_id", s.SrcImageId)
		return multistep.ActionContinue
	}
	imageIds, err := client.ListImages(s.SrcImageType, s.SrcImageName)
	if err != nil {
		return halt(state, err, "Error querying bdcloud image")
	}

	if len(imageIds) == 0 {
		err := fmt.Errorf("No bdcloud image was found matching filters: imageType[%s] and imageName[%s]", s.SrcImageType, s.SrcImageName)
		return halt(state, err, "")
	}

	ui.Message(fmt.Sprintf("Found image ID: %s", imageIds[0]))

	state.Put("source_image_id", imageIds[0])
	return multistep.ActionContinue
}

func (s *stepCheckAlicloudSourceImage) Cleanup(multistep.StateBag) {}
