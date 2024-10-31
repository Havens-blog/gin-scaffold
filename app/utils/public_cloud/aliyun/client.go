package aliyun

import (
	"fmt"
	"yunion.io/x/cloudmux/pkg/multicloud/aliyun"
)

type BaseOptions struct {
	Debug      bool   `help:"debug mode"`
	CloudEnv   string `help:"Cloud environment" default:"$ALIYUN_CLOUD_ENV" choices:"InternationalCloud|FinanceCloud" metavar:"ALIYUN_CLOUD_ENV"`
	AccessKey  string `help:"Access key" default:"$ALIYUN_ACCESS_KEY" metavar:"ALIYUN_ACCESS_KEY"`
	Secret     string `help:"Secret" default:"$ALIYUN_SECRET" metavar:"ALIYUN_SECRET"`
	RegionId   string `help:"RegionId" default:"$ALIYUN_REGION" metavar:"ALIYUN_REGION"`
	AccountId  string `help:"AccountId" default:"$ALIYUN_ACCOUNT_ID" metavar:"ALIYUN_ACCOUNT_ID"`
	SUBCOMMAND string `help:"aliyuncli subcommand" subcommand:"true"`
}

func NewClient(options *BaseOptions) (*aliyun.SRegion, error) {
	if len(options.AccessKey) == 0 {
		return nil, fmt.Errorf("Missing accessKey")
	}

	if len(options.Secret) == 0 {
		return nil, fmt.Errorf("Missing secret")
	}

	cli, err := aliyun.NewAliyunClient(
		aliyun.NewAliyunClientConfig(
			options.CloudEnv,
			options.AccessKey,
			options.Secret,
		),
	)
	if err != nil {
		return nil, err
	}

	region := cli.GetRegion(options.RegionId)
	if region == nil {

		return nil, fmt.Errorf("No such region %s", options.RegionId)
	}

	return region, nil
}
