package qcloud

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

func (d *Client) CreateOrUpdateRecord(ip string) error {
	if len(d.domain) == 0 || len(d.sub) == 0 {
		return nil
	}
	recordID, err := d.searchRecord()
	if err != nil {
		return err
	}
	if recordID == nil {
		return d.createRecord(ip)
	}
	return d.updateRecord(recordID, ip)
}

func (d *Client) DeleteRecord(ip string) error {
	if len(d.domain) == 0 || len(d.sub) == 0 {
		return nil
	}
	recordID, err := d.searchRecord()
	if err != nil {
		return err
	}
	if recordID == nil {
		return nil
	}
	request := dnspod.NewDeleteRecordRequest()
	request.Domain = common.StringPtr(d.domain)
	request.RecordId = recordID
	_, err = d.dnsCliet.DeleteRecord(request)
	if err != nil {
		return err
	}
	logrus.Infof("delete record success %s.%s ---> %s", d.sub, d.domain, ip)
	return nil
}

func (d *Client) searchRecord() (*uint64, error) {
	request := dnspod.NewDescribeRecordListRequest()
	request.Domain = common.StringPtr(d.domain)
	request.Subdomain = common.StringPtr(d.sub)
	request.RecordType = common.StringPtr("A")
	response, err := d.dnsCliet.DescribeRecordList(request)
	if err != nil {
		return nil, err
	}
	if len(response.Response.RecordList) == 0 {
		return nil, nil
	}
	if len(response.Response.RecordList) > 1 {
		logrus.Warnf("find more than one record %s.%s", d.sub, d.domain)
		return nil, errors.New("find more than one record")
	}
	for _, record := range response.Response.RecordList {
		return record.RecordId, nil
	}
	return nil, nil
}

func (d *Client) createRecord(ip string) error {
	request := dnspod.NewCreateRecordRequest()
	request.Domain = common.StringPtr(d.domain)
	request.SubDomain = common.StringPtr(d.sub)
	request.RecordType = common.StringPtr("A")
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(ip)
	request.Remark = common.StringPtr(ip)
	_, err := d.dnsCliet.CreateRecord(request)
	if err != nil {
		return err
	}
	logrus.Infof("create record success %s.%s ---> %s", d.sub, d.domain, ip)
	return nil
}

func (d *Client) updateRecord(id *uint64, ip string) error {
	request := dnspod.NewModifyRecordRequest()
	request.Domain = common.StringPtr(d.domain)
	request.SubDomain = common.StringPtr(d.sub)
	request.RecordType = common.StringPtr("A")
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(ip)
	request.RecordId = id
	_, err := d.dnsCliet.ModifyRecord(request)
	if err != nil {
		return err
	}
	logrus.Infof("update record success %s.%s ---> %s", d.sub, d.domain, ip)
	return nil
}
