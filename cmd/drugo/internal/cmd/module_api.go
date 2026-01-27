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

var moduleApiCmd = &cobra.Command{
	Use:   "new-api <模块名称> <API名称>",
	Short: "在现有模块中创建新的 API 结构",
	Long: `在现有模块中快速生成符合项目结构的 API 相关文件。

命令格式：drugo module new-api <module_name> <api_name>

参数说明：
  <module_name>: 指定目标模块的名称（例如：goods）。该模块必须已存在于项目中。
  <api_name>:    指定要创建的新 API 名称（例如：category）。

生成的文件结构：
  internal/<module_name>/
  ├── api/
  │   └── <api_name>.go    # API 层定义
  ├── biz/
  │   └── <api_name>.go    # 业务逻辑层
  ├── data/
  │   └── <api_name>.go    # 数据访问层
  └── service/
  │   └── <api_name>.go    # 服务层`,
	Example: `  drugo module new-api goods category
  drugo module new-api user address`,
	Args: cobra.ExactArgs(2),
	RunE: runNewModuleApi,
}

func init() {
	moduleCmd.AddCommand(moduleApiCmd)
}

func runNewModuleApi(cmd *cobra.Command, args []string) error {
	moduleName := strings.ToLower(args[0])
	apiName := strings.ToLower(args[1])

	// Validate names
	if err := validateName(moduleName, "模块名称"); err != nil {
		return err
	}
	if err := validateName(apiName, "API名称"); err != nil {
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

	// Check if module exists
	moduleBasePath := filepath.Join(projectRoot, "internal", moduleName)
	if _, err := os.Stat(moduleBasePath); os.IsNotExist(err) {
		return fmt.Errorf("模块 %q 不存在于 %s，请先使用 'drugo module new %s' 创建模块", moduleName, moduleBasePath, moduleName)
	}

	fmt.Printf("正在模块 %q 中创建 API %q...\n", moduleName, apiName)

	// Create API structure
	if err := createModuleApi(projectRoot, modPath, moduleName, apiName); err != nil {
		return fmt.Errorf("创建 API 失败: %w", err)
	}

	fmt.Printf(`
API %q 创建成功！

结构:
  internal/%s/
  ├── api/
  │   └── %s.go
  ├── biz/
  │   └── %s.go
  ├── data/
  │   └── %s.go
  └── service/
      └── %s.go

`, apiName, moduleName, apiName, apiName, apiName, apiName)

	return nil
}

func validateName(name, field string) error {
	if name == "" {
		return fmt.Errorf("%s不能为空", field)
	}
	if strings.ContainsAny(name, " \t\n/\\.-") {
		return fmt.Errorf("%s不能包含空格、点号、连字符或路径分隔符", field)
	}
	// Check if starts with a letter
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return fmt.Errorf("%s只能包含字母和数字", field)
		}
	}
	return nil
}

func createModuleApi(projectRoot, modPath, moduleName, apiName string) error {
	data := ModuleApiData{
		Name:       apiName,
		NameTitle:  toTitle(apiName),
		ModuleName: moduleName,
		ModPath:    modPath,
	}

	basePath := filepath.Join(projectRoot, "internal", moduleName)

	// Check and create files
	files := map[string]string{
		filepath.Join(basePath, "api", apiName+".go"):     tpl.ModuleApiApiTpl,
		filepath.Join(basePath, "biz", apiName+".go"):     tpl.ModuleApiBizTpl,
		filepath.Join(basePath, "data", apiName+".go"):    tpl.ModuleApiDataTpl,
		filepath.Join(basePath, "service", apiName+".go"): tpl.ModuleApiServiceTpl,
	}

	// First check if any file exists
	for path := range files {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("文件 %q 已存在，请先删除或使用不同名称", path)
		}
	}

	// Ensure directories exist (they should, but just in case)
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

	for path, tplContent := range files {
		if err := createModuleApiFileFromTemplate(path, tplContent, data); err != nil {
			// If one fails, we stop. We don't rollback previous files to avoid deleting user data if they partially existed (though we checked existence before).
			// Since we checked existence, rollback might be safe, but let's keep it simple.
			return err
		}
		fmt.Printf("创建文件: %s\n", path)
	}

	return nil
}

func createModuleApiFileFromTemplate(path, tplContent string, data ModuleApiData) error {
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

// ModuleApiData holds data for module api templates.
type ModuleApiData struct {
	Name       string // lowercase api name (e.g., "category")
	NameTitle  string // title case api name (e.g., "Category")
	ModuleName string // lowercase module name (e.g., "goods")
	ModPath    string // go module path
}
