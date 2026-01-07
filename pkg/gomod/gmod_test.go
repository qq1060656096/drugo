package gomod

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHasGoMod 测试 HasGoMod 函数
func TestHasGoMod(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("目录中存在go.mod文件", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "with_gomod")
		err := os.Mkdir(testDir, 0755)
		require.NoError(t, err)

		gomodPath := filepath.Join(testDir, "go.mod")
		err = os.WriteFile(gomodPath, []byte("module test"), 0644)
		require.NoError(t, err)

		result := HasGoMod(testDir)
		assert.True(t, result, "目录中存在go.mod文件时应该返回true")
	})

	t.Run("目录中不存在go.mod文件", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "no_gomod")
		err := os.Mkdir(testDir, 0755)
		require.NoError(t, err)

		result := HasGoMod(testDir)
		assert.False(t, result, "目录中不存在go.mod文件时应该返回false")
	})

	t.Run("目录不存在", func(t *testing.T) {
		nonExistentDir := filepath.Join(tempDir, "non_existent")
		result := HasGoMod(nonExistentDir)
		assert.False(t, result, "目录不存在时应该返回false")
	})

	t.Run("空字符串路径", func(t *testing.T) {
		result := HasGoMod("")
		// 空字符串会被解析为当前目录，取决于当前目录是否有 go.mod
		// 这里主要测试不会 panic
		_ = result
	})

	t.Run("go.mod是目录而非文件", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "gomod_is_dir")
		err := os.Mkdir(testDir, 0755)
		require.NoError(t, err)

		// 创建一个名为 go.mod 的目录
		gomodDir := filepath.Join(testDir, "go.mod")
		err = os.Mkdir(gomodDir, 0755)
		require.NoError(t, err)

		result := HasGoMod(testDir)
		// os.Stat 会返回成功，因为路径存在
		assert.True(t, result, "go.mod路径存在时返回true，即使是目录")
	})
}

