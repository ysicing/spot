package cmd

import (
	"errors"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ysicing/spot/cloud/qcloud"
)

func cmdDnspod() *cobra.Command {
	c := &cobra.Command{
		Use:     "dnspod",
		Short:   "虚拟机创建解析记录",
		Version: "0.3.0",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			domain := viper.GetString("qcloud.dnspod.main")
			sub := viper.GetString("qcloud.dnspod.sub")
			if len(domain) == 0 || len(sub) == 0 {
				return errors.New("请配置qcloud.dnspod.main和qcloud.dnspod.sub")
			}
			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			client := qcloud.NewClient()
			vms, err := client.List()
			if err != nil {
				return err
			}

			okvms := []qcloud.Instance{}

			for _, vm := range vms {
				if vm.InstanceState == "RUNNING" {
					okvms = append(okvms, vm)
				}
			}
			if len(okvms) == 0 {
				logrus.Info("没有可用虚拟机")
				return nil
			}
			prompt := promptui.Select{
				Label: "选择虚拟机",
				Items: okvms,
				Templates: &promptui.SelectTemplates{
					Label:    "{{ . }}?",
					Active:   "\U0001F449 {{ .PrivateIPAddresses | cyan }} ({{ .InstanceName | red }})",
					Inactive: "  {{ .PrivateIPAddresses | cyan }} ({{ .InstanceName | red }})",
					Selected: "\U0001F389 {{ .PrivateIPAddresses | green }}",
				},
				Size: 4,
				Searcher: func(input string, index int) bool {
					vm := okvms[index]
					name := vm.PrivateIPAddresses
					return strings.Contains(name, input)
				},
			}

			i, _, err := prompt.Run()
			if err != nil {
				return err
			}
			return client.CreateOrUpdateRecord(okvms[i].PublicIPAddresses)
		},
	}
	return c
}
