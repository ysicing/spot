package cmd

import (
	"github.com/ysicing/spot/cloud/qcloud"

	"github.com/spf13/cobra"
)

func cmdList() *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Aliases: []string{"ls", "show"},
		Short: "列出腾讯云竞价虚拟机",
		RunE: func(c *cobra.Command, args []string) error {
			client := qcloud.NewClient()
			return client.Show()
		},
	}
	return c
}
