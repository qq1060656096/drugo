package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/qq1060656096/drugo/cmd/drugo/internal/tpl"
	"github.com/qq1060656096/drugo/pkg/gomod"
	"github.com/spf13/cobra"
)

var moduleCmd = &cobra.Command{
	Use:   "module",
	Short: "模块管理",
	Long:  `管理 Drugo 项目中的模块。`,
}

var moduleNewCmd = &cobra.Command{
	Use:   "new <模块名称>",
	Short: "在当前 Drugo 项目中创建新模块",
	Long: `在当前项目中创建具有标准 CRUD 结构的新模块。

模块将在 internal/<模块名称>/ 目录中创建，包含:
  - api/       HTTP 处理器和路由注册
  - biz/       业务逻辑和领域实体
  - data/      数据访问层（仓储实现）
  - service/   服务层（DTO 和编排）

此命令必须在 Drugo 项目根目录（go.mod 所在位置）运行。`,
	Example: `  drugo module new user
  drugo module new order
  drugo module new product`,
	Args: cobra.ExactArgs(1),
	RunE: runNewModule,
}

func init() {
	rootCmd.AddCommand(moduleCmd)
	moduleCmd.AddCommand(moduleNewCmd)
}

func runNewModule(cmd *cobra.Command, args []string) error {
	moduleName := strings.ToLower(args[0])

	// Validate module name
	if err := validateModuleName(moduleName); err != nil {
		return err
	}

	// Find project root (where go.mod exists)
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取工作目录失败: %w", err)
	}

	projectRoot := gomod.ProjectRoot(wd)
	if projectRoot == "" {
		return fmt.Errorf("不在 %s 目录中，请在 Drugo 项目根目录运行", wd)
	}

	// Get module path from go.mod
	modPath, err := gomod.ModuleName(projectRoot)
	if err != nil {
		return fmt.Errorf("读取 go.mod 失败: %w", err)
	}

	// Check if module already exists
	modulePath := filepath.Join(projectRoot, "internal", moduleName)
	if _, err := os.Stat(modulePath); err == nil {
		return fmt.Errorf("模块 %q 已存在于 %s", moduleName, modulePath)
	}

	fmt.Printf("正在 %s 中创建模块 %q...\n", projectRoot, moduleName)

	// Create module structure
	if err := createModule(projectRoot, modPath, moduleName); err != nil {
		// Clean up on failure
		os.RemoveAll(modulePath)
		return fmt.Errorf("创建模块失败: %w", err)
	}

	fmt.Printf(`
模块 %q 创建成功！

结构:
  internal/%s/
  ├── api/
  │   └── %s.go      # HTTP 处理器和路由
  ├── biz/
  │   └── %s.go      # 业务逻辑
  ├── data/
  │   └── %s.go      # 数据仓储
  └── service/
      └── %s.go      # 服务层

下一步:
  1. 在 cmd/app/main.go 中导入模块:
     import _ "%s/internal/%s/api"
  2. 根据需要自定义生成的代码。

`, moduleName, moduleName, moduleName, moduleName, moduleName, moduleName, modPath, moduleName)

	return nil
}

func validateModuleName(name string) error {
	if name == "" {
		return fmt.Errorf("模块名称不能为空")
	}
	if strings.ContainsAny(name, " \t\n/\\.-") {
		return fmt.Errorf("模块名称不能包含空格、点号、连字符或路径分隔符")
	}
	// Check if starts with a letter
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return fmt.Errorf("模块名称只能包含字母和数字")
		}
	}
	return nil
}

func createModule(projectRoot, modPath, moduleName string) error {
	data := ModuleData{
		Name:      moduleName,
		NameTitle: toTitle(moduleName),
		ModPath:   modPath,
	}

	basePath := filepath.Join(projectRoot, "internal", moduleName)

	// Create directories
	dirs := []string{
		filepath.Join(basePath, "api"),
		filepath.Join(basePath, "biz"),
		filepath.Join(basePath, "data"),
		filepath.Join(basePath, "service"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录 %q 失败: %w", dir, err)
		}
	}

	// Create files from templates
	files := map[string]string{
		filepath.Join(basePath, "api", moduleName+".go"):     tpl.ModuleAPITpl,
		filepath.Join(basePath, "biz", moduleName+".go"):     tpl.ModuleBizTpl,
		filepath.Join(basePath, "data", moduleName+".go"):    tpl.ModuleDataTpl,
		filepath.Join(basePath, "service", moduleName+".go"): tpl.ModuleServiceTpl,
	}

	for path, tplContent := range files {
		if err := createModuleFileFromTemplate(path, tplContent, data); err != nil {
			return err
		}
	}

	return nil
}

func createModuleFileFromTemplate(path, tplContent string, data ModuleData) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件 %q 失败: %w", path, err)
	}
	defer f.Close()

	tpl, err := template.New(filepath.Base(path)).Parse(tplContent)
	if err != nil {
		return fmt.Errorf("解析模板 %q 失败: %w", path, err)
	}

	if err := tpl.Execute(f, data); err != nil {
		return fmt.Errorf("执行模板 %q 失败: %w", path, err)
	}

	return nil
}

// ModuleData holds data for module templates.
type ModuleData struct {
	Name      string // lowercase module name (e.g., "user")
	NameTitle string // title case module name (e.g., "User")
	ModPath   string // go module path (e.g., "github.com/myorg/myapp")
}

// toTitle converts a string to title case (first letter uppercase).
func toTitle(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
