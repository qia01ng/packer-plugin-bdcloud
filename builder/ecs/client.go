package ecs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	// "github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"icode.baidu.com/baidu/lsqm-iam/sdk-go/client"
	rshttp "icode.baidu.com/baidu/lsqm-iam/sdk-go/http"
	"icode.baidu.com/baidu/lsqm-iam/sdk-go/iam"
)

type ClientWrapper struct {
	*bcc.Client
}

type RsClientWrapper struct {
	*client.AuthApiClient
}

// type RsClient interface {
// 	Get() error
// }

// type RsAKSKAuthApiClient struct {
// 	*client.AuthApiClient
// }

const (
	InstanceStatusRunning  = "Running"
	InstanceStatusStarting = "Starting"
	InstanceStatusStopped  = "Stopped"
	InstanceStatusStopping = "Stopping"
)

const (
	ImageStatusWaiting      = "Waiting"
	ImageStatusCreating     = "Creating"
	ImageStatusCreateFailed = "CreateFailed"
	ImageStatusAvailable    = "Available"
)

var ImageStatusQueried = fmt.Sprintf("%s,%s,%s,%s", ImageStatusWaiting, ImageStatusCreating, ImageStatusCreateFailed, ImageStatusAvailable)

const (
	SnapshotStatusAll          = "all"
	SnapshotStatusProgressing  = "progressing"
	SnapshotStatusAccomplished = "accomplished"
	SnapshotStatusFailed       = "failed"
)

const (
	DiskStatusInUse     = "In_use"
	DiskStatusAvailable = "Available"
	DiskStatusAttaching = "Attaching"
	DiskStatusDetaching = "Detaching"
	DiskStatusCreating  = "Creating"
	DiskStatusReIniting = "ReIniting"
)

const (
	VpcStatusPending   = "Pending"
	VpcStatusAvailable = "Available"
)

const (
	VSwitchStatusPending   = "Pending"
	VSwitchStatusAvailable = "Available"
)

const (
	EipStatusAssociating   = "Associating"
	EipStatusUnassociating = "Unassociating"
	EipStatusInUse         = "InUse"
	EipStatusAvailable     = "Available"
)

const (
	ImageOwnerSystem      = "system"
	ImageOwnerSelf        = "self"
	ImageOwnerOthers      = "others"
	ImageOwnerMarketplace = "marketplace"
)

const (
	IOOptimizedNone      = "none"
	IOOptimizedOptimized = "optimized"
)

const (
	InstanceNetworkClassic = "classic"
	InstanceNetworkVpc     = "vpc"
)

const (
	DiskTypeSystem = "system"
	DiskTypeData   = "data"
)

const (
	TagResourceImage    = "image"
	TagResourceInstance = "instance"
	TagResourceSnapshot = "snapshot"
	TagResourceDisk     = "disk"
)

const (
	IpProtocolAll  = "all"
	IpProtocolTCP  = "tcp"
	IpProtocolUDP  = "udp"
	IpProtocolICMP = "icmp"
	IpProtocolGRE  = "gre"
)

const (
	NicTypeInternet = "internet"
	NicTypeIntranet = "intranet"
)

const (
	DefaultPortRange = "-1/-1"
	DefaultCidrIp    = "0.0.0.0/0"
	DefaultCidrBlock = "172.16.0.0/24"
)

const (
	defaultRetryInterval = 5 * time.Second
	defaultRetryTimes    = 12
	shortRetryTimes      = 36
	mediumRetryTimes     = 360
)

type WaitForExpectEvalResult struct {
	evalPass  bool
	stopRetry bool
}

var (
	WaitForExpectSuccess = WaitForExpectEvalResult{
		evalPass:  true,
		stopRetry: true,
	}

	WaitForExpectToRetry = WaitForExpectEvalResult{
		evalPass:  false,
		stopRetry: false,
	}

	WaitForExpectFailToStop = WaitForExpectEvalResult{
		evalPass:  false,
		stopRetry: true,
	}
)

type WaitForExpectArgs struct {
	RequestFunc   func() (interface{}, error)
	EvalFunc      func(response interface{}, err error) WaitForExpectEvalResult
	RetryInterval time.Duration
	RetryTimes    int
	RetryTimeout  time.Duration
}

// func (c *ClientWrapper) WaitForExpected(args *WaitForExpectArgs) (responses.AcsResponse, error) {
// 	if args.RetryInterval <= 0 {
// 		args.RetryInterval = defaultRetryInterval
// 	}
// 	if args.RetryTimes <= 0 {
// 		args.RetryTimes = defaultRetryTimes
// 	}

