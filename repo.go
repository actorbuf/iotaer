package main

import (
	"fmt"
	"github.com/actorbuf/builder/toolkit"
	"github.com/elliotchance/pie/pie"
	"github.com/spf13/cobra"
	"os"
)

// 仓库配置
var (
	repoCommand = cobra.Command{
		Use:   "repo",
		Short: "gitlab仓库操作",
	}

	// 60 黄志浩
	// 111 徐业
	whiteList = []int{60, 111} // 操作白名单
)

func addRepoCommand() *cobra.Command {
	repoCommand.AddCommand(gitlabShowGroups())       // 显示所有组列表

	return &repoCommand
}

// gitlabShowGroups 显示所有组列表
func gitlabShowGroups() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups",
		Short: "输出gitlab的组列表",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := getBuilderConfig()
			if cfg == nil {
				_, _ = fmt.Fprintf(os.Stderr, "获取配置信息失败\n")
				return
			}

			if !pie.Ints(whiteList).Contains(int(cfg.UserID)) {
				_, _ = fmt.Fprintf(os.Stderr, "无操作权限\n")
				return
			}

			list, err := toolkit.GetGroupList()
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "获取gitlab组列表失败: %+v\n", err)
				return
			}

			for _, info := range *list {
				_, _ = fmt.Fprintf(os.Stdout, "组名: %-25s组ID: %d\n", info.Name, info.ID)
			}
		},
	}

	return cmd
}

