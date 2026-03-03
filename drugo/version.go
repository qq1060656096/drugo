package drugo

import (
	"sync"

	"github.com/qq1060656096/drugo/pkg/gomod"
)

var version = "dev"

func init() {
	sync.OnceFunc(func() {
		_version, _ := gomod.Version("github.com/qq1060656096/drugo")
		if _version != "" {
			version = _version
		}
	})
}

// Version 获取当前框架的版本号
// 调用 gomod.Version 方法获取指定包的版本信息
// 返回值：版本号字符串，如果获取失败则返回空字符串
func Version() string {
	return version
}
