package cmd

import (
	"github.com/ysicing/spot/cloud/qcloud"

	"github.com/spf13/cobra"
)

func cmdNew() *cobra.Command {
	var count int64 
	c := &cobra.Command{
		Use:   "new",
		Aliases: []string{"up", "create"},
		Short: "新建腾讯云虚拟机",
		RunE: func(c *cobra.Command, args []string) error {
			client := qcloud.NewClient()
			return client.Create(count)
		},
	}
	c.Flags().Int64VarP(&count, "count", "c", 1, "虚拟机数量")
	return c
}
