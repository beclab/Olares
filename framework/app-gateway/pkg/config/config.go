package config

import internal "github.com/beclab/Olares/framework/app-gateway/internal/config"

type Defaults = internal.Defaults

func Load() (Defaults, error) { return internal.Load() }

func Namespace() string { return internal.Namespace() }
