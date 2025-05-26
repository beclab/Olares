//go:build ignore

package main

import (
	"log"
	"net/http"
	"path"

	"github.com/shurcooL/vfsgen"
)

func main() {
	fs := http.Dir(path.Join("pkg", "kubesphere", "plugins", "files"))
	dst := path.Join("pkg", "kubesphere", "plugins", "assets_vfsdata.go")
	err := vfsgen.Generate(fs, vfsgen.Options{
		Filename:    dst,
		PackageName: "plugins",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