// TestFindGoModRoot 测试 FindGoModRoot 函数
func TestFindGoModRoot(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("当前目录包含go.mod文件", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "project_root")
		err := os.Mkdir(testDir, 0755)
		require.NoError(t, err)

		gomodPath := filepath.Join(testDir, "go.mod")
		err = os.WriteFile(gomodPath, []byte("module test-project"), 0644)
		require.NoError(t, err)

		result, found := FindGoModRoot(testDir)
		assert.True(t, found, "应该找到go.mod")
		assert.Equal(t, testDir, result, "包含go.mod文件的目录应该返回自身")
	})

	t.Run("需要向上查找go.mod文件", func(t *testing.T) {
		projectRoot := filepath.Join(tempDir, "project")
		err := os.Mkdir(projectRoot, 0755)
		require.NoError(t, err)

		gomodPath := filepath.Join(projectRoot, "go.mod")
		err = os.WriteFile(gomodPath, []byte("module test-project"), 0644)
		require.NoError(t, err)

		subDir := filepath.Join(projectRoot, "subdir")
		err = os.Mkdir(subDir, 0755)
		require.NoError(t, err)

		result, found := FindGoModRoot(subDir)
		assert.True(t, found, "应该找到go.mod")
		assert.Equal(t, projectRoot, result, "应该找到项目根目录")
	})

	t.Run("多层嵌套目录向上查找", func(t *testing.T) {
		projectRoot := filepath.Join(tempDir, "deep_project")
		err := os.Mkdir(projectRoot, 0755)
		require.NoError(t, err)

		gomodPath := filepath.Join(projectRoot, "go.mod")
		err = os.WriteFile(gomodPath, []byte("module test-project"), 0644)
		require.NoError(t, err)

		deepDir := filepath.Join(projectRoot, "a/b/c/d/e")
		err = os.MkdirAll(deepDir, 0755)
		require.NoError(t, err)

		result, found := FindGoModRoot(deepDir)
		assert.True(t, found, "应该从深层目录找到go.mod")
		assert.Equal(t, projectRoot, result, "应该从深层目录找到项目根目录")
	})

	t.Run("没有找到go.mod文件", func(t *testing.T) {
		noGomodDir := filepath.Join(tempDir, "no_gomod_project")
		err := os.Mkdir(noGomodDir, 0755)
		require.NoError(t, err)

		subDir := filepath.Join(noGomodDir, "subdir")
		err = os.Mkdir(subDir, 0755)
		require.NoError(t, err)

		// 注意：这个测试可能会找到当前工作目录的 go.mod
		// 为了确保测试隔离，我们使用绝对路径
		result, found := FindGoModRoot(subDir)

		// 如果在 tempDir 的父目录链上有 go.mod，可能会找到
		// 我们只验证返回的路径是 subDir 的祖先目录
		if found {
			rel, err := filepath.Rel(result, subDir)
			assert.NoError(t, err)
			assert.NotContains(t, rel, "..", "找到的根目录应该是起始目录的祖先")
		} else {
			assert.Empty(t, result, "未找到时应该返回空字符串")
		}
	})

	t.Run("嵌套项目选择最近的go.mod", func(t *testing.T) {
		outerProject := filepath.Join(tempDir, "outer")
		err := os.Mkdir(outerProject, 0755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(outerProject, "go.mod"), []byte("module outer"), 0644)
		require.NoError(t, err)

		innerProject := filepath.Join(outerProject, "inner")
		err = os.Mkdir(innerProject, 0755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(innerProject, "go.mod"), []byte("module inner"), 0644)
		require.NoError(t, err)

		subDir := filepath.Join(innerProject, "sub")
		err = os.Mkdir(subDir, 0755)
		require.NoError(t, err)

		result, found := FindGoModRoot(subDir)
		assert.True(t, found)
		assert.Equal(t, innerProject, result, "应该找到最近的go.mod所在目录")
	})

	t.Run("相对路径转换为绝对路径", func(t *testing.T) {
		// 使用当前目录测试相对路径处理
		result, found := FindGoModRoot(".")
		if found {
			assert.True(t, filepath.IsAbs(result), "返回的路径应该是绝对路径")
		}
	})
}

// TestProjectRoot 测试 ProjectRoot 函数
func TestProjectRoot(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("找到项目根目录", func(t *testing.T) {
		projectRoot := filepath.Join(tempDir, "myproject")
		err := os.Mkdir(projectRoot, 0755)
		require.NoError(t, err)

		gomodPath := filepath.Join(projectRoot, "go.mod")
		err = os.WriteFile(gomodPath, []byte("module github.com/user/myproject"), 0644)
		require.NoError(t, err)

		subDir := filepath.Join(projectRoot, "cmd/server")
		err = os.MkdirAll(subDir, 0755)
		require.NoError(t, err)

		result := ProjectRoot(subDir)
		assert.Equal(t, projectRoot, result, "应该返回项目根目录")
	})

	t.Run("未找到go.mod返回runDir", func(t *testing.T) {
		// 创建一个独立的目录结构，确保向上找不到任何 go.mod
		// 由于 tempDir 可能在某个有 go.mod 的项目中，这个测试比较特殊
		noGomodDir := filepath.Join(tempDir, "isolated")
		err := os.Mkdir(noGomodDir, 0755)
		require.NoError(t, err)

		result := ProjectRoot(noGomodDir)
		// ProjectRoot 在未找到时返回 runDir，但由于可能找到父目录的 go.mod
		// 我们验证返回的是一个有效路径
		assert.NotEmpty(t, result, "应该返回有效路径")
		assert.Equal(t, noGomodDir, result, "应该返回原始目录")
	})

	t.Run("从深层子目录查找", func(t *testing.T) {
		projectRoot := filepath.Join(tempDir, "deep_structure")
		err := os.Mkdir(projectRoot, 0755)
		require.NoError(t, err)

		gomodPath := filepath.Join(projectRoot, "go.mod")
		err = os.WriteFile(gomodPath, []byte("module deep-project"), 0644)
		require.NoError(t, err)

		deepDir := filepath.Join(projectRoot, "internal/pkg/utils/helpers")
		err = os.MkdirAll(deepDir, 0755)
		require.NoError(t, err)

		result := ProjectRoot(deepDir)
		assert.Equal(t, projectRoot, result, "应该从深层目录找到项目根目录")
	})
}

