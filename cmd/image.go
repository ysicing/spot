package cmd

import (
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ysicing/spot/cloud/qcloud"
)

type Action struct {
	Key   string
	Value string
}

func cmdImage() *cobra.Command {
	i := &cobra.Command{
		Use:     "image",
		Aliases: []string{"i"},
		Short:   "管理腾讯云镜像",
	}
	i.AddCommand(cmdImageList())
	i.AddCommand(cmdImagManage())
	return i
}

func cmdImageList() *cobra.Command {
	var notPublic bool
	c := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "show"},
		Short:   "列出腾讯云镜像",
		RunE: func(_ *cobra.Command, _ []string) error {
			client := qcloud.NewClient()
			return client.ImageShow(notPublic)
		},
	}
	c.Flags().BoolVar(&notPublic, "skip-public", true, "忽略官方镜像")
	return c
}

func cmdImagManage() *cobra.Command {
	var notPublic, netaccess bool
	action := []Action{{Key: "创建虚拟机", Value: "create"}, {Key: "删除镜像", Value: "delete"}}

	c := &cobra.Command{
		Use:     "manage",
		Aliases: []string{"op"},
		Short:   "管理腾讯云镜像\t 启动虚拟机 \t 删除镜像",
		RunE: func(_ *cobra.Command, _ []string) error {
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
					Selected: "\U0001F389 管理镜像: {{ .ImageID | green }}",
				},
				Size: 5,
			}
			i, _, err := prompt.Run()
			if err != nil {
				return err
			}
			if images[i].ImageType != "官方" {
				actionPrompt := promptui.Select{
					Label: "操作",
					Items: action,
					Templates: &promptui.SelectTemplates{
						Label:    "{{ . }}",
						Active:   "\U0001F449 {{ .Key | cyan }}",
						Inactive: " {{ .Key }}",
						Selected: "\U0001F389 {{ .Key | green }}",
					},
				}
				a, _, err := actionPrompt.Run()
				if err != nil {
					return err
				}
				if action[a].Value == "delete" {
					logrus.Infof("删除镜像 %s", images[i].ImageID)
					return client.ImageDrop([]string{images[i].ImageID})
				}
			}

			logrus.Infof("使用镜像 %s 创建竞价机器", images[i].ImageID)
			return client.Create(1, netaccess, false, images[i].ImageID)
		},
	}
	c.Flags().BoolVar(&notPublic, "skip-public", true, "忽略官方镜像")
	c.Flags().BoolVar(&netaccess, "net", true, "是否开启公网访问, 单节点生效")
	return c
}
