package cmd

import (
	"github.com/spf13/cobra"
	"github.com/ysicing/spot/cloud/qcloud"
)

func cmdImage() *cobra.Command {
	i := &cobra.Command{
		Use:     "image",
		Aliases: []string{"i"},
		Short:   "管理腾讯云镜像",
	}
	i.AddCommand(cmdImageList())
	return i
}

func cmdImageList() *cobra.Command {
	c := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "show"},
		Short:   "列出腾讯云镜像",
		RunE: func(c *cobra.Command, args []string) error {
			client := qcloud.NewClient()
			return client.ImageList()
		},
	}
	return c
}
