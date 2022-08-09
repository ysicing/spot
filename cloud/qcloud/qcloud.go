package qcloud

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ergoapi/util/output"
	"github.com/gosuri/uitable"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type Client struct {
	*cvm.Client
}

type Instance struct {
	CreatedTime        string
	InstanceName       string
	InstanceId         string
	InstanceType       string
	InstanceChargeType string
	InstanceState      string
	PrivateIpAddresses string
}

func NewClient() *Client {
	// 实例化一个认证对象，入参需要传入腾讯云账户secretId，secretKey,此处还需注意密钥对的保密
	// 密钥可前往https://console.cloud.tencent.com/cam/capi网站进行获取
	credential := common.NewCredential(
		viper.GetString("qcloud.account.id"),
		viper.GetString("qcloud.account.secret"),
	)
	logrus.Debugf("credential: %s, %s", credential.SecretId, credential.SecretKey)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "cvm.tencentcloudapi.com"
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, _ := cvm.NewClient(credential, viper.GetString("qcloud.region"), cpf)
	return &Client{client}
}

func (c *Client) Create(count int64) error {

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := cvm.NewRunInstancesRequest()

	request.InstanceChargeType = common.StringPtr("SPOTPAID")
	request.Placement = &cvm.Placement{
		Zone: common.StringPtr(viper.GetString("qcloud.zone")),
	}
	if viper.GetInt64("qcloud.project.id") > 0 {
		request.Placement.ProjectId = common.Int64Ptr(viper.GetInt64("qcloud.project.id"))
	}
	request.InstanceType = common.StringPtr(viper.GetString("qcloud.instance.type"))
	request.ImageId = common.StringPtr(viper.GetString("qcloud.instance.image"))
	request.SystemDisk = &cvm.SystemDisk{
		DiskType: common.StringPtr("CLOUD_PREMIUM"),
		DiskSize: common.Int64Ptr(50),
	}
	request.VirtualPrivateCloud = &cvm.VirtualPrivateCloud{
		VpcId:            common.StringPtr(viper.GetString("qcloud.instance.network.vpc.id")),
		SubnetId:         common.StringPtr(viper.GetString("qcloud.instance.network.subnet.id")),
		AsVpcGateway:     common.BoolPtr(false),
		Ipv6AddressCount: common.Uint64Ptr(0),
	}
	request.InternetAccessible = &cvm.InternetAccessible{
		InternetMaxBandwidthOut: common.Int64Ptr(0),
		PublicIpAssigned:        common.BoolPtr(false),
	}
	request.InstanceCount = common.Int64Ptr(int64(count))
	request.InstanceName = common.StringPtr(fmt.Sprintf("spot-%s", time.Now().Format("20060102150405")))
	request.LoginSettings = &cvm.LoginSettings{
		KeyIds: common.StringPtrs(viper.GetStringSlice("qcloud.instance.auth.sshkey.ids")),
	}
	request.SecurityGroupIds = common.StringPtrs(viper.GetStringSlice("qcloud.instance.securitygroup.ids"))
	request.EnhancedService = &cvm.EnhancedService{
		SecurityService: &cvm.RunSecurityServiceEnabled{
			Enabled: common.BoolPtr(true),
		},
		MonitorService: &cvm.RunMonitorServiceEnabled{
			Enabled: common.BoolPtr(true),
		},
		AutomationService: &cvm.RunAutomationServiceEnabled{
			Enabled: common.BoolPtr(true),
		},
	}
	request.InstanceMarketOptions = &cvm.InstanceMarketOptionsRequest{
		SpotOptions: &cvm.SpotMarketOptions{
			MaxPrice: common.StringPtr("1000"),
		},
	}
	request.DisableApiTermination = common.BoolPtr(false)

	// 返回的resp是一个RunInstancesResponse的实例，与请求对象对应
	response, err := c.RunInstances(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("tencent api error has returned: %v", err)
	}
	if err != nil {
		return err
	}
	logrus.Debugf("%s", response.ToJsonString())
	return nil
}

func (c *Client) List() ([]Instance, error) {
	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := cvm.NewDescribeInstancesRequest()

	request.Filters = []*cvm.Filter{
		{
			Name:   common.StringPtr("instance-charge-type"),
			Values: common.StringPtrs([]string{"SPOTPAID"}),
		},
	}

	// 返回的resp是一个DescribeInstancesResponse的实例，与请求对象对应
	response, err := c.DescribeInstances(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return nil, fmt.Errorf("tencent api error has returned: %v", err)
	}
	if err != nil {
		return nil, err
	}
	var ins []Instance
	for _, i := range response.Response.InstanceSet {
		if strings.HasPrefix(*i.InstanceName, "spot") {
			ip := ""
			if len(i.PrivateIpAddresses) != 0 {
				ip = *i.PrivateIpAddresses[0]
			}
			ins = append(ins, Instance{
				CreatedTime:        *i.CreatedTime,
				InstanceId:         *i.InstanceId,
				InstanceName:       *i.InstanceName,
				InstanceType:       *i.InstanceType,
				InstanceChargeType: *i.InstanceChargeType,
				InstanceState:      *i.InstanceState,
				PrivateIpAddresses: ip,
			})
		}
	}
	return ins, nil
}

func (c *Client) Show() error {
	list, err := c.List()
	if err != nil {
		return err
	}
	table := uitable.New()
	table.AddRow("创建时间", "Name", "ID", "IP", "规格", "类型", "状态")
	for _, i := range list {
		table.AddRow(i.CreatedTime, i.InstanceName, i.InstanceId, i.PrivateIpAddresses, i.InstanceType, i.InstanceChargeType, i.InstanceState)
	}
	return output.EncodeTable(os.Stdout, table)
}

func (c *Client) Drop(ids []string) error {
	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := cvm.NewTerminateInstancesRequest()

	request.InstanceIds = common.StringPtrs(ids)

	// 返回的resp是一个TerminateInstancesResponse的实例，与请求对象对应
	_, err := c.TerminateInstances(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("tencent api error has returned: %v", err)
	}
	if err != nil {
		return err
	}
	return nil
}
