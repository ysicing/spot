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
		RunE: func(c *cobra.Command, args []string) error {
			client := qcloud.NewClient()
			vms, err := client.List()
			if err != nil {
				return err
			}
			if len(vms) == 0 {
				logrus.Info("没有可销毁的虚拟机")
				return nil
			}
			if all {
				logrus.Info("销毁所有虚拟机")
				var ids []string
				for _, vm := range vms {
					ids = append(ids, vm.InstanceID)
				}
				return client.Drop(ids)
			}
			prompt := promptui.Select{
				Label: "选择虚拟机",
				Items: vms,
				Templates: &promptui.SelectTemplates{
					Label:    "{{ . }}?",
					Active:   "\U0001F449 {{ .PrivateIpAddresses | cyan }} ({{ .InstanceName | red }})",
					Inactive: "  {{ .PrivateIpAddresses | cyan }} ({{ .InstanceName | red }})",
					Selected: "\U0001F389 {{ .PrivateIpAddresses | green }}",
				},
				Size: 4,
				Searcher: func(input string, index int) bool {
					vm := vms[index]
					name := vm.PrivateIPAddresses
					return strings.Contains(name, input)
				},
			}

			i, _, err := prompt.Run()
			if err != nil {
				return err
			}

			return client.Drop([]string{vms[i].InstanceID})
		},
	}
	c.Flags().BoolVarP(&all, "all", "a", false, "销毁所有虚拟机")
	return c
}
