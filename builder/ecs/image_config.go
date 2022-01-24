//go:generate packer-sdc struct-markdown

package ecs

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type AlicloudImageConfig struct {
	SrcImageId      string `mapstructure:"src_image_id" required:"false"`
	SrcImageType    string `mapstructure:"src_image_type" required:"false"`
	SrcImageName    string `mapstructure:"src_image_name" required:"false"`
	DstImageName    string `mapstructure:"dst_image_name" required:"true"`
	DstImageVersion string `mapstructure:"dst_image_version" required:"true"`
}

func (c *AlicloudImageConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	if c.SrcImageId == "" && c.SrcImageName == "" {
		errs = append(errs, errors.New("must provide src_image_id or src_image_name in template file"))
	}

	if err := c.ValidImageType(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (c *AlicloudImageConfig) ValidImageType() error {
	//// 指定要查询何种类型的镜像，包括All(所有)，System(系统镜像/公共镜像)，Custom(自定义镜像)，Integration(服务集成镜像)，Sharing(共享镜像)，GpuBccSystem(GPU专用公共镜像)，GpuBccCustom(GPU专用自定义镜像)，FpgaBccSystem(FPGA专用公共镜像)，FpgaBccCustom(FPGA专用自定义镜像)，缺省值为All

	validImageTypes := []string{"All", "System", "Custom", "Integration", "Sharing", "GpuBccSystem", "GpuBccCustom",
		"FpgaBccSystem", "FpgaBccCustom"}

	if c.SrcImageType == "" {
		return nil
	}

	for _, ele := range validImageTypes {
		if ele == c.SrcImageType {
			return nil
		}
	}

	return fmt.Errorf("invalid image type, valid types are: %s", strings.Join(validImageTypes, ","))
}
