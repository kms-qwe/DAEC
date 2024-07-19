package main

import (
	"log/slog"

	"github.com/kms-qwe/DAEC/internal/config"
	"github.com/kms-qwe/DAEC/internal/lib/logger/setup"
)

func main() {

	cfg := config.MastLoad()

	log := setup.SetupLogger(cfg.Env)

	log.Info(
		"starting agent", slog.Any("cfg", cfg),
	)

}
