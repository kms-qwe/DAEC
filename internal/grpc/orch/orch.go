package orch

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/kms-qwe/DAEC/internal/lib/logger/sl"
	daecv1 "github.com/kms-qwe/DAEC/internal/protos/gen/go/daec"
	"google.golang.org/grpc"
)

type ServerApi struct {
	daecv1.UnimplementedOrchServiceServer
	TaskPull *TaskPuller
}

type TaskPuller struct {
	Log         *slog.Logger
	ChToAgent   chan *daecv1.TaskResponse
	ChFromAgent chan *daecv1.ResultRequest
	ExprID      int64
	Expr        string
	ExpStrg     ExpStorage
}

type ExpStorage interface {
	GetExpr(ctx context.Context) (int64, string, error)
	SaveExpr(ctx context.Context, exprID int64, expr string) error
}

func Register(gRPC *grpc.Server, TaskPull *TaskPuller) {
	daecv1.RegisterOrchServiceServer(gRPC, &ServerApi{TaskPull: TaskPull})
}

func (s *ServerApi) GiveTask(ctx context.Context, req *daecv1.TaskRequest) (*daecv1.TaskResponse, error) {
	return <-s.TaskPull.ChToAgent, nil
}

func (s *ServerApi) GetResult(ctx context.Context, req *daecv1.ResultRequest) (*daecv1.ResultResponse, error) {
	s.TaskPull.ChFromAgent <- req
	return &daecv1.ResultResponse{}, nil
}
func (t *TaskPuller) FakeEval() {
	const op = "orch.FakeEval"
	log := t.Log.With(
		slog.String("op", op),
	)
	log.Info("Eval starts")
	go func() {
		for {
			t.ChToAgent <- &daecv1.TaskResponse{}
			log.Info("Задача отослана")
			time.Sleep(1 * time.Second)
		}
	}()
	for m := range t.ChFromAgent {
		log.Info("get m", slog.Any("m", m))
	}
}
func (t *TaskPuller) Eval() {
	const op = "orch.Eval"
	log := t.Log.With(
		slog.String("op", op),
	)
	log.Info("Eval starts")

	ctx := context.Background()
	var err error

	for {
		// time.Sleep(5 * time.Second)
		t.ExprID, t.Expr, err = t.ExpStrg.GetExpr(ctx)
		if err != nil {
			log.Info("falied to get expr", sl.Err(err))
			time.Sleep(10 * time.Second)
			continue
		}

		log.Info("get expr", slog.String("expr", t.Expr))

		//Считаем части, которые можно выполнить параллельно
		elementsOfExpr := strings.Fields(t.Expr)
		numOp := 0
		res := map[int]float64{}

		for i := range len(elementsOfExpr) - 2 {
			if isNumber(elementsOfExpr[i]) && isNumber(elementsOfExpr[i+1]) && isOperator([]rune(elementsOfExpr[i+2])[0]) {
				numOp += 1
			}
		}

		go func() {
			cnt := 0
			for i := range len(elementsOfExpr) - 2 {
				if isNumber(elementsOfExpr[i]) && isNumber(elementsOfExpr[i+1]) && isOperator([]rune(elementsOfExpr[i+2])[0]) {
					cnt += 1
					tsk := &daecv1.TaskResponse{Id: int64(cnt)}
					tsk.Arg1, _ = strconv.ParseFloat(elementsOfExpr[i], 64)
					tsk.Arg2, _ = strconv.ParseFloat(elementsOfExpr[i+1], 64)
					tsk.Operation = elementsOfExpr[i+2]
					log.Info("отправлено в chToAgent", slog.Any("task", tsk))
					t.ChToAgent <- tsk
				}
			}
		}()

		log.Info("ожидается результат", slog.Int("len(res)", len(res)), slog.Int("numOp", numOp))
		for len(res) != numOp {
			r := <-t.ChFromAgent
			log.Info(
				"Получен новый результат",
				slog.Int("len(res)",
					len(res)+1),
				slog.Int("numOp", numOp),
				slog.Int64("номер результата", r.Id),
				slog.Float64("Результат", r.Result),
			)
			res[int(r.GetId())] = r.GetResult()
		}

		cnt := 0
		for i := range len(elementsOfExpr) - 2 {
			if isNumber(elementsOfExpr[i]) && isNumber(elementsOfExpr[i+1]) && isOperator([]rune(elementsOfExpr[i+2])[0]) {
				cnt += 1
				r := res[cnt]
				strRes := strconv.FormatFloat(r, 'f', 6, 64)
				elementsOfExpr[i], elementsOfExpr[i+1], elementsOfExpr[i+2] = strRes, "", ""
			}
		}

		expr := strings.Join(elementsOfExpr, " ")

		err = t.ExpStrg.SaveExpr(ctx, t.ExprID, expr)
		if err != nil {
			log.Info("falied to save new expr", sl.Err(err))
		}
		log.Info("новое выражение сохранено", slog.String("NewExpr", expr))
	}
}

func isNumber(str string) bool {
	_, err := strconv.ParseFloat(str, 64)
	return err == nil
}

func isOperator(c rune) bool {
	return c == '+' || c == '-' || c == '*' || c == '/' || c == '(' || c == ')'
}
