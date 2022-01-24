package ecs

import (
	"context"
	"fmt"

	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepCreateAlicloudImage struct {
}

var createImageRetryErrors = []string{
	"IdempotentProcessing",
	"InProcessing",
}

func (s *stepCreateAlicloudImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// config := state.Get("config").(*Config)
	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packersdk.Ui)

	createImageRequest := s.buildCreateImageRequest(state)
	createImageResponse, err := client.WaitForExpected(&WaitForExpectArgs{
		// RequestFunc: func() (api.CreateImageResult, error) {
		RequestFunc: func() (interface{}, error) {
			return client.CreateImage(createImageRequest)
		},
		EvalFunc: client.EvalCouldRetryImageCreateResponse(createImageRetryErrors, EvalRetryErrorType),
	})

	if err != nil {
		return halt(state, err, "Error creating image")
	}

	imageId := createImageResponse.(api.CreateImageResult).ImageId
	state.Put("dest_image_id", imageId)

	imageDetailInfo, err := client.WaitForExpected(&WaitForExpectArgs{
		RetryInterval: 60,
		RetryTimeout:  3600,
		RequestFunc: func() (interface{}, error) {
			return client.CheckImageOk(imageId)
		},
		EvalFunc: client.EvalCouldRetryCheckImageOk(createImageRetryErrors, EvalRetryErrorType),
	})

	if err != nil {
		return halt(state, err, "Error creating image")
	}
	imageName := imageDetailInfo.(api.GetImageDetailResult).Image.Name
	state.Put("dest_image_name", imageName)
	ui.Message(fmt.Sprintf("Create image success, image id is: %s, image name is: %s", imageId, imageName))

	return multistep.ActionContinue
}

func (s *stepCreateAlicloudImage) Cleanup(state multistep.StateBag) {}

func (s *stepCreateAlicloudImage) buildCreateImageRequest(state multistep.StateBag) *api.CreateImageArgs {
	config := state.Get("config").(*Config)

	request := &api.CreateImageArgs{
		ImageName:   config.DstImageName + config.DstImageVersion,
		InstanceId:  state.Get("instance_id").(string),
		IsRelateCds: false,
	}

	return request
}
