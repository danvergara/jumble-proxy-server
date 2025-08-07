package config

import (
	"log/slog"

	"github.com/coocood/freecache"
)

type Config struct {
	Host   string
	Port   string
	Logger *slog.Logger
	Cache  *freecache.Cache
}
