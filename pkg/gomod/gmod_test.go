package gomod

import (
"os"
"path/filepath"
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

// TestGmodNotExist 测试 GmodNotExist 函数
func TestGmodNotExist(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()

	// 测试用例1: 目录中不存在 go.mod 文件
	t.Run("目录中不存在go.mod文件", func(t *testing.T) {
testDir := filepath.Join(tempDir, "no_gomod")
err := os.Mkdir(testDir, 0755)
require.NoError(t, err)

result := GmodNotExist(testDir)
assert.True(t, result, "目录中不存在go.mod文件时应该返回true")
})

	// 测试用例2: 目录中存在 go.mod 文件
	t.Run("目录中存在go.mod文件", func(t *testing.T) {
testDir := filepath.Join(tempDir, "with_gomod")
err := os.Mkdir(testDir, 0755)
require.NoError(t, err)

gomodPath := filepath.Join(testDir, "go.mod")
err = os.WriteFile(gomodPath, []byte("module test"), 0644)
require.NoError(t, err)

result := GmodNotExist(testDir)
assert.False(t, result, "目录中存在go.mod文件时应该返回false")
})

	// 测试用例3: 目录不存在
	t.Run("目录不存在", func(t *testing.T) {
nonExistentDir := filepath.Join(tempDir, "non_existent")
result := GmodNotExist(nonExistentDir)
assert.True(t, result, "目录不存在时应该返回true")
})

	// 测试用例4: 空字符串路径
	t.Run("空字符串路径", func(t *testing.T) {
result := GmodNotExist("")
assert.True(t, result, "空字符串路径应该返回true")
})
}

// TestGmodRoot 测试 GmodRoot 函数
func TestGmodRoot(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()

	// 测试用例1: 当前目录就是根目录
	t.Run("当前目录是根目录", func(t *testing.T) {
result := GmodRoot(".")
assert.Equal(t, ".", result, "根目录应该返回自身")
})

	// 测试用例2: 当前目录包含 go.mod 文件
	t.Run("当前目录包含go.mod文件", func(t *testing.T) {
testDir := filepath.Join(tempDir, "project_root")
err := os.Mkdir(testDir, 0755)
require.NoError(t, err)

gomodPath := filepath.Join(testDir, "go.mod")
err = os.WriteFile(gomodPath, []byte("module test-project"), 0644)
require.NoError(t, err)

result := GmodRoot(testDir)
assert.Equal(t, testDir, result, "包含go.mod文件的目录应该返回自身")
})

	// 测试用例3: 需要向上查找 go.mod 文件
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

deepSubDir := filepath.Join(subDir, "deep")
err = os.Mkdir(deepSubDir, 0755)
require.NoError(t, err)

result := GmodRoot(subDir)
assert.Equal(t, projectRoot, result, "应该找到项目根目录")

result = GmodRoot(deepSubDir)
assert.Equal(t, projectRoot, result, "应该从深层目录找到项目根目录")
})

	// 测试用例4: 没有找到 go.mod 文件
	t.Run("没有找到go.mod文件", func(t *testing.T) {
noGomodDir := filepath.Join(tempDir, "no_gomod_project")
err := os.Mkdir(noGomodDir, 0755)
require.NoError(t, err)

subDir := filepath.Join(noGomodDir, "subdir")
err = os.Mkdir(subDir, 0755)
require.NoError(t, err)

result := GmodRoot(subDir)
assert.NotEmpty(t, result, "应该返回一个有效路径")
})

	// 测试用例5: 空字符串输入
	t.Run("空字符串输入", func(t *testing.T) {
result := GmodRoot("")
assert.NotEmpty(t, result, "空字符串输入应该返回有效路径")
})
}

// TestGmodRoot_Integration 集成测试
func TestGmodRoot_Integration(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("完整项目结构测试", func(t *testing.T) {
projectRoot := filepath.Join(tempDir, "myproject")
err := os.Mkdir(projectRoot, 0755)
require.NoError(t, err)

gomodPath := filepath.Join(projectRoot, "go.mod")
err = os.WriteFile(gomodPath, []byte("module github.com/user/myproject"), 0644)
require.NoError(t, err)

dirs := []string{
"cmd/server",
"internal/config",
"pkg/utils",
"web/static/css",
"docs/api",
}

for _, dir := range dirs {
fullPath := filepath.Join(projectRoot, dir)
err = os.MkdirAll(fullPath, 0755)
require.NoError(t, err)

result := GmodRoot(fullPath)
assert.Equal(t, projectRoot, result, "从 %s 查找应该返回项目根目录", dir)
}
})
}

// BenchmarkGmodNotExist 性能测试 GmodNotExist 函数
func BenchmarkGmodNotExist(b *testing.B) {
	tempDir := b.TempDir()
	testDir := filepath.Join(tempDir, "bench_test")
	err := os.Mkdir(testDir, 0755)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GmodNotExist(testDir)
	}
}

// BenchmarkGmodRoot 性能测试 GmodRoot 函数
func BenchmarkGmodRoot(b *testing.B) {
	tempDir := b.TempDir()
	projectRoot := filepath.Join(tempDir, "bench_project")
	err := os.Mkdir(projectRoot, 0755)
	if err != nil {
		b.Fatal(err)
	}
	
	err = os.WriteFile(filepath.Join(projectRoot, "go.mod"), []byte("module bench"), 0644)
	if err != nil {
		b.Fatal(err)
	}

	deepDir := filepath.Join(projectRoot, "a/b/c/d/e")
	err = os.MkdirAll(deepDir, 0755)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GmodRoot(deepDir)
	}
}
