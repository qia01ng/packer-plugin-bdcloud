//go:generate packer-sdc struct-markdown

package ecs

import (
	"errors"
	"fmt"
	"os"
	"time"

	// "github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"

	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"icode.baidu.com/baidu/lsqm-iam/sdk-go/auth"
	"icode.baidu.com/baidu/lsqm-iam/sdk-go/client"
	"icode.baidu.com/baidu/lsqm-iam/sdk-go/iam"
)

// Config of alicloud
type AlicloudAccessConfig struct {
	RsAccessKey string `mapstructure:"rs_access_key" required:"true"`
	RsSecretKey string `mapstructure:"rs_secret_key" required:"true"`
	RsEndpoint  string `mapstructure:"rs_endpoint" required:"false"`
	rsClient    *RsClientWrapper

	BceAccessKey string `mapstructure:"bce_access_key" required:"true"`
	BceSecretKey string `mapstructure:"bce_secret_key" required:"true"`
	Region       string `mapstructure:"region" required:"false"`
	Endpoint     string `mapstructure:"endpoint" required:"false"`
	client       *ClientWrapper
}

const Packer = "HashiCorp-Packer"
const DefaultRequestReadTimeout = 10 * time.Second

// Client for AlicloudClient
func (c *AlicloudAccessConfig) RsClient() (*RsClientWrapper, error) {
	if c.rsClient != nil {
		return c.rsClient, nil
	}

	if c.RsEndpoint == "" {
		c.RsEndpoint = os.Getenv("RS_ENDPOINT")
	}
	if c.RsEndpoint == "" {
		return nil, errors.New("rs_endpoint must be provided in template file or environment variables")
	}

	if c.RsAccessKey == "" {
		c.RsAccessKey = os.Getenv("RS_ACCESS_KEY")
	}

	if c.RsSecretKey == "" {
		c.RsSecretKey = os.Getenv("RS_SECRET_KEY")
	}

	if c.RsAccessKey != "" && c.RsSecretKey != "" {
		config := iam.BceClientConfiguration{
			Endpoint: c.RsEndpoint,
			Credentials: &auth.BceCredentials{
				AccessKeyId:     c.RsAccessKey,
				SecretAccessKey: c.RsSecretKey,
			},
		}
		authApiClient := client.NewAuthApiClient(&config)

		c.rsClient = &RsClientWrapper{authApiClient}

		return c.rsClient, nil
	} else {
		return nil, errors.New("rs_access_key and rs_secret_key must be provided in template file or environment variables")
	}
}

// Client for AlicloudClient
func (c *AlicloudAccessConfig) Client() (*ClientWrapper, error) {
	if c.client != nil {
		return c.client, nil
	}

	if c.Endpoint == "" {
		c.Endpoint = os.Getenv("ENDPOINT")
	}
	if c.Endpoint == "" {
		return nil, errors.New("endpoint must be provided in template file or environment variables")
	}

	if c.Region == "" {
		c.Region = os.Getenv("REGION")
	}
	if c.Region == "" {
		return nil, errors.New("region must be provided in template file or environment variables")
	}

	if c.BceAccessKey == "" {
		c.BceAccessKey = os.Getenv("BCE_ACCESS_KEY")
	}

	if c.BceSecretKey == "" {
		c.BceSecretKey = os.Getenv("BCE_SECRET_KEY")
	}

	if c.BceAccessKey != "" && c.BceSecretKey != "" {
		bccClient, err := bcc.NewClient(c.BceAccessKey, c.BceSecretKey, c.Endpoint)
		if err != nil {
			return nil, err
		}
		// 配置不进行重试，默认为Back Off重试
		bccClient.Config.Retry = bce.NewNoRetryPolicy()
		// 配置连接超时时间为30秒
		bccClient.Config.ConnectionTimeoutInMillis = 30 * 1000

		c.client = &ClientWrapper{bccClient}

		return c.client, nil
	} else {
		return nil, errors.New("access_key and secret_key must be provided in template file or environment variables")
	}

	// var getProviderConfig = func(str string, key string) string {
	// 	value, err := getConfigFromProfile(c, key)
	// 	if err == nil && value != nil {
	// 		str = value.(string)
	// 	}
	// 	return str
	// }

	// // read config from profile
	// if c.AlicloudAccessKey == "" || c.AlicloudSecretKey == "" {
	// 	c.AlicloudAccessKey = getProviderConfig(c.AlicloudAccessKey, "access_key_id")
	// 	c.AlicloudSecretKey = getProviderConfig(c.AlicloudSecretKey, "access_key_secret")
	// 	c.AlicloudRegion = getProviderConfig(c.AlicloudRegion, "region_id")
	// 	c.SecurityToken = getProviderConfig(c.SecurityToken, "sts_token")
	// 	c.CustomEndpointEcs = getProviderConfig(c.CustomEndpointEcs, "endpoint")
	// }

	// if c.CustomEndpointEcs != "" && c.AlicloudRegion != "" {
	// 	_ = endpoints.AddEndpointMapping(c.AlicloudRegion, "Ecs", c.CustomEndpointEcs)
	// }

	// client, err := ecs.NewClientWithStsToken(c.AlicloudRegion, c.AlicloudAccessKey, c.AlicloudSecretKey, c.SecurityToken)
	// if err != nil {
	// 	return nil, err
	// }

	// client.AppendUserAgent(Packer, version.PluginVersion.FormattedVersion())
	// client.SetReadTimeout(DefaultRequestReadTimeout)
	// c.client = &ClientWrapper{client}

	// return c.client, nil
}

func (c *AlicloudAccessConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	if err := c.Config(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (c *AlicloudAccessConfig) Config() error {
	if c.Endpoint == "" {
		c.Endpoint = os.Getenv("endpoint")
	}

	if c.RsAccessKey == "" {
		c.RsAccessKey = os.Getenv("RS_ACCESS_KEY")
	}

	if c.RsSecretKey == "" {
		c.RsSecretKey = os.Getenv("RS_SECRET_KEY")
	}

	if c.RsAccessKey == "" || c.RsSecretKey == "" {
		return fmt.Errorf("rs_access_key/rs_secret_key must be set in template file")
	}

	if c.BceAccessKey == "" {
		c.BceAccessKey = os.Getenv("BCE_ACCESS_KEY")
	}

	if c.BceSecretKey == "" {
		c.BceSecretKey = os.Getenv("BCE_SECRET_KEY")
	}

	if c.BceAccessKey == "" {
		c.BceAccessKey = os.Getenv("access_key")
	}
	if c.BceSecretKey == "" {
		c.BceSecretKey = os.Getenv("secret_key")
	}
	if c.BceAccessKey == "" || c.BceSecretKey == "" || c.Endpoint == "" || c.Region == "" {
		return fmt.Errorf("access_key/secret_key/region/endpoint must be set in template file or environment variables.")
	}
	return nil

}
