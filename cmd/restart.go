package cmd

import (
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ysicing/spot/cloud/qcloud"
)

func cmdRestart() *cobra.Command {
	c := &cobra.Command{
		Use:     "restart",
		Aliases: []string{"rs"},
		Short:   "重启腾讯云竞价虚拟机",
		RunE: func(c *cobra.Command, args []string) error {
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
			okvms = append(okvms, qcloud.Instance{
				InstanceID:         "ins-4yw9saqt",
				InstanceName:       "name",
				PrivateIPAddresses: "pip",
				InstanceState:      "RUNNING",
			})
			if len(okvms) == 0 {
				logrus.Info("没有可重启的虚拟机")
				return nil
			}
			templates := &promptui.SelectTemplates{
				Label:    "{{ . }}",
				Active:   "\U0001F449 {{ .PrivateIPAddresses | cyan }} ({{ .InstanceName | red }})",
				Inactive: "  {{ .PrivateIPAddresses | cyan }} ({{ .InstanceName | red }})",
				Selected: "\U0001F389 {{ .PrivateIPAddresses | green }}",
			}
			prompt := promptui.Select{
				Label:     "选择虚拟机",
				Items:     okvms,
				Templates: templates,
			}

			i, _, err := prompt.Run()
			if err != nil {
				return err
			}
			return client.Restart(okvms[i].InstanceID)
		},
	}
	return c
}
