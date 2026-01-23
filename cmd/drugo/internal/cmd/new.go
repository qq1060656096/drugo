package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/qq1060656096/drugo/cmd/drugo/internal/tpl"
	"github.com/spf13/cobra"
)

var (
	// Project flags
	projectModPath string
)

var newCmd = &cobra.Command{
	Use:   "new <项目名称>",
	Short: "创建一个新的 Drugo 项目",
	Long: `创建一个带有标准目录结构的新 Drugo 项目。

项目将在名为 <项目名称> 的新目录中创建。
如果未指定 --mod，将使用 "github.com/<项目名称>" 作为模块路径。

目录结构:
  <项目名称>/
  ├── cmd/
  │   └── app/
  │       └── main.go
  ├── conf/
  │   ├── gin.yaml
  │   └── log.yaml
  ├── internal/
  ├── runtime/
  │   └── logs/
  ├── go.mod
  ├── Makefile
  ├── .gitignore
  └── README.md`,
	Example: `  drugo new myapp
  drugo new myapp --mod github.com/myorg/myapp`,
	Args: cobra.ExactArgs(1),
	RunE: runNew,
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVarP(&projectModPath, "mod", "m", "", "go 模块路径 (默认: github.com/<项目名称>)")
}

func runNew(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// Validate project name
	if err := validateProjectName(projectName); err != nil {
		return err
	}

	// Set module path if not specified
	modPath := projectModPath
	if modPath == "" {
		modPath = fmt.Sprintf("%s", projectName)
	}

	// Check if directory exists
	if _, err := os.Stat(projectName); err == nil {
		return fmt.Errorf("目录 %q 已存在", projectName)
	}

	fmt.Printf("正在创建项目 %q，模块路径为 %q...\n", projectName, modPath)
	version := getVersion()
	// Create project structure
	if err := createProject(projectName, modPath, version); err != nil {
		// Clean up on failure
		os.RemoveAll(projectName)
		return fmt.Errorf("创建项目失败: %w", err)
	}

	fmt.Printf(`
项目 %q 创建成功！

下一步:
  cd %s
  go mod tidy
  make run

`, projectName, projectName)

	return nil
}

func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("项目名称不能为空")
	}
	if strings.ContainsAny(name, " \t\n/\\") {
		return fmt.Errorf("项目名称不能包含空格或路径分隔符")
	}
	return nil
}

func createProject(name, modPath, version string) error {
	data := ProjectData{
		Name:    name,
		ModPath: modPath,
		Version: version,
	}

	// Create directories
	dirs := []string{
		filepath.Join(name, "cmd", "app"),
		filepath.Join(name, "conf"),
		filepath.Join(name, "internal"),
		filepath.Join(name, "runtime", "logs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录 %q 失败: %w", dir, err)
		}
	}

	// Create files from templates
	files := map[string]string{
		filepath.Join(name, "cmd", "app", "main.go"):       tpl.MainGoTpl,
		filepath.Join(name, "conf", "gin.yaml"):            tpl.GinYamlTpl,
		filepath.Join(name, "conf", "log.yaml"):            tpl.LogYamlTpl,
		filepath.Join(name, "conf", "db.yaml"):             tpl.DbYamlTpl,
		filepath.Join(name, "conf", "redis.yaml"):          tpl.RedisYamlTpl,
		filepath.Join(name, "go.mod"):                      tpl.GoModTpl,
		filepath.Join(name, "Makefile"):                    tpl.MakefileTpl,
		filepath.Join(name, ".gitignore"):                  tpl.GitignoreTpl,
		filepath.Join(name, "README.md"):                   tpl.ReadmeTpl,
		filepath.Join(name, ".air.toml"):                   tpl.AirTomlTpl,
		filepath.Join(name, "runtime", "logs", ".gitkeep"): "",
	}

	for path, tplContent := range files {
		if err := createFileFromTemplate(path, tplContent, data); err != nil {
			return err
		}
	}

	return nil
}

func createFileFromTemplate(path, tplContent string, data any) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件 %q 失败: %w", path, err)
	}
	defer f.Close()

	if tplContent == "" {
		return nil
	}

	tpl, err := template.New(filepath.Base(path)).Parse(tplContent)
	if err != nil {
		return fmt.Errorf("解析模板 %q 失败: %w", path, err)
	}

	if err := tpl.Execute(f, data); err != nil {
		return fmt.Errorf("执行模板 %q 失败: %w", path, err)
	}

	return nil
}

// ProjectData holds data for project templates.
type ProjectData struct {
	Name    string
	ModPath string
	Version string
}
