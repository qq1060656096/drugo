package gomod_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/qq1060656096/drugo/pkg/gomod"
)

// ExampleHasGoMod 演示如何检查目录中是否存在 go.mod 文件。
func ExampleHasGoMod() {
	// 创建临时目录用于演示
	tempDir, _ := os.MkdirTemp("", "gomod-example-*")
	defer os.RemoveAll(tempDir)

	// 检查空目录（没有 go.mod）
	exists := gomod.HasGoMod(tempDir)
	fmt.Printf("空目录存在 go.mod: %v\n", exists)

	// 创建 go.mod 文件
	gomodPath := filepath.Join(tempDir, "go.mod")
	_ = os.WriteFile(gomodPath, []byte("module example.com/demo"), 0644)

	// 再次检查（有 go.mod）
	exists = gomod.HasGoMod(tempDir)
	fmt.Printf("创建后存在 go.mod: %v\n", exists)

	// Output:
	// 空目录存在 go.mod: false
	// 创建后存在 go.mod: true
}

// ExampleFindGoModRoot 演示如何从子目录向上查找 go.mod 所在的项目根目录。
func ExampleFindGoModRoot() {
	// 创建临时项目结构
	tempDir, _ := os.MkdirTemp("", "gomod-example-*")
	defer os.RemoveAll(tempDir)

	// 创建项目根目录和 go.mod
	projectRoot := filepath.Join(tempDir, "myproject")
	_ = os.Mkdir(projectRoot, 0755)
	_ = os.WriteFile(filepath.Join(projectRoot, "go.mod"), []byte("module example.com/myproject"), 0644)

	// 创建深层子目录
	deepDir := filepath.Join(projectRoot, "internal", "pkg", "utils")
	_ = os.MkdirAll(deepDir, 0755)

	// 从深层子目录查找项目根
	root, found := gomod.FindGoModRoot(deepDir)
	fmt.Printf("找到项目根: %v\n", found)
	fmt.Printf("根目录名称: %s\n", filepath.Base(root))

	// 从没有 go.mod 的目录查找
	isolatedDir := filepath.Join(tempDir, "isolated")
	_ = os.Mkdir(isolatedDir, 0755)
	_, found = gomod.FindGoModRoot(isolatedDir)
	fmt.Printf("隔离目录找到项目根: %v\n", found)

	// Output:
	// 找到项目根: true
	// 根目录名称: myproject
	// 隔离目录找到项目根: false
}

// ExampleProjectRoot 演示如何获取项目根目录。
func ExampleProjectRoot() {
	// 创建临时项目结构
	tempDir, _ := os.MkdirTemp("", "gomod-example-*")
	defer os.RemoveAll(tempDir)

	// 创建项目根目录和 go.mod
	projectRoot := filepath.Join(tempDir, "webapp")
	_ = os.Mkdir(projectRoot, 0755)
	_ = os.WriteFile(filepath.Join(projectRoot, "go.mod"), []byte("module example.com/webapp"), 0644)

	// 创建子目录
	cmdDir := filepath.Join(projectRoot, "cmd", "server")
	_ = os.MkdirAll(cmdDir, 0755)

	// 获取项目根目录
	root := gomod.ProjectRoot(cmdDir)
	fmt.Printf("项目根目录名称: %s\n", filepath.Base(root))

	// 当找不到 go.mod 时，返回传入的目录
	isolatedDir := filepath.Join(tempDir, "standalone")
	_ = os.Mkdir(isolatedDir, 0755)
	root = gomod.ProjectRoot(isolatedDir)
	fmt.Printf("独立目录返回自身: %v\n", root == isolatedDir)

	// Output:
	// 项目根目录名称: webapp
	// 独立目录返回自身: true
}

// ExampleFindGoModRoot_nestedProjects 演示嵌套项目场景下如何正确找到最近的 go.mod。
func ExampleFindGoModRoot_nestedProjects() {
	// 创建临时目录
	tempDir, _ := os.MkdirTemp("", "gomod-example-*")
	defer os.RemoveAll(tempDir)

	// 创建外层项目
	outerProject := filepath.Join(tempDir, "monorepo")
	_ = os.Mkdir(outerProject, 0755)
	_ = os.WriteFile(filepath.Join(outerProject, "go.mod"), []byte("module example.com/monorepo"), 0644)

	// 创建内层子模块
	innerModule := filepath.Join(outerProject, "services", "api")
	_ = os.MkdirAll(innerModule, 0755)
	_ = os.WriteFile(filepath.Join(innerModule, "go.mod"), []byte("module example.com/monorepo/services/api"), 0644)

	// 创建子模块的子目录
	handlerDir := filepath.Join(innerModule, "internal", "handlers")
	_ = os.MkdirAll(handlerDir, 0755)

	// 从子模块子目录查找，应找到最近的（内层）go.mod
	root, _ := gomod.FindGoModRoot(handlerDir)
	fmt.Printf("找到最近的模块: %s\n", filepath.Base(root))

	// 从外层项目直接查找
	root, _ = gomod.FindGoModRoot(outerProject)
	fmt.Printf("外层项目根: %s\n", filepath.Base(root))

	// Output:
	// 找到最近的模块: api
	// 外层项目根: monorepo
}
