package cmd

import (
	"strings"

	"github.com/ysicing/spot/cloud/qcloud"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func cmdDestroy() *cobra.Command {
	var all bool
	c := &cobra.Command{
		Use:     "destroy",
		Aliases: []string{"down", "delete"},
		Short:   "销毁腾讯云竞价虚拟机",
		RunE: func(_ *cobra.Command, _ []string) error {
			client := qcloud.NewClient()
			vms, err := client.List()
			if err != nil {
				return err
			}

			okvms := []qcloud.Instance{}
			var ids []string
			for _, vm := range vms {
				if vm.InstanceState == "RUNNING" {
					ids = append(ids, vm.InstanceID)
					okvms = append(okvms, vm)
				}
			}
			if len(okvms) == 0 {
				logrus.Info("没有可销毁的虚拟机")
				return nil
			}
			if all {
				logrus.Infof("销毁所有虚拟机, 数目: %d", len(ids))
				for _, vm := range okvms {
					client.DeleteRecord(vm.PublicIPAddresses)
				}
				return client.Drop(ids)
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
			client.DeleteRecord(okvms[i].PublicIPAddresses)
			return client.Drop([]string{okvms[i].InstanceID})
		},
	}
	c.Flags().BoolVarP(&all, "all", "a", false, "销毁所有虚拟机")
	return c
}
