package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/kms-qwe/DAEC/internal/config"
	"github.com/kms-qwe/DAEC/internal/lib/logger/setup"
	"github.com/kms-qwe/DAEC/internal/lib/logger/sl"
	pb "github.com/kms-qwe/DAEC/internal/protos/gen/go/daec"
	"google.golang.org/grpc"
)

func main() {

	cfg := config.MastLoad()

	log := setup.SetupLogger(cfg.Env)

	log.Info(
		"starting agent", slog.Any("cfg", cfg),
	)

	for range cfg.ComputingPower {
		go worker(log, cfg)
	}

}

func worker(Oldlog *slog.Logger, cfg *config.Config) {
	const op = "agent.main.worker"
	log := Oldlog.With(
		slog.String("op", op),
	)

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Warn("did not connect", sl.Err(err))
	}
	defer conn.Close()

	client := pb.NewOrchServiceClient(conn)

	// Создаем таймер для отправки запроса каждую секунду
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {

		// Отправка запроса на получение задачи
		ctx := context.Background()

		taskResponse, err := client.GiveTask(ctx, &pb.TaskRequest{})
		if err != nil {
			log.Info("could not give task", sl.Err(err))
			continue
		}
		log.Info("Received Task", slog.Any("taskResponse", taskResponse))

		// Отправка результата

		resultRequest := &pb.ResultRequest{
			Id:     taskResponse.Id,
			Result: 0.0,
		}

		switch taskResponse.Operation {
		case "+":
			resultRequest.Result = taskResponse.Arg1 + taskResponse.Arg2
			time.Sleep(cfg.Addition)
		case "-":
			resultRequest.Result = taskResponse.Arg1 - taskResponse.Arg2
			time.Sleep(cfg.Subtraction)
		case "*":
			resultRequest.Result = taskResponse.Arg1 * taskResponse.Arg2
			time.Sleep(cfg.Multiplication)
		case "/":
			resultRequest.Result = taskResponse.Arg1 / taskResponse.Arg2
			time.Sleep(cfg.Division)

		}

		resultResponse, err := client.GetResult(ctx, resultRequest)
		if err != nil {
			log.Info("could not get result", sl.Err(err))
			continue
		}
		log.Info("Received Result", slog.Any("resultResponse", resultResponse))
	}
}
