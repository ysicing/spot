package cmd

import (
	"github.com/ysicing/spot/cloud/qcloud"

	"github.com/spf13/cobra"
)

func cmdNewArm() *cobra.Command {
	var count, exp int64
	var netaccess bool
	var image string
	c := &cobra.Command{
		Use:     "arm",
		Short:   "新建腾讯云ARM虚拟机",
		Version: "0.1.0",
		RunE: func(c *cobra.Command, args []string) error {
			client := qcloud.NewClient("ap-guangzhou")
			return client.CreateArm(count, exp, netaccess, image)
		},
	}
	c.Flags().Int64VarP(&count, "count", "c", 1, "虚拟机数量")
	c.Flags().BoolVar(&netaccess, "net", true, "是否开启公网访问, 单节点生效")
	c.Flags().StringVarP(&image, "image", "i", "", "指定linux arm镜像")
	c.Flags().Int64VarP(&exp, "exp", "e", 2, "销毁时间(区间1-12小时)")
	return c
}