// 	var timeoutPoint time.Time
// 	if args.RetryTimeout > 0 {
// 		timeoutPoint = time.Now().Add(args.RetryTimeout)
// 	}

// 	var lastResponse responses.AcsResponse
// 	var lastError error

// 	for i := 0; ; i++ {
// 		if args.RetryTimeout > 0 && time.Now().After(timeoutPoint) {
// 			break
// 		}

// 		if args.RetryTimeout <= 0 && i >= args.RetryTimes {
// 			break
// 		}

// 		response, err := args.RequestFunc()
// 		lastResponse = response
// 		lastError = err

// 		evalResult := args.EvalFunc(response, err)
// 		if evalResult.evalPass {
// 			return response, nil
// 		}
// 		if evalResult.stopRetry {
// 			return response, err
// 		}

// 		time.Sleep(args.RetryInterval)
// 	}

// 	if lastError == nil {
// 		lastError = fmt.Errorf("<no error>")
// 	}

// 	if args.RetryTimeout > 0 {
// 		return lastResponse, fmt.Errorf("evaluate failed after %d seconds timeout with %d seconds retry interval: %s", int(args.RetryTimeout.Seconds()), int(args.RetryInterval.Seconds()), lastError)
// 	}

// 	return lastResponse, fmt.Errorf("evaluate failed after %d times retry with %d seconds retry interval: %s", args.RetryTimes, int(args.RetryInterval.Seconds()), lastError)
// }

func (c *ClientWrapper) WaitForExpected(args *WaitForExpectArgs) (interface{}, error) {
	if args.RetryInterval <= 0 {
		args.RetryInterval = defaultRetryInterval
	}
	if args.RetryTimes <= 0 {
		args.RetryTimes = defaultRetryTimes
	}

	var timeoutPoint time.Time
	if args.RetryTimeout > 0 {
		timeoutPoint = time.Now().Add(args.RetryTimeout)
	}

	var lastResponse interface{}
	var lastError error

	for i := 0; ; i++ {
		if args.RetryTimeout > 0 && time.Now().After(timeoutPoint) {
			break
		}

		if args.RetryTimeout <= 0 && i >= args.RetryTimes {
			break
		}

		response, err := args.RequestFunc()
		lastResponse = response
		lastError = err

		evalResult := args.EvalFunc(response, err)
		if evalResult.evalPass {
			return response, nil
		}
		if evalResult.stopRetry {
			return response, err
		}

		time.Sleep(args.RetryInterval)
	}

	if lastError == nil {
		lastError = fmt.Errorf("<no error>")
	}

	if args.RetryTimeout > 0 {
		return lastResponse, fmt.Errorf("evaluate failed after %d seconds timeout with %d seconds retry interval: %s", int(args.RetryTimeout.Seconds()), int(args.RetryInterval.Seconds()), lastError)
	}

	return lastResponse, fmt.Errorf("evaluate failed after %d times retry with %d seconds retry interval: %s", args.RetryTimes, int(args.RetryInterval.Seconds()), lastError)
}

// func (c *ClientWrapper) WaitForInstanceStatus(regionId string, instanceId string, expectedStatus string) (responses.AcsResponse, error) {
// 	return c.WaitForExpected(&WaitForExpectArgs{
// 		RequestFunc: func() (responses.AcsResponse, error) {
// 			request := ecs.CreateDescribeInstancesRequest()
// 			request.RegionId = regionId
// 			request.InstanceIds = fmt.Sprintf("[\"%s\"]", instanceId)
// 			return c.DescribeInstances(request)
// 		},
// 		EvalFunc: func(response responses.AcsResponse, err error) WaitForExpectEvalResult {
// 			if err != nil {
// 				return WaitForExpectToRetry
// 			}

// 			instancesResponse := response.(*ecs.DescribeInstancesResponse)
// 			instances := instancesResponse.Instances.Instance
// 			for _, instance := range instances {
// 				if instance.Status == expectedStatus {
// 					return WaitForExpectSuccess
// 				}
// 			}
// 			return WaitForExpectToRetry
// 		},
// 		RetryTimes: mediumRetryTimes,
// 	})
// }

// func (c *ClientWrapper) WaitForImageStatus(regionId string, imageId string, expectedStatus string, timeout time.Duration) (responses.AcsResponse, error) {
// 	return c.WaitForExpected(&WaitForExpectArgs{
// 		RequestFunc: func() (responses.AcsResponse, error) {
// 			request := ecs.CreateDescribeImagesRequest()
// 			request.RegionId = regionId
// 			request.ImageId = imageId
// 			request.Status = ImageStatusQueried
// 			return c.DescribeImages(request)
// 		},
// 		EvalFunc: func(response responses.AcsResponse, err error) WaitForExpectEvalResult {
// 			if err != nil {
// 				return WaitForExpectToRetry
// 			}

