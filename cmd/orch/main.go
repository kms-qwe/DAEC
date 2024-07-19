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
		"starting orchestrator", slog.Any("cfg", cfg),
	)

	application := orchApp.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)
}
