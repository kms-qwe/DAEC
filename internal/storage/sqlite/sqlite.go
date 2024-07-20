package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/kms-qwe/DAEC/internal/app/auth"
	_ "github.com/mattn/go-sqlite3"
)

type OrchStorage struct {
	db *sql.DB
}

func NewOrchStorage(path string) (*OrchStorage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("can't open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("can't open database: %w", err)
	}
	return &OrchStorage{db: db}, nil
}

func (s *OrchStorage) GetExpr(ctx context.Context) (int64, string, error) {
	q := `SELECT expr_id, polish_expr FROM expressions WHERE status = "computing" LIMIT 1`

	var exprID int64
	var polishExpr string

	err := s.db.QueryRowContext(ctx, q).Scan(&exprID, &polishExpr)
	if err == sql.ErrNoRows {
		return 0, "", fmt.Errorf("no computing expr in db: %w", err)
	}
	if err != nil {
		return 0, "", fmt.Errorf("can't get computing expr: %w", err)
	}

	return exprID, polishExpr, nil
}

func (s *OrchStorage) SaveExpr(ctx context.Context, exprID int64, expr string) error {

	var isNumber = func(str string) bool {
		_, err := strconv.ParseFloat(str, 64)
		return err == nil
	}
	var q string

	if !isNumber(expr) {
		q = `UPDATE expressions SET expr = ?, status = "done", result = ? WHERE expr_id = ?`

		res, _ := strconv.ParseFloat(expr, 64)

		_, err := s.db.ExecContext(ctx, q, expr, res, exprID)

		if err != nil {
			return fmt.Errorf("can't update expr with result: %w", err)
		}

		return nil
	}

	q = `UPDATE expressions SET expr = ? WHERE expr_id = ?`

	_, err := s.db.ExecContext(ctx, q, expr, exprID)

	if err != nil {
		return fmt.Errorf("can't update expr: %w", err)
	}

	return nil
}

type AuthStorage struct {
	db *sql.DB
}

func NewAuthStorage(path string) (*AuthStorage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("can't open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("can't open database: %w", err)
	}
	return &AuthStorage{db: db}, nil
}

func (s *AuthStorage) SaveNewUsr(ctx context.Context, user auth.User) (int64, error) {
	q := `INSERT INTO users (login, password) VALUES (?, ?)`

	result, err := s.db.ExecContext(ctx, q, user.Login, user.Password)
	if err != nil {
		return 0, fmt.Errorf("cant't save user: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("cant't save user: %w", err)
	}

	return id, nil
}

func (s *AuthStorage) IsUsrLoggin(ctx context.Context, user auth.User) (bool, error) {
	q := `SELECT user_id FROM users WHERE login = ?`

	var userID int64

	err := s.db.QueryRowContext(ctx, q, user.Login).Scan(&userID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("can't check login in db: %w", err)
	}

	return true, nil

}

func (s *AuthStorage) GetPassword(ctx context.Context, login string) (string, int64, error) {
	q := `SELECT user_id, password FROM users WHERE login = ?`

	var userID int64
	var pass string
	err := s.db.QueryRowContext(ctx, q, login).Scan(&userID, &pass)
	if err == sql.ErrNoRows {
		return "", 0, fmt.Errorf("no such user in db: %w", err)
	}
	if err != nil {
		return "", 0, fmt.Errorf("can't get password by login: %w", err)
	}

	return pass, userID, nil
}

func (s *AuthStorage) GetById(ctx context.Context, exprID int64) (auth.Expr, error) {
	q := `SELECT expr_id, expr, status, result FROM expressions WHERE expr_id = ?`

	var ans auth.Expr

	err := s.db.QueryRowContext(ctx, q, exprID).Scan(&ans.Id, &ans.Exp, &ans.Status, &ans.Result)
	if err == sql.ErrNoRows {
		return auth.Expr{}, fmt.Errorf("no such expr in db: %w", err)
	}
	if err != nil {
		return auth.Expr{}, fmt.Errorf("can't get expr: %w", err)
	}

	return ans, nil
}
func (s *AuthStorage) SaveNewExpr(ctx context.Context, userID int64, expr string, polishExpr string) (int64, error) {
	q := `INSERT INTO expressions (expr, polish_expr, user_id) VALUE (?, ?, ?)`

	result, err := s.db.ExecContext(ctx, q, expr, polishExpr, userID)
	if err != nil {
		return 0, fmt.Errorf("cant't save new expression: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("cant't save new expression: %w", err)
	}

	return id, nil
}
func (s *AuthStorage) GetAll(ctx context.Context, userID int64) ([]auth.Expr, error) {
	q := `SELECT expr_id, expr, status, result FROM expressions WHERE user_id = ?`

	var ans []auth.Expr

	rows, err := s.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("can't get all expressions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		expr := auth.Expr{}
		err := rows.Scan(&expr.Id, &expr.Exp, &expr.Status, &expr.Result)
		if err != nil {
			return nil, fmt.Errorf("can't get all expressions: %w", err)
		}
		ans = append(ans, expr)
	}

	return ans, nil
}

type InitStorage struct {
	db *sql.DB
}

func NewInitStorage(path string) (*InitStorage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("can't open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("can't open database: %w", err)
	}
	return &InitStorage{db: db}, nil
}

func (s *InitStorage) Init(ctx context.Context) error {
	usersTable := `CREATE TABLE IF NOT EXISTS users (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    login TEXT UNIQUE CHECK(login != ""),
    password TEXT CHECK(password != "")
	);`

	exprTable := `CREATE TABLE IF NOT EXISTS expressions (
    expr_id INTEGER PRIMARY KEY AUTOINCREMENT,
    expr TEXT,
    polish_expr TEXT,
    status TEXT DEFAULT 'computing',
    result DOUBLE,
    user_id INTEGER,
    FOREIGN KEY (user_id) REFERENCES users (user_id)
	);`

	if _, err := s.db.ExecContext(ctx, usersTable); err != nil {
		return err
	}

	if _, err := s.db.ExecContext(ctx, exprTable); err != nil {
		return err
	}

	return nil
}

func (s *InitStorage) Drop(ctx context.Context) error {
	usersTable := `DROP TABLE IF EXISTS users;`

	exprTable := `DROP TABLE IF EXISTS expressions;`

	if _, err := s.db.ExecContext(ctx, exprTable); err != nil {
		return err
	}

	if _, err := s.db.ExecContext(ctx, usersTable); err != nil {
		return err
	}

	return nil
}