// TestFindGoModRoot_Integration 集成测试
func TestFindGoModRoot_Integration(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("完整项目结构测试", func(t *testing.T) {
		projectRoot := filepath.Join(tempDir, "fullproject")
		err := os.Mkdir(projectRoot, 0755)
		require.NoError(t, err)

		gomodPath := filepath.Join(projectRoot, "go.mod")
		err = os.WriteFile(gomodPath, []byte("module github.com/user/fullproject"), 0644)
		require.NoError(t, err)

		dirs := []string{
			"cmd/server",
			"cmd/client",
			"internal/config",
			"internal/service",
			"pkg/utils",
			"pkg/models",
			"web/static/css",
			"web/templates",
			"docs/api",
			"test/integration",
		}

		for _, dir := range dirs {
			fullPath := filepath.Join(projectRoot, dir)
			err = os.MkdirAll(fullPath, 0755)
			require.NoError(t, err)

			result, found := FindGoModRoot(fullPath)
			assert.True(t, found, "从 %s 查找应该成功", dir)
			assert.Equal(t, projectRoot, result, "从 %s 查找应该返回项目根目录", dir)
		}
	})

	t.Run("monorepo结构测试", func(t *testing.T) {
		monorepo := filepath.Join(tempDir, "monorepo")
		err := os.Mkdir(monorepo, 0755)
		require.NoError(t, err)

		// 创建多个子模块
		modules := []string{"services/api", "services/worker", "libs/common"}
		for _, mod := range modules {
			modPath := filepath.Join(monorepo, mod)
			err = os.MkdirAll(modPath, 0755)
			require.NoError(t, err)

			gomodPath := filepath.Join(modPath, "go.mod")
			err = os.WriteFile(gomodPath, []byte("module "+mod), 0644)
			require.NoError(t, err)

			// 创建子目录
			subDir := filepath.Join(modPath, "internal/handlers")
			err = os.MkdirAll(subDir, 0755)
			require.NoError(t, err)

			result, found := FindGoModRoot(subDir)
			assert.True(t, found)
			assert.Equal(t, modPath, result, "应该找到子模块的go.mod，而不是父目录")
		}
	})
}

// BenchmarkHasGoMod 性能测试 HasGoMod 函数
func BenchmarkHasGoMod(b *testing.B) {
	tempDir := b.TempDir()

	b.Run("存在go.mod", func(b *testing.B) {
		testDir := filepath.Join(tempDir, "with_gomod")
		_ = os.Mkdir(testDir, 0755)
		_ = os.WriteFile(filepath.Join(testDir, "go.mod"), []byte("module bench"), 0644)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			HasGoMod(testDir)
		}
	})

	b.Run("不存在go.mod", func(b *testing.B) {
		testDir := filepath.Join(tempDir, "no_gomod")
		_ = os.Mkdir(testDir, 0755)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			HasGoMod(testDir)
		}
	})
}

// BenchmarkFindGoModRoot 性能测试 FindGoModRoot 函数
func BenchmarkFindGoModRoot(b *testing.B) {
	tempDir := b.TempDir()

	b.Run("浅层目录", func(b *testing.B) {
		projectRoot := filepath.Join(tempDir, "shallow")
		_ = os.Mkdir(projectRoot, 0755)
		_ = os.WriteFile(filepath.Join(projectRoot, "go.mod"), []byte("module bench"), 0644)

		subDir := filepath.Join(projectRoot, "cmd")
		_ = os.Mkdir(subDir, 0755)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			FindGoModRoot(subDir)
		}
	})

	b.Run("深层目录", func(b *testing.B) {
		projectRoot := filepath.Join(tempDir, "deep")
		_ = os.Mkdir(projectRoot, 0755)
		_ = os.WriteFile(filepath.Join(projectRoot, "go.mod"), []byte("module bench"), 0644)

		deepDir := filepath.Join(projectRoot, "a/b/c/d/e/f/g/h")
		_ = os.MkdirAll(deepDir, 0755)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			FindGoModRoot(deepDir)
		}
	})
}

