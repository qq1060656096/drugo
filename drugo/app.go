package drugo

import "sync/atomic"

var app atomic.Value // 存 *AppInfo

func SetApp(drugoApp *Drugo) {
	if drugoApp == nil {
		panic("global: drugo app info is nil")
	}
	app.Store(drugoApp)
}

func App() *Drugo {
	v := app.Load()
	if v == nil {
		panic("global: drugo app not initialized")
	}
	return v.(*Drugo)
}
