package ecs

import (
	"fmt"
)

type Artifact struct {
	// A map of regions to alicloud image IDs.
	Region  string
	ImageId string
	// AlicloudImages map[string]string

	// BuilderId is the unique ID for the builder that created this alicloud image
	BuilderIdValue string

	// Alcloud connection for performing API stuff.
	Client *ClientWrapper
}

func (a *Artifact) BuilderId() string {
	return a.BuilderIdValue
}

func (*Artifact) Files() []string {
	// We have no files
	return nil
}

func (a *Artifact) Id() string {
	return a.Region + "." + a.ImageId
}

func (a *Artifact) String() string {
	return fmt.Sprintf("Alicloud image were created:\n%s", a.Region+"."+a.ImageId)
}

func (a *Artifact) State(name string) interface{} {
	return nil
	// switch name {
	// case "atlas.artifact.metadata":
	// 	return a.stateAtlasMetadata()
	// default:
	// 	return nil
	// }
}

func (a *Artifact) Destroy() error {
	return a.Client.Client.DeleteImage(a.ImageId)
}

// func (a *Artifact) unsharedAccountsOnImages(regionId string, imageId string) []error {
// 	var errors []error

// 	describeImageShareRequest := ecs.CreateDescribeImageSharePermissionRequest()
// 	describeImageShareRequest.RegionId = regionId
// 	describeImageShareRequest.ImageId = imageId
// 	imageShareResponse, err := a.Client.DescribeImageSharePermission(describeImageShareRequest)
// 	if err != nil {
// 		errors = append(errors, err)
// 		return errors
// 	}

// 	accountsNumber := len(imageShareResponse.Accounts.Account)
// 	if accountsNumber > 0 {
// 		accounts := make([]string, accountsNumber)
// 		for index, account := range imageShareResponse.Accounts.Account {
// 			accounts[index] = account.AliyunId
// 		}

// 		modifyImageShareRequest := ecs.CreateModifyImageSharePermissionRequest()
// 		modifyImageShareRequest.RegionId = regionId
// 		modifyImageShareRequest.ImageId = imageId
// 		modifyImageShareRequest.RemoveAccount = &accounts
// 		_, err := a.Client.ModifyImageSharePermission(modifyImageShareRequest)
// 		if err != nil {
// 			errors = append(errors, err)
// 		}
// 	}

// 	return errors
// }

// func (a *Artifact) stateAtlasMetadata() interface{} {
// 	metadata := make(map[string]string)
// 	for region, imageId := range a.AlicloudImages {
// 		k := fmt.Sprintf("region.%s", region)
// 		metadata[k] = imageId
// 	}

// 	return metadata
// }
