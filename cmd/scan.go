package cmd

import (
	"github.com/spf13/cobra"
	"github.com/ysicing/spot/cloud/qcloud"
)

func cmdScan() *cobra.Command {
	var uuid string
	scan := &cobra.Command{
		Use:   "scan",
		Short: "扫描虚拟机镜像漏洞",
		RunE: func(c *cobra.Command, args []string) error {
			client := qcloud.NewClient()
			return client.Scan(uuid)
		},
	}
	scan.Flags().StringVarP(&uuid, "uuid", "i", "", "主机安全UUID")
	return scan
}
