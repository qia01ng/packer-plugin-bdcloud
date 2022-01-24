package ecs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	rshttp "icode.baidu.com/baidu/lsqm-iam/sdk-go/http"
	"icode.baidu.com/baidu/lsqm-iam/sdk-go/iam"
)

type stepCreateAlicloudInstance struct {
	AvailableZone string
	InstanceSpec  string
	RootDiskType  string
	RootDiskSize  int
	MountBns      string

	instance *ecs.Instance
}

type InstanceSpec struct {
	AutoSeqSuffix       bool
	Billing             map[string]string
	Hostname            string
	Spec                string
	PurchaseCount       int
	CreateCdsList       []DiskSpec
	RootDiskSizeInGb    int
	RootDiskStorageType string
	ImageId             string
	RelationTag         bool
	Tags                []map[string]string
}

type InstancePurchaseSpec struct {
	Request        InstanceSpec
	AvailZone      string
	ProductName    string
	CreateUsername string
	Remark         string
	NoahPath       string
	BatchSize      int
}

type DiskSpec struct {
	StorageType string
	CdsSizeInGb int
	SnapshotId  string
}

var createInstanceRetryErrors = []string{
	"rs_servier_error",
	"waiting_all_instances_ok",
}

var deleteInstanceRetryErrors = []string{
	"IncorrectInstanceStatus.Initializing",
}

func (s *stepCreateAlicloudInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	rsClient := state.Get("rsClient").(*RsClientWrapper)
	// client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Creating instance...")
	createInstanceRequest, err := s.buildCreateInstanceRequest(state)
	if err != nil {
		return halt(state, err, "")
	}

	createInstanceResponse, err := rsClient.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (interface{}, error) {
			return rsClient.CreateInstance(createInstanceRequest)
		},
		EvalFunc: rsClient.EvalCouldRetryResponse(createInstanceRetryErrors, EvalRetryErrorType),
	})
	if err != nil {
		return halt(state, err, "Error creating instance")
	}

	order_id, err := rsClient.GetPurchaseId(createInstanceResponse.(*iam.BceResponse))
	if err != nil {
		return halt(state, err, "Error creating instance")
	}

	_okInstances, err := rsClient.WaitForExpected(&WaitForExpectArgs{
		RetryInterval: 60,
		RetryTimeout:  3600,
		RequestFunc: func() (interface{}, error) {
			return rsClient.GetOkInstances(order_id, 1)
		},
		EvalFunc: rsClient.EvalCouldRetryResponse(createInstanceRetryErrors, EvalRetryErrorType),
	})

	okInstances := _okInstances.([]InstanceInfo)
	if err != nil || len(okInstances) == 0 {
		return halt(state, err, "Error creating instance")
	}

	okInstance := okInstances[0]

	ui.Message(fmt.Sprintf("Created instance: %s", okInstance.InsId))
	state.Put("instance_id", okInstance.InsId)
	state.Put("instance_ip", okInstance.Ip)
	state.Put("instance_hostname", okInstance.Hostname)

	return multistep.ActionContinue
}

func (s *stepCreateAlicloudInstance) Cleanup(state multistep.StateBag) {}

func (s *stepCreateAlicloudInstance) buildCreateInstanceRequest(state multistep.StateBag) (*iam.BceRequest, error) {
	instancePurchaseSpec := InstancePurchaseSpec{
		Request: InstanceSpec{
			AutoSeqSuffix: true,
			Billing: map[string]string{
				"paymentTiming": "Postpaid",
			},
			Hostname:            s.AvailableZone + "image-marker-",
			Spec:                s.InstanceSpec,
			PurchaseCount:       1,
			RootDiskSizeInGb:    s.RootDiskSize,
			RootDiskStorageType: s.RootDiskType,
			ImageId:             state.Get("source_image_id").(string),
			RelationTag:         true,
			Tags: []map[string]string{
				map[string]string{
					"tagKey":   "业务",
					"tagValue": "search",
				},
			},
		},
		AvailZone:      s.AvailableZone,
		ProductName:    "COS",
		CreateUsername: "guozhiqiang06",
		Remark:         "搜索单元化弹性上云",
		NoahPath:       "bcc.www.all",
		BatchSize:      1,
	}

	json_data, err := json.Marshal(instancePurchaseSpec)
	if err != nil {
		return nil, err
	}

	bceRequest := &iam.BceRequest{}
	body, err := iam.NewBodyFromBytes(json_data)
	if err != nil {
		return nil, err
	}
	bceRequest.SetBody(body)
	bceRequest.SetUri("/api/asset/v1/bcc/instanceBySpec")
	bceRequest.SetMethod(rshttp.POST)
	bceRequest.SetHeader(rshttp.CONTENT_TYPE, iam.DEFAULT_CONTENT_TYPE)

	return bceRequest, nil
}

func (s *stepCreateAlicloudInstance) buildCreateInstanceTags(tags map[string]string) *[]ecs.CreateInstanceTag {
	var ecsTags []ecs.CreateInstanceTag

	for k, v := range tags {
		ecsTags = append(ecsTags, ecs.CreateInstanceTag{Key: k, Value: v})
	}

	return &ecsTags
}
