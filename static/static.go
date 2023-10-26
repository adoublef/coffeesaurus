package static

import (
	"embed"
	"net/http"

	"github.com/benbjohnson/hashfs"
)

var (
	//go:embed all:*
	embedFS embed.FS

	fsys = hashfs.NewFS(embedFS)
)

type Static struct {
	Prefix string
}

func (s *Static) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix(s.Prefix, hashfs.FileServer(hashfs.NewFS(fsys))).ServeHTTP(w, r)
}

func HashName(name string) string {
	return fsys.HashName(name)
}
