package qcloud

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

type Client struct {
	cvmCliet *cvm.Client
	dnsCliet *dnspod.Client
	domain   string
	sub      string
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

	dnscpf := profile.NewClientProfile()
	dnscpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"
	dnsClient, _ := dnspod.NewClient(credential, "", dnscpf)
	domain := viper.GetString("qcloud.dnspod.main")
	sub := viper.GetString("qcloud.dnspod.sub")
	return &Client{cvmCliet: cvmClient, dnsCliet: dnsClient, domain: domain, sub: sub}
}