// BenchmarkProjectRoot 性能测试 ProjectRoot 函数
func BenchmarkProjectRoot(b *testing.B) {
	tempDir := b.TempDir()
	projectRoot := filepath.Join(tempDir, "bench_project")
	_ = os.Mkdir(projectRoot, 0755)
	_ = os.WriteFile(filepath.Join(projectRoot, "go.mod"), []byte("module bench"), 0644)

	deepDir := filepath.Join(projectRoot, "internal/pkg/utils")
	_ = os.MkdirAll(deepDir, 0755)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ProjectRoot(deepDir)
	}
}

// TestModuleName 测试 ModuleName 函数
func TestModuleName(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("正常读取模块名称", func(t *testing.T) {
		projectDir := filepath.Join(tempDir, "normal_project")
		err := os.Mkdir(projectDir, 0755)
		require.NoError(t, err)

		gomodContent := "module github.com/user/myproject\n\ngo 1.21\n"
		err = os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte(gomodContent), 0644)
		require.NoError(t, err)

		name, err := ModuleName(projectDir)
		assert.NoError(t, err)
		assert.Equal(t, "github.com/user/myproject", name)
	})

	t.Run("目录不存在", func(t *testing.T) {
		_, err := ModuleName(filepath.Join(tempDir, "nonexistent"))
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("go.mod文件不存在", func(t *testing.T) {
		noGomodDir := filepath.Join(tempDir, "no_gomod")
		err := os.Mkdir(noGomodDir, 0755)
		require.NoError(t, err)

		_, err = ModuleName(noGomodDir)
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("go.mod无效内容", func(t *testing.T) {
		invalidDir := filepath.Join(tempDir, "invalid_gomod")
		err := os.Mkdir(invalidDir, 0755)
		require.NoError(t, err)

		// 写入无效的 go.mod 内容（没有 module 声明）
		err = os.WriteFile(filepath.Join(invalidDir, "go.mod"), []byte("go 1.21\n"), 0644)
		require.NoError(t, err)

		_, err = ModuleName(invalidDir)
		assert.Error(t, err)
		var notFoundErr *ModuleNameNotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})
}

// TestParseModuleName 测试 ParseModuleName 函数
func TestParseModuleName(t *testing.T) {
	t.Run("标准模块声明", func(t *testing.T) {
		content := []byte("module github.com/user/project\n\ngo 1.21\n")
		name, err := ParseModuleName(content)
		assert.NoError(t, err)
		assert.Equal(t, "github.com/user/project", name)
	})

	t.Run("简单模块名", func(t *testing.T) {
		content := []byte("module myapp")
		name, err := ParseModuleName(content)
		assert.NoError(t, err)
		assert.Equal(t, "myapp", name)
	})

	t.Run("模块名带版本号", func(t *testing.T) {
		content := []byte("module github.com/user/project/v2\n\ngo 1.21\n")
		name, err := ParseModuleName(content)
		assert.NoError(t, err)
		assert.Equal(t, "github.com/user/project/v2", name)
	})

	t.Run("模块声明前有注释", func(t *testing.T) {
		content := []byte("// This is a comment\n// Another comment\nmodule example.com/app\n")
		name, err := ParseModuleName(content)
		assert.NoError(t, err)
		assert.Equal(t, "example.com/app", name)
	})

	t.Run("模块声明前有空行", func(t *testing.T) {
		content := []byte("\n\n\nmodule example.com/app\n")
		name, err := ParseModuleName(content)
		assert.NoError(t, err)
		assert.Equal(t, "example.com/app", name)
	})

	t.Run("模块声明带行尾注释", func(t *testing.T) {
		content := []byte("module example.com/app // some comment\n")
		name, err := ParseModuleName(content)
		assert.NoError(t, err)
		assert.Equal(t, "example.com/app", name)
	})

	t.Run("模块声明有额外空格", func(t *testing.T) {
		content := []byte("  module   example.com/app  \n")
		name, err := ParseModuleName(content)
		assert.NoError(t, err)
		assert.Equal(t, "example.com/app", name)
	})

	t.Run("空内容", func(t *testing.T) {
		content := []byte("")
		_, err := ParseModuleName(content)
		assert.Error(t, err)
		var notFoundErr *ModuleNameNotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})

	t.Run("只有注释", func(t *testing.T) {
		content := []byte("// comment only\n// another comment\n")
		_, err := ParseModuleName(content)
		assert.Error(t, err)
		var notFoundErr *ModuleNameNotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})

	t.Run("只有空行", func(t *testing.T) {
		content := []byte("\n\n\n")
		_, err := ParseModuleName(content)
		assert.Error(t, err)
		var notFoundErr *ModuleNameNotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})

	t.Run("module关键字后没有名称", func(t *testing.T) {
		content := []byte("module \n")
		_, err := ParseModuleName(content)
		assert.Error(t, err)
		var notFoundErr *ModuleNameNotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})

	t.Run("module关键字后只有注释", func(t *testing.T) {
		content := []byte("module // comment\n")
		_, err := ParseModuleName(content)
		assert.Error(t, err)
		var notFoundErr *ModuleNameNotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})

	t.Run("完整go.mod文件", func(t *testing.T) {
		content := []byte(`module github.com/qq1060656096/drugo

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/spf13/viper v1.16.0
)
`)
		name, err := ParseModuleName(content)
		assert.NoError(t, err)
		assert.Equal(t, "github.com/qq1060656096/drugo", name)
	})

	t.Run("没有module声明的go.mod", func(t *testing.T) {
		content := []byte(`go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
)
`)
		_, err := ParseModuleName(content)
		assert.Error(t, err)
		var notFoundErr *ModuleNameNotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})

	t.Run("module在其他行中间出现", func(t *testing.T) {
		// 这个测试确保只识别行首的 module 关键字
		content := []byte("// module fake\nmodule real.example.com/app\n")
		name, err := ParseModuleName(content)
		assert.NoError(t, err)
		assert.Equal(t, "real.example.com/app", name)
	})
}

// TestModuleNameNotFoundError 测试 ModuleNameNotFoundError 错误类型
func TestModuleNameNotFoundError(t *testing.T) {
	t.Run("错误消息", func(t *testing.T) {
		err := &ModuleNameNotFoundError{}
		assert.Equal(t, "module name not found in go.mod", err.Error())
	})

	t.Run("实现error接口", func(t *testing.T) {
		var err error = &ModuleNameNotFoundError{}
		assert.NotNil(t, err)
		assert.Equal(t, "module name not found in go.mod", err.Error())
	})

	t.Run("errors.As匹配", func(t *testing.T) {
		err := &ModuleNameNotFoundError{}
		var target *ModuleNameNotFoundError
		assert.True(t, errors.As(err, &target))
	})
}

// BenchmarkModuleName 性能测试 ModuleName 函数
func BenchmarkModuleName(b *testing.B) {
	tempDir := b.TempDir()
	projectDir := filepath.Join(tempDir, "bench_module")
	_ = os.Mkdir(projectDir, 0755)
	_ = os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte("module github.com/user/benchmark\n\ngo 1.21\n"), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ModuleName(projectDir)
	}
}

// BenchmarkParseModuleName 性能测试 ParseModuleName 函数
func BenchmarkParseModuleName(b *testing.B) {
	content := []byte(`module github.com/qq1060656096/drugo

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/spf13/viper v1.16.0
	go.uber.org/zap v1.25.0
)
`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseModuleName(content)
	}
}
