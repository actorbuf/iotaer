package main

import (
	"github.com/spf13/cobra"
)

func toolBuilderCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tool",
		Short: "将文件构建为任意工具箱",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	return cmd
}