// 			imagesResponse := response.(*ecs.DescribeImagesResponse)
// 			images := imagesResponse.Images.Image
// 			for _, image := range images {
// 				if image.Status == expectedStatus {
// 					return WaitForExpectSuccess
// 				}
// 			}

// 			return WaitForExpectToRetry
// 		},
// 		RetryTimeout: timeout,
// 	})
// }

// func (c *ClientWrapper) WaitForSnapshotStatus(regionId string, snapshotId string, expectedStatus string, timeout time.Duration) (responses.AcsResponse, error) {
// 	return c.WaitForExpected(&WaitForExpectArgs{
// 		RequestFunc: func() (responses.AcsResponse, error) {
// 			request := ecs.CreateDescribeSnapshotsRequest()
// 			request.RegionId = regionId
// 			request.SnapshotIds = fmt.Sprintf("[\"%s\"]", snapshotId)
// 			return c.DescribeSnapshots(request)
// 		},
// 		EvalFunc: func(response responses.AcsResponse, err error) WaitForExpectEvalResult {
// 			if err != nil {
// 				return WaitForExpectToRetry
// 			}

// 			snapshotsResponse := response.(*ecs.DescribeSnapshotsResponse)
// 			snapshots := snapshotsResponse.Snapshots.Snapshot
// 			for _, snapshot := range snapshots {
// 				if snapshot.Status == expectedStatus {
// 					return WaitForExpectSuccess
// 				}
// 			}
// 			return WaitForExpectToRetry
// 		},
// 		RetryTimeout: timeout,
// 	})
// }

type EvalErrorType bool

const (
	EvalRetryErrorType    = EvalErrorType(true)
	EvalNotRetryErrorType = EvalErrorType(false)
)

func (c *ClientWrapper) EvalCouldRetryImageCreateResponse(evalErrors []string, evalErrorType EvalErrorType) func(response interface{}, err error) WaitForExpectEvalResult {
	return func(response interface{}, err error) WaitForExpectEvalResult {
		if err == nil {
			return WaitForExpectSuccess
		}

		if evalErrorType == EvalRetryErrorType && !ContainsInArray(evalErrors, err.Error()) {
			return WaitForExpectFailToStop
		}

		if evalErrorType == EvalNotRetryErrorType && ContainsInArray(evalErrors, err.Error()) {
			return WaitForExpectFailToStop
		}

		return WaitForExpectToRetry
	}
}

func (c *ClientWrapper) CheckImageOk(image_id string) (*api.GetImageDetailResult, error) {
	result, err := c.GetImageDetail(image_id)
	if err != nil {
		return nil, errors.New("request_error")
	}
	if result.Image.Status == "OK" {
		return result, nil
	} else if result.Image.Status == "Making" {
		return result, errors.New("InProcessing")
	} else {
		return result, errors.New("CreateImageFailed")
	}
}

func (c *ClientWrapper) EvalCouldRetryCheckImageOk(evalErrors []string, evalErrorType EvalErrorType) func(response interface{}, err error) WaitForExpectEvalResult {
	return func(response interface{}, err error) WaitForExpectEvalResult {
		if err == nil {
			return WaitForExpectSuccess
		}

		if evalErrorType == EvalRetryErrorType && !ContainsInArray(evalErrors, err.Error()) {
			return WaitForExpectFailToStop
		}

		if evalErrorType == EvalNotRetryErrorType && ContainsInArray(evalErrors, err.Error()) {
			return WaitForExpectFailToStop
		}

		return WaitForExpectToRetry
	}
}

func (c *ClientWrapper) CheckImageExist(image_id string) bool {
	_, err := c.Client.GetImageDetail(image_id)
	if err != nil {
		return false
	}
	return true
}

func (c *ClientWrapper) ListImages(imageType string, imageName string) ([]string, error) {
	if imageType == "" {
		imageType = "All"
	}
	args := &api.ListImageArgs{
		ImageType: imageType,
		ImageName: imageName,
	}

	if res, err := c.Client.ListImage(args); err != nil {
		return nil, err
	} else {
		image_ids := make([]string, 1)
		for _, imageModel := range res.Images {
			image_ids = append(image_ids, imageModel.Id)
		}
		return image_ids, nil
	}
}

