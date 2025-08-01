package config

import "log/slog"

type Config struct {
	Host   string
	Port   string
	Logger *slog.Logger
}
