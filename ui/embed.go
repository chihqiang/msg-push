package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/chihqiang/infra-go/httpx"
)

//go:embed all:dist
var distDir embed.FS

// DistDirFS 提供编译进二进制的 Vue 前端静态资源，并支持 SPA history 路由。
// 注意：embed 会保留 "dist" 前缀（dist/index.html），需用 fs.Sub 剥离，
// 使静态资源在根路径下能被正确命中。
// 同时处理 Vue Router history 模式：直接访问/刷新非静态资源路径时回退到 index.html。
var DistDirFS = httpx.Route{
	Method:  http.MethodGet,
	Path:    "/",
	Handler: spaHandler(mustSub(distDir, "dist")),
}

func mustSub(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return sub
}

// spaHandler 先尝试提供静态文件（JS/CSS/图片等）；若文件不存在且不是 API 请求，
// 则回退到 index.html，以支持 Vue Router 的 history 模式。
func spaHandler(fsys fs.FS) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(fsys))
	return func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p == "" {
			p = "index.html"
		}
		// 命中真实静态资源（含目录）→ 交给 http.FileServer
		if _, err := fs.Stat(fsys, p); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}
		// API 请求不做 SPA 回退，保持 404 语义
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		// SPA history 路由 → 返回 index.html
		data, err := fs.ReadFile(fsys, "index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
	}
}
