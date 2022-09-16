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
	cwp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cwp/v20180228"
)

type Client struct {
	cvmCliet  *cvm.Client
	cwpClient *cwp.Client
}

type Instance struct {
	UUID               string
	CreatedTime        string
	InstanceName       string
	InstanceID         string
	InstanceType       string
	InstanceChargeType string
	InstanceState      string
	PrivateIPAddresses string
	PublicIPAddresses  string
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
	cvmcpf := profile.NewClientProfile()
	cvmcpf.HttpProfile.Endpoint = "cvm.tencentcloudapi.com"
	// 实例化要请求产品的client对象,clientProfile是可选的
	cvmClient, _ := cvm.NewClient(credential, viper.GetString("qcloud.region"), cvmcpf)
	cwpcpf := profile.NewClientProfile()
	cwpcpf.HttpProfile.Endpoint = "cwp.tencentcloudapi.com"
	cwpClient, _ := cwp.NewClient(credential, viper.GetString("qcloud.region"), cwpcpf)
	return &Client{cvmCliet: cvmClient, cwpClient: cwpClient}
}

func (c *Client) Create(count int64, netaccess, windows bool) error {

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
	disk := viper.GetInt64("qcloud.instance.disk")
	if disk == 0 {
		disk = 50
	}
	if windows && disk < 100 {
		disk = 100
	}
	request.SystemDisk = &cvm.SystemDisk{
		DiskType: common.StringPtr("CLOUD_PREMIUM"),
		DiskSize: common.Int64Ptr(disk),
	}
	request.VirtualPrivateCloud = &cvm.VirtualPrivateCloud{
		VpcId:            common.StringPtr(viper.GetString("qcloud.instance.network.vpc.id")),
		SubnetId:         common.StringPtr(viper.GetString("qcloud.instance.network.subnet.id")),
		AsVpcGateway:     common.BoolPtr(false),
		Ipv6AddressCount: common.Uint64Ptr(0),
	}
	if count == 1 && netaccess {
		request.InternetAccessible = &cvm.InternetAccessible{
			InternetChargeType:      common.StringPtr("TRAFFIC_POSTPAID_BY_HOUR"),
			InternetMaxBandwidthOut: common.Int64Ptr(100),
			PublicIpAssigned:        common.BoolPtr(true),
		}
	} else {
		request.InternetAccessible = &cvm.InternetAccessible{
			InternetMaxBandwidthOut: common.Int64Ptr(0),
			PublicIpAssigned:        common.BoolPtr(false),
		}
	}
	request.InstanceCount = common.Int64Ptr(int64(count))
	namePrefix := "spot"
	if windows {
		request.LoginSettings = &cvm.LoginSettings{
			KeepImageLogin: common.StringPtr("true"),
		}
		namePrefix = "spot-windows"
	} else {
		request.LoginSettings = &cvm.LoginSettings{
			KeyIds: common.StringPtrs(viper.GetStringSlice("qcloud.instance.auth.sshkey.ids")),
		}
	}
	request.InstanceName = common.StringPtr(fmt.Sprintf("%s-%s", namePrefix, time.Now().Format("20060102150405")))
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
	response, err := c.cvmCliet.RunInstances(request)
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
	response, err := c.cvmCliet.DescribeInstances(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return nil, fmt.Errorf("tencent api error has returned: %v", err)
	}
	if err != nil {
		return nil, err
	}
	var ins []Instance
	for _, i := range response.Response.InstanceSet {
		if strings.HasPrefix(*i.InstanceName, "spot") {
			ip := "-"
			if len(i.PrivateIpAddresses) != 0 {
				ip = *i.PrivateIpAddresses[0]
			}
			eip := "-"
			if len(i.PublicIpAddresses) != 0 {
				eip = *i.PublicIpAddresses[0]
			}
			ins = append(ins, Instance{
				CreatedTime:        *i.CreatedTime,
				InstanceID:         *i.InstanceId,
				InstanceName:       *i.InstanceName,
				InstanceType:       *i.InstanceType,
				InstanceChargeType: *i.InstanceChargeType,
				InstanceState:      *i.InstanceState,
				PrivateIPAddresses: ip,
				PublicIPAddresses:  eip,
				UUID:               *i.Uuid,
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
	table.AddRow("创建时间", "Name", "ID", "内网IP", "公网IP", "规格", "类型", "状态")
	for _, i := range list {
		table.AddRow(i.CreatedTime, i.InstanceName, i.InstanceID, i.PrivateIPAddresses, i.PublicIPAddresses, i.InstanceType, i.InstanceChargeType, i.InstanceState)
	}
	return output.EncodeTable(os.Stdout, table)
}

func (c *Client) Drop(ids []string) error {
	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := cvm.NewTerminateInstancesRequest()

	request.InstanceIds = common.StringPtrs(ids)

	// 返回的resp是一个TerminateInstancesResponse的实例，与请求对象对应
	_, err := c.cvmCliet.TerminateInstances(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("tencent api error has returned: %v", err)
	}
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Restart(id string) error {
	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := cvm.NewRebootInstancesRequest()

	request.InstanceIds = common.StringPtrs([]string{id})
	// SOFT 表示软关机
	// HARD 表示硬关机
	// SOFT_FIRST 表示优先软关机，失败再执行硬关机
	request.StopType = common.StringPtr("SOFT_FIRST")

	// 返回的resp是一个RebootInstancesResponse的实例，与请求对象对应
	_, err := c.cvmCliet.RebootInstances(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("tencent api error has returned: %v", err)
	}
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Scan(id string) error {
	request := cwp.NewScanVulRequest()
	request.VulLevels = common.StringPtr("1;2;3;4")
	request.HostType = common.Uint64Ptr(2)
	request.VulCategories = common.StringPtr("1;2;4")
	request.QuuidList = common.StringPtrs([]string{id})
	request.TimeoutPeriod = common.Uint64Ptr(3600)
	response, err := c.cwpClient.ScanVul(request)
	if terr, ok := err.(*errors.TencentCloudSDKError); ok {
		if terr.Code == "OperationDenied" {
			logrus.Warnf("%s %s", id, terr.Message)
			return nil
		}
		return fmt.Errorf("tencent api error has returned: %v", err)
	}
	if err != nil {
		return err
	}
	logrus.Infof("Scan %s task create %d", id, response.Response.TaskId)
	return nil
}
