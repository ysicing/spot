package qcloud

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ergoapi/util/color"
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
	cvmCliet *cvm.Client
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

type Image struct {
	ImageID          string
	ImageName        string
	ImageState       string
	ImageType        string
	ImageDescription string
	OsName           string
}

func NewClient(regions ...string) *Client {
	// 实例化一个认证对象，入参需要传入腾讯云账户secretId，secretKey,此处还需注意密钥对的保密
	// 密钥可前往https://console.cloud.tencent.com/cam/capi网站进行获取
	credential := common.NewCredential(
		viper.GetString("qcloud.account.id"),
		viper.GetString("qcloud.account.secret"),
	)
	logrus.Debugf("credential: %s, %s", credential.SecretId, credential.SecretKey)
	region := viper.GetString("qcloud.region")
	if len(regions) > 0 {
		region = regions[0]
	}
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cvmcpf := profile.NewClientProfile()
	cvmcpf.HttpProfile.Endpoint = "cvm.tencentcloudapi.com"
	cvmClient, _ := cvm.NewClient(credential, region, cvmcpf)
	return &Client{cvmCliet: cvmClient}
}

func (c *Client) Create(count int64, netaccess, windows bool, image string) error {

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
	defaultImage := viper.GetString("qcloud.instance.image")
	if len(image) == 0 || windows {
		image = defaultImage
	}
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
	request.ImageId = common.StringPtr(image)
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
	tags := viper.GetStringSlice("qcloud.tags")
	if len(tags) > 0 {
		var ts []*cvm.Tag
		for _, t := range tags {
			tt := strings.Split(t, "::")
			if len(tt) == 2 {
				ts = append(ts, &cvm.Tag{
					Key:   common.StringPtr(tt[0]),
					Value: common.StringPtr(tt[1]),
				})
			}
		}
		request.TagSpecification = []*cvm.TagSpecification{
			{
				ResourceType: common.StringPtr("instance"),
				Tags:         ts,
			},
		}
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

func (c *Client) CreateArm(count, exp int64, netaccess bool, image string) error {

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := cvm.NewRunInstancesRequest()
	// arm 只支持按量
	request.InstanceChargeType = common.StringPtr("POSTPAID_BY_HOUR")
	qzone := viper.GetString("qcloud.zone")
	if !strings.HasPrefix(qzone, "ap-guangzhou-") {
		qzone = "ap-guangzhou-6"
	}
	request.Placement = &cvm.Placement{
		Zone: common.StringPtr(qzone),
	}
	if viper.GetInt64("qcloud.project.id") > 0 {
		request.Placement.ProjectId = common.Int64Ptr(viper.GetInt64("qcloud.project.id"))
	}
	intype := viper.GetString("qcloud.instance.type")
	if !strings.HasPrefix(intype, "SR") {
		intype = "SR1.MEDIUM2"
	}
	request.InstanceType = common.StringPtr(intype)
	disk := viper.GetInt64("qcloud.instance.disk")
	if disk == 0 {
		disk = 50
	}
	request.SystemDisk = &cvm.SystemDisk{
		DiskType: common.StringPtr("CLOUD_PREMIUM"),
		DiskSize: common.Int64Ptr(disk),
	}
	vpcid := viper.GetString("qcloud.instance.network.vpc.id")
	vpcsub := viper.GetString("qcloud.instance.network.subnet.id")
	if len(vpcid) > 4 && len(vpcsub) > 7 {
		request.VirtualPrivateCloud = &cvm.VirtualPrivateCloud{
			VpcId:            common.StringPtr(vpcid),
			SubnetId:         common.StringPtr(vpcsub),
			AsVpcGateway:     common.BoolPtr(false),
			Ipv6AddressCount: common.Uint64Ptr(0),
		}
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
	defaultImage := viper.GetString("qcloud.instance.image")
	if len(image) == 0 {
		image = defaultImage
	}
	namePrefix := "spot-arm"
	request.LoginSettings = &cvm.LoginSettings{
		KeyIds: common.StringPtrs(viper.GetStringSlice("qcloud.instance.auth.sshkey.ids")),
	}
	request.ImageId = common.StringPtr(image)
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
	tags := viper.GetStringSlice("qcloud.tags")
	if len(tags) > 0 {
		var ts []*cvm.Tag
		for _, t := range tags {
			tt := strings.Split(t, "::")
			if len(tt) == 2 {
				ts = append(ts, &cvm.Tag{
					Key:   common.StringPtr(tt[0]),
					Value: common.StringPtr(tt[1]),
				})
			}
		}
		request.TagSpecification = []*cvm.TagSpecification{
			{
				ResourceType: common.StringPtr("instance"),
				Tags:         ts,
			},
		}
	}
	request.InstanceMarketOptions = &cvm.InstanceMarketOptionsRequest{
		SpotOptions: &cvm.SpotMarketOptions{
			MaxPrice: common.StringPtr("1000"),
		},
	}
	request.DisableApiTermination = common.BoolPtr(false)
	if exp >= 12 {
		exp = 12
	}
	if exp <= 1 {
		exp = 1
	}
	request.ActionTimer = &cvm.ActionTimer{
		TimerAction: common.StringPtr("TerminateInstances"),
		ActionTime:  common.StringPtr(time.Now().Add(time.Hour * time.Duration(exp)).Format("2006-01-02 15:04:05")),
		Externals: &cvm.Externals{
			ReleaseAddress: common.BoolPtr(true),
		},
	}

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
	var ins []Instance
	sins, err := c.ListSPOTPAID()
	if err != nil {
		return nil, err
	}
	ins = append(ins, sins...)
	hins, err := c.ListPOSTPAIDBYHOUR()
	if err != nil {
		return nil, err
	}
	ins = append(ins, hins...)
	return ins, nil
}

func (c *Client) ListSPOTPAID() ([]Instance, error) {
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

func (c *Client) ListPOSTPAIDBYHOUR() ([]Instance, error) {
	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := cvm.NewDescribeInstancesRequest()

	request.Filters = []*cvm.Filter{
		{
			Name:   common.StringPtr("instance-charge-type"),
			Values: common.StringPtrs([]string{"POSTPAID_BY_HOUR"}),
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

func (c *Client) ImageList(notPublic bool) ([]Image, error) {
	request := cvm.NewDescribeImagesRequest()
	request.Offset = common.Uint64Ptr(0)
	request.Limit = common.Uint64Ptr(100)
	imageList := make([]Image, 0)
	totalCount := uint64(100)
	for *request.Offset < totalCount {
		response, err := c.cvmCliet.DescribeImages(request)
		if err != nil {
			return nil, err
		}
		if response.Response.ImageSet != nil && len(response.Response.ImageSet) > 0 {
			for _, i := range response.Response.ImageSet {
				if notPublic && *i.ImageType == "PUBLIC_IMAGE" {
					continue
				}
				imageType := ""
				imageState := color.SGreen("正常")
				if *i.ImageState != "NORMAL" {
					imageState = *i.ImageState
				}
				if *i.ImageType == "PUBLIC_IMAGE" {
					imageType = "官方"
				} else if *i.ImageType == "PRIVATE_IMAGE" {
					imageType = "自定义镜像"
				} else {
					imageType = "共享镜像"
				}
				imageList = append(imageList, Image{
					ImageID:          *i.ImageId,
					ImageName:        *i.ImageName,
					ImageState:       imageState,
					ImageType:        imageType,
					ImageDescription: *i.ImageDescription,
					OsName:           *i.OsName,
				})
			}
		}
		totalCount = uint64(*response.Response.TotalCount)
		request.Offset = common.Uint64Ptr(*request.Offset + uint64(len(response.Response.ImageSet)))
	}
	return imageList, nil
}

func (c *Client) ImageShow(notPublic bool) error {
	images, err := c.ImageList(notPublic)
	if err != nil {
		if _, ok := err.(*errors.TencentCloudSDKError); ok {
			return fmt.Errorf("tencent api error has returned: %v", err)
		}
		return err
	}
	table := uitable.New()
	table.AddRow("Name", "ID", "OS", "类型", "状态", "描述")

	for _, i := range images {
		table.AddRow(i.ImageName, i.ImageID, i.OsName, i.ImageType, i.ImageState, i.ImageDescription)
	}
	return output.EncodeTable(os.Stdout, table)
}

func (c *Client) ImageDrop(ids []string) error {
	request := cvm.NewDeleteImagesRequest()
	request.ImageIds = common.StringPtrs(ids)
	request.DeleteBindedSnap = common.BoolPtr(true)
	_, err := c.cvmCliet.DeleteImages(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("tencent api error has returned: %v", err)
	}
	if err != nil {
		return err
	}
	return nil
}
