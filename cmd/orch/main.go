package main

import (
	"log/slog"

	orchApp "github.com/kms-qwe/DAEC/internal/app/orch"
	"github.com/kms-qwe/DAEC/internal/config"
	"github.com/kms-qwe/DAEC/internal/grpc/orch"
	"github.com/kms-qwe/DAEC/internal/lib/logger/setup"
	"github.com/kms-qwe/DAEC/internal/lib/logger/sl"
	daecv1 "github.com/kms-qwe/DAEC/internal/protos/gen/go/daec"
	"github.com/kms-qwe/DAEC/internal/storage/sqlite"
)

func main() {
	cfg := config.MastLoad()

	log := setup.SetupLogger(cfg.Env)

	log.Info(
		"starting orchestrator", slog.Any("cfg", cfg),
	)

	orchStorage, err := sqlite.NewOrchStorage(cfg.StoragePath)
	if err != nil {
		log.Error("orch can't connect to db", sl.Err(err))
		panic("orch can't connect ot db")
	}
	log.Info("orch connect to db")
	tP := &orch.TaskPuller{
		Log:         log,
		ExpStrg:     orchStorage,
		ChToAgent:   make(chan *daecv1.TaskResponse),
		ChFromAgent: make(chan *daecv1.ResultRequest),
		ExprID:      0,
		Expr:        "",
	}
	application := orchApp.New(log, tP, cfg.GRPC.Port)
	application.MustRun()

}
