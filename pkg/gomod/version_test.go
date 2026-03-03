package gomod

import (
	"runtime/debug"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// setBuildInfo 用于在测试中注入构建信息。
// 由于被测代码使用 sync.Once 缓存构建信息，这里通过提前将 once 标记为已执行，
// 来避免 load() 再次读取真实的构建信息，保证测试稳定性。
func setBuildInfo(bi *debug.BuildInfo) {
	once = sync.Once{}
	once.Do(func() {})
	info = bi
}

func TestMainVersion(t *testing.T) {
	t.Run("构建信息不可用", func(t *testing.T) {
		setBuildInfo(nil)
		v, ok := MainVersion()
		assert.False(t, ok, "构建信息不可用时应该返回 false")
		assert.Equal(t, "", v, "构建信息不可用时版本应该为空")
	})

	t.Run("主模块版本为空", func(t *testing.T) {
		setBuildInfo(&debug.BuildInfo{Main: debug.Module{Path: "example.com/app", Version: ""}})
		v, ok := MainVersion()
		assert.False(t, ok, "主模块版本为空时应该返回 false")
		assert.Equal(t, "", v, "主模块版本为空时版本应该为空")
	})

	t.Run("成功返回主模块版本", func(t *testing.T) {
		setBuildInfo(&debug.BuildInfo{Main: debug.Module{Path: "example.com/app", Version: "v1.2.3"}})
		v, ok := MainVersion()
		assert.True(t, ok, "主模块版本存在时应该返回 true")
		assert.Equal(t, "v1.2.3", v, "主模块版本应该匹配")
	})
}

func TestVersion(t *testing.T) {
	t.Run("构建信息不可用", func(t *testing.T) {
		setBuildInfo(nil)
		v, ok := Version("example.com/app")
		assert.False(t, ok, "构建信息不可用时应该返回 false")
		assert.Equal(t, "", v, "构建信息不可用时版本应该为空")
	})

	t.Run("查询主模块-版本为空", func(t *testing.T) {
		setBuildInfo(&debug.BuildInfo{Main: debug.Module{Path: "example.com/app", Version: ""}})
		v, ok := Version("example.com/app")
		assert.False(t, ok, "主模块版本为空时应该返回 false")
		assert.Equal(t, "", v, "主模块版本为空时版本应该为空")
	})

	t.Run("查询主模块-成功返回", func(t *testing.T) {
		setBuildInfo(&debug.BuildInfo{Main: debug.Module{Path: "example.com/app", Version: "v0.0.1"}})
		v, ok := Version("example.com/app")
		assert.True(t, ok, "主模块版本存在时应该返回 true")
		assert.Equal(t, "v0.0.1", v, "主模块版本应该匹配")
	})

	t.Run("查询依赖模块-存在 replace 且 replace 版本非空", func(t *testing.T) {
		setBuildInfo(&debug.BuildInfo{
			Main: debug.Module{Path: "example.com/app", Version: "v0.0.1"},
			Deps: []*debug.Module{
				{
					Path:    "github.com/gin-gonic/gin",
					Version: "v1.9.0",
					Replace: &debug.Module{Path: "github.com/gin-gonic/gin", Version: "v1.10.0"},
				},
			},
		})
		v, ok := Version("github.com/gin-gonic/gin")
		assert.True(t, ok, "找到依赖模块且 replace 版本非空时应该返回 true")
		assert.Equal(t, "v1.10.0", v, "应优先返回 replace 版本")
	})

	t.Run("查询依赖模块-replace 存在但版本为空则回退到原版本", func(t *testing.T) {
		setBuildInfo(&debug.BuildInfo{
			Main: debug.Module{Path: "example.com/app", Version: "v0.0.1"},
			Deps: []*debug.Module{
				{
					Path:    "github.com/gin-gonic/gin",
					Version: "v1.9.0",
					Replace: &debug.Module{Path: "github.com/gin-gonic/gin", Version: ""},
				},
			},
		})
		v, ok := Version("github.com/gin-gonic/gin")
		assert.True(t, ok, "找到依赖模块且原版本非空时应该返回 true")
		assert.Equal(t, "v1.9.0", v, "replace 版本为空时应返回原版本")
	})

	t.Run("查询依赖模块-版本都为空", func(t *testing.T) {
		setBuildInfo(&debug.BuildInfo{
			Main: debug.Module{Path: "example.com/app", Version: "v0.0.1"},
			Deps: []*debug.Module{
				{Path: "github.com/gin-gonic/gin", Version: ""},
			},
		})
		v, ok := Version("github.com/gin-gonic/gin")
		assert.False(t, ok, "依赖模块版本为空时应该返回 false")
		assert.Equal(t, "", v, "依赖模块版本为空时版本应该为空")
	})

	t.Run("未找到模块", func(t *testing.T) {
		setBuildInfo(&debug.BuildInfo{Main: debug.Module{Path: "example.com/app", Version: "v0.0.1"}})
		v, ok := Version("not-found")
		assert.False(t, ok, "未找到模块时应该返回 false")
		assert.Equal(t, "", v, "未找到模块时版本应该为空")
	})

	t.Run("空路径参数", func(t *testing.T) {
		setBuildInfo(&debug.BuildInfo{Main: debug.Module{Path: "example.com/app", Version: "v0.0.1"}})
		v, ok := Version("")
		assert.False(t, ok, "空路径参数时应该返回 false")
		assert.Equal(t, "", v, "空路径参数时版本应该为空")
	})
}

func TestAllVersion(t *testing.T) {
	t.Run("构建信息不可用", func(t *testing.T) {
		setBuildInfo(nil)
		m := AllVersion()
		assert.Nil(t, m, "构建信息不可用时应该返回 nil")
	})

	t.Run("返回主模块与依赖模块版本", func(t *testing.T) {
		setBuildInfo(&debug.BuildInfo{
			Main: debug.Module{Path: "example.com/app", Version: "v0.0.1"},
			Deps: []*debug.Module{
				{Path: "github.com/gin-gonic/gin", Version: "v1.9.0"},
				{Path: "github.com/spf13/viper", Version: "", Replace: &debug.Module{Path: "github.com/spf13/viper", Version: "v1.20.0"}},
				{Path: "example.com/empty", Version: ""},
			},
		})

		m := AllVersion()
		assert.NotNil(t, m, "构建信息存在时应该返回 map")
		assert.Equal(t, "v0.0.1", m["example.com/app"], "主模块版本应该存在")
		assert.Equal(t, "v1.9.0", m["github.com/gin-gonic/gin"], "依赖模块版本应该存在")
		assert.Equal(t, "v1.20.0", m["github.com/spf13/viper"], "replace 版本应该优先")
		_, ok := m["example.com/empty"]
		assert.False(t, ok, "空版本模块不应出现在返回结果中")
	})
}

func BenchmarkMainVersion(b *testing.B) {
	setBuildInfo(&debug.BuildInfo{Main: debug.Module{Path: "example.com/app", Version: "v1.2.3"}})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MainVersion()
	}
}

func BenchmarkVersion_DepReplace(b *testing.B) {
	setBuildInfo(&debug.BuildInfo{
		Main: debug.Module{Path: "example.com/app", Version: "v0.0.1"},
		Deps: []*debug.Module{
			{
				Path:    "github.com/gin-gonic/gin",
				Version: "v1.9.0",
				Replace: &debug.Module{Path: "github.com/gin-gonic/gin", Version: "v1.10.0"},
			},
		},
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Version("github.com/gin-gonic/gin")
	}
}

func BenchmarkAllVersion(b *testing.B) {
	setBuildInfo(&debug.BuildInfo{
		Main: debug.Module{Path: "example.com/app", Version: "v0.0.1"},
		Deps: []*debug.Module{
			{Path: "github.com/gin-gonic/gin", Version: "v1.9.0"},
			{Path: "github.com/spf13/viper", Version: "v1.18.0"},
			{Path: "go.uber.org/zap", Version: "v1.27.0"},
		},
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AllVersion()
	}
}
