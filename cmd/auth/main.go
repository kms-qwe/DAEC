package main

import (
	"log/slog"
	"strconv"

	"github.com/kms-qwe/DAEC/internal/app/auth"
	"github.com/kms-qwe/DAEC/internal/config"
	"github.com/kms-qwe/DAEC/internal/lib/logger/setup"
	"github.com/kms-qwe/DAEC/internal/lib/logger/sl"
	"github.com/kms-qwe/DAEC/internal/storage/sqlite"
)

func main() {
	cfg := config.MastLoad()

	log := setup.SetupLogger(cfg.Env)

	log.Info(
		"starting auth", slog.Any("cfg", cfg),
	)

	port := ":" + strconv.Itoa(cfg.HTTP.Port)
	authStorage, err := sqlite.NewAuthStorage(cfg.StoragePath)
	if err != nil {
		log.Error("auth can't connect to db", sl.Err(err))
		panic("auth can't connect ot db")
	}
	log.Info("auth connect to db")
	app := auth.NewServer(log, port, cfg.TokenTTL, authStorage)
	app.MustRun()

}
