// Package cmd contains all CLI commands for the drugo tool.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	// Version is the current version of drugo CLI.
	Version = "0.1.0"
)

var rootCmd = &cobra.Command{
	Use:   "drugo",
	Short: "Drugo 是一个轻量级、模块化的 Go 应用程序框架 CLI 工具",
	Long: `Drugo CLI 是 Drugo 框架的命令行工具。
它帮助你创建新项目和模块，并提供标准的目录结构。

用法:
  drugo new <项目名称>           创建一个新的 Drugo 项目
  drugo new module <模块名称>    在现有项目中创建新模块

示例:
  drugo new myapp                创建一个名为 'myapp' 的新项目
  drugo new module user          创建一个带有 CRUD 模板的 user 模块`,
	Version: Version,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	// Add version template
	rootCmd.SetVersionTemplate(fmt.Sprintf("drugo version %s\n", Version))
}