func (c *RsClientWrapper) WaitForExpected(args *WaitForExpectArgs) (interface{}, error) {
	if args.RetryInterval <= 0 {
		args.RetryInterval = defaultRetryInterval
	}
	if args.RetryTimes <= 0 {
		args.RetryTimes = defaultRetryTimes
	}

	var timeoutPoint time.Time
	if args.RetryTimeout > 0 {
		timeoutPoint = time.Now().Add(args.RetryTimeout)
	}

	var lastResponse interface{}
	var lastError error

	for i := 0; ; i++ {
		if args.RetryTimeout > 0 && time.Now().After(timeoutPoint) {
			break
		}

		if args.RetryTimeout <= 0 && i >= args.RetryTimes {
			break
		}

		response, err := args.RequestFunc()
		lastResponse = response
		lastError = err

		evalResult := args.EvalFunc(response, err)
		if evalResult.evalPass {
			return response, nil
		}
		if evalResult.stopRetry {
			return response, err
		}

		time.Sleep(args.RetryInterval)
	}

	if lastError == nil {
		lastError = fmt.Errorf("<no error>")
	}

	if args.RetryTimeout > 0 {
		return lastResponse, fmt.Errorf("evaluate failed after %d seconds timeout with %d seconds retry interval: %s", int(args.RetryTimeout.Seconds()), int(args.RetryInterval.Seconds()), lastError)
	}

	return lastResponse, fmt.Errorf("evaluate failed after %d times retry with %d seconds retry interval: %s", args.RetryTimes, int(args.RetryInterval.Seconds()), lastError)
}

func (c *RsClientWrapper) EvalCouldRetryResponse(evalErrors []string, evalErrorType EvalErrorType) func(response interface{}, err error) WaitForExpectEvalResult {
	return func(response interface{}, err error) WaitForExpectEvalResult {
		if err == nil {
			return WaitForExpectSuccess
		}

		if evalErrorType == EvalRetryErrorType && !ContainsInArray(evalErrors, err.Error()) {
			return WaitForExpectFailToStop
		}

		if evalErrorType == EvalNotRetryErrorType && ContainsInArray(evalErrors, err.Error()) {
			return WaitForExpectFailToStop
		}

		return WaitForExpectToRetry
	}
}

type PurchaseRet struct {
	PurchaseOrderUuid string
}

type InstanceInfo struct {
	Ip       string
	Hostname string
	InsId    string
}

type PurchaseProgress struct {
	MachineList []InstanceInfo
}

func (c *RsClientWrapper) CreateInstance(request *iam.BceRequest) (*iam.BceResponse, error) {
	resp := &iam.BceResponse{}

	if err := c.AuthApiClient.SendRequest(request, resp); err != nil {
		return nil, err
	}
	defer func() { resp.Body().Close() }()

	if resp.StatusCode() >= 500 {
		return nil, errors.New("rs_server_error")
	}

	if resp.StatusCode() == 200 {
		return resp, nil
	} else {
		return nil, fmt.Errorf("serror_error_with_code: %d", resp.StatusCode())
	}
}

func (c *RsClientWrapper) GetOkInstances(purchaseOrderId string, expectedCount int) ([]InstanceInfo, error) {

	uri := fmt.Sprintf("/api/asset/v1/order/detail/%s", purchaseOrderId)
	bceRequest := &iam.BceRequest{}
	bceRequest.SetUri(uri)
	bceRequest.SetMethod(rshttp.GET)
	bceRequest.SetHeader(rshttp.CONTENT_TYPE, iam.DEFAULT_CONTENT_TYPE)

	resp := &iam.BceResponse{}
	if err := c.AuthApiClient.SendRequest(bceRequest, resp); err != nil {
		return nil, err
	}
	defer func() { resp.Body().Close() }()

	if resp.StatusCode() >= 500 {
		return nil, errors.New("rs_server_error")
	}

	if resp.StatusCode() == 200 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body())
		purchaseProgress := PurchaseProgress{}
		err := json.Unmarshal(buf.Bytes(), &purchaseProgress)
		if err != nil {
			return nil, err
		}

		// 获取ok的实例ip
		okInstances := make([]InstanceInfo, 0, expectedCount)
		for _, insInfo := range purchaseProgress.MachineList {
			if insInfo.Ip != "" && insInfo.Hostname != "" && insInfo.InsId != "" {
				okInstances = append(okInstances, insInfo)
			}
		}

		if len(okInstances) == expectedCount {
			return okInstances, nil
		} else {
			return okInstances, errors.New("waiting_all_instances_ok")
		}
	} else {
		return nil, fmt.Errorf("serror_error_with_code: %d", resp.StatusCode())
	}
}

func (c *RsClientWrapper) GetPurchaseId(resp *iam.BceResponse) (string, error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body())
	purchaseRet := PurchaseRet{}
	err := json.Unmarshal(buf.Bytes(), &purchaseRet)
	if err != nil {
		return "", err
	} else {
		return purchaseRet.PurchaseOrderUuid, nil
	}
}
