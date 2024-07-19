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
	//нельзя запустить, пока нет слоя работы с бд
	//application := orchApp.New(log, &orch.TaskPuller{})
}
