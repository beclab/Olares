package scripts

import "embed"

//go:embed files/*
var scripts embed.FS

func Assets() embed.FS {
	return scripts
}
