package main

import (
	"context"
	"log/slog"

	"github.com/kms-qwe/DAEC/internal/config"
	"github.com/kms-qwe/DAEC/internal/lib/logger/setup"
	"github.com/kms-qwe/DAEC/internal/lib/logger/sl"
	"github.com/kms-qwe/DAEC/internal/storage/sqlite"
)

func main() {
	cfg := config.MastLoad()

	log := setup.SetupLogger(cfg.Env)

	log.Info(
		"init storage", slog.Any("cfg", cfg),
	)
	storage, err := sqlite.NewInitStorage(cfg.StoragePath)
	if err != nil {
		log.Error("Init storage is not running", sl.Err(err))
		panic("init storage is not running")
	}
	if err := storage.Init(context.TODO()); err != nil {
		log.Info("tables are not init", sl.Err(err))
	}
	log.Info("tables are init successfully")
}
