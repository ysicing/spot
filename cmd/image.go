package cmd

import (
	"github.com/manifoldco/promptui"
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
	i.AddCommand(cmdImagDeploy())
	return i
}

func cmdImageList() *cobra.Command {
	var notPublic bool
	c := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "show"},
		Short:   "列出腾讯云镜像",
		RunE: func(c *cobra.Command, args []string) error {
			client := qcloud.NewClient()
			return client.ImageShow(notPublic)
		},
	}
	c.Flags().BoolVar(&notPublic, "skip-public", true, "忽略官方镜像")
	return c
}

func cmdImagDeploy() *cobra.Command {
	var notPublic, netaccess bool
	c := &cobra.Command{
		Use:     "deploy",
		Aliases: []string{"run", "show"},
		Short:   "选择腾讯云镜像起虚拟机",
		RunE: func(c *cobra.Command, args []string) error {
			client := qcloud.NewClient()
			images, err := client.ImageList(notPublic)
			if err != nil {
				return err
			}
			prompt := promptui.Select{
				Label: "选择镜像",
				Items: images,
				Templates: &promptui.SelectTemplates{
					Label:    "{{ . }}?",
					Active:   "\U0001F449 {{ .ImageID | cyan }} ({{ .ImageName | red }})",
					Inactive: "  {{ .ImageID | cyan }} ({{ .ImageName | red }})",
					Selected: "\U0001F389 选择 {{ .ImageID | green }} 创建虚拟机",
				},
				Size: 5,
			}
			i, _, err := prompt.Run()
			if err != nil {
				return err
			}
			return client.Create(1, netaccess, false, images[i].ImageID)
		},
	}
	c.Flags().BoolVar(&notPublic, "skip-public", true, "忽略官方镜像")
	c.Flags().BoolVar(&netaccess, "net", true, "是否开启公网访问, 单节点生效")
	return c
}
