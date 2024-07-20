package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kms-qwe/DAEC/internal/lib/logger/sl"
)

const hmacSampleSecret = "super_secret_signature"

type Server struct {
	tokenTTL   time.Duration
	log        *slog.Logger
	Port       string
	router     *http.ServeMux
	UsrStorage UsrStorage
}

type UsrStorage interface {
	SaveNewUsr(context.Context, User) (int64, error)
	IsUsrLoggin(context.Context, User) (bool, error)
	GetPassword(context.Context, string) (string, int64, error)
	GetAll(context.Context, int64) ([]Expr, error)
	GetById(context.Context, int64) (Expr, error)
	SaveNewExpr(context.Context, int64, string, string) (int64, error)
}
type User struct {
	Login    string
	Password string
}
type Expr struct {
	Id     int64
	Exp    string
	Status string
	Result float64
}
type calculateRequest struct {
	Expression string `json:"expression"`
}
type ResponseToNewExpr struct {
	ID int64 `json:"id"`
}

type ResponseToGiveAllExpr struct {
	Exprs []Expr `json:"expressions"`
}

// NewServer - конструктор для создания нового сервера
func NewServer(log *slog.Logger, port string, tokenTTL time.Duration, UsrStorage UsrStorage) *Server {
	return &Server{
		log:        log,
		Port:       port,
		router:     http.NewServeMux(),
		tokenTTL:   tokenTTL,
		UsrStorage: UsrStorage,
	}
}

// SetupRoutes - метод для настройки маршрутов
func (s *Server) SetupRoutes() {
	s.router.HandleFunc("/", s.handleRoot())
	s.router.HandleFunc("/api/v1/calculate", s.NewExprRoot())
	s.router.HandleFunc("/api/v1/expressions", s.AllExprRoot())
	s.router.HandleFunc("/api/v1/expression", s.ExprByIdRoot())
	s.router.HandleFunc("/api/v1/register", s.NewUsrRoot())
	s.router.HandleFunc("/api/v1/login", s.GiveTokenRoot())
}

// handleRoot - обработчик для корневого маршрута
func (s *Server) handleRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to the root route!")
	}
}

func (s *Server) NewUsrRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "auth.NewUsrRoot"
		log := s.log.With(slog.String("op", op))
		if r.Method != http.MethodPost {
			http.Error(w, "Метод не поддерживается", http.StatusInternalServerError)
			log.Info("Ошибка регистрации: Метод не поддерживается")
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Ошибка при чтении тела запроса", http.StatusInternalServerError)
			log.Info("Ошибка регистрации: Ошибка при чтении тела запроса", sl.Err(err))
			return
		}
		defer r.Body.Close()

		var User User

		err = json.Unmarshal(body, &User)
		if err != nil {
			http.Error(w, "Ошибка при декодировании JSON", http.StatusInternalServerError)
			log.Info("Ошибка регистрации: Ошибка при декодировании JSON", sl.Err(err))
			return
		}

		if User.Login == "" || User.Password == "" {
			http.Error(w, "Ошибка при декодировании JSON", http.StatusInternalServerError)
			log.Info("Ошибка регистрации: пустой логин или пароль", sl.Err(err))
			return
		}

		isLoggin, err := s.UsrStorage.IsUsrLoggin(context.TODO(), User)
		if err != nil {
			http.Error(w, "Ошибка при обращении к бд", http.StatusInternalServerError)
			log.Info("Ошибка регистрации: ошибка при обращении к бд", sl.Err(err))
			return
		}
		if isLoggin {
			http.Error(w, "Пользователь существует", http.StatusInternalServerError)
			log.Info("Ошибка регистрации: пользователь существует", sl.Err(err))
			return
		}
		if _, err := s.UsrStorage.SaveNewUsr(context.TODO(), User); err != nil {
			http.Error(w, "Ошибка при обращении к бд", http.StatusInternalServerError)
			log.Info("Ошибка регистрации: ошибка при сохранении в  бд", sl.Err(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Пользователь зарегистрирован. Добро пожаловать!\n"))
	}
}
func (s *Server) GiveTokenRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "auth.NewUsrRoot"
		log := s.log.With(slog.String("op", op))
		if r.Method != http.MethodGet {
			http.Error(w, "Метод не поддерживается", http.StatusInternalServerError)
			log.Info("Вывод списка выражений не произведен: Метод не поддерживается")
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Ошибка при чтении тела запроса", http.StatusInternalServerError)
			log.Info("Ошибка регистрации: Ошибка при чтении тела запроса", sl.Err(err))
			return
		}
		defer r.Body.Close()

		var User User

		err = json.Unmarshal(body, &User)
		if err != nil {
			http.Error(w, "Ошибка при декодировании JSON", http.StatusInternalServerError)
			log.Info("Ошибка регистрации: Ошибка при декодировании JSON", sl.Err(err))
			return
		}

		if User.Login == "" || User.Password == "" {
			http.Error(w, "Ошибка при декодировании JSON", http.StatusInternalServerError)
			log.Info("Ошибка регистрации: пустой логин или пароль", sl.Err(err))
			return
		}

		pass, id, err := s.UsrStorage.GetPassword(context.TODO(), User.Login)
		if err != nil {
			http.Error(w, "Ошибка при обращении к бд", http.StatusInternalServerError)
			log.Info("Ошибка регистрации: ошибка при обращении к бд", sl.Err(err))
			return
		}
		if User.Password != pass {
			http.Error(w, "Ошибка при обращении к бд", http.StatusInternalServerError)
			log.Info("Ошибка регистрации: неверный пароль", slog.Any("User", User), slog.String("acPass", pass))
			return
		}
		now := time.Now()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"login": User.Login,
			"id":    id,
			"nbf":   now.Unix(),
			"exp":   now.Add(s.tokenTTL).Unix(),
			"iat":   now.Unix(),
		})

		tokenString, err := token.SignedString([]byte(hmacSampleSecret))
		if err != nil {
			http.Error(w, "Could not generate token", http.StatusInternalServerError)
			log.Info("Could not generate token")
			return
		}

		// Возвращаем JWT токен в теле ответа
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	}
}

func (s *Server) NewExprRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "auth.NewExprRoot"
		log := s.log.With(slog.String("op", op))

		if r.Method != http.MethodPost {
			http.Error(w, "Метод не поддерживается", http.StatusInternalServerError)
			log.Info("Не принято на вычисление: Метод не поддерживается")
			return
		}

		isValid, userID, err := s.validateJWTToken(r)
		if err != nil || !isValid {
			if err != nil {
				http.Error(w, "Ошибка валидации токена", http.StatusInternalServerError)
				log.Info("Не принято на вычисление: ошибка при валидации токена", sl.Err(err))
				return
			}
			http.Error(w, "Ошибка валидации токена", http.StatusInternalServerError)
			log.Info("Не принято на вычисление: токен не валиден", sl.Err(err))
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Ошибка при чтении тела запроса", http.StatusInternalServerError)
			log.Info("Не принято на вычисление: Ошибка при чтении тела запроса", sl.Err(err))
			return
		}
		defer r.Body.Close()

		var data calculateRequest

		err = json.Unmarshal(body, &data)
		if err != nil {
			http.Error(w, "Ошибка при декодировании JSON", http.StatusInternalServerError)
			log.Info("Не принято на вычисление: Ошибка при декодировании JSON", sl.Err(err))
			return
		}

		polishExpr, err := infixToPostfix(data.Expression)
		if err != nil {
			http.Error(w, "Невалидные данные", http.StatusUnprocessableEntity)
			log.Info("Не принято на вычисление: Невалидные данные", sl.Err(err), slog.Any("data", data))
			return
		}

		id, err := s.UsrStorage.SaveNewExpr(context.TODO(), userID, data.Expression, polishExpr)
		if err != nil {
			http.Error(w, "Ошибка при обращении к бд", http.StatusInternalServerError)
			log.Info("Не принято на вычисление: ошибка при обращении к бд", sl.Err(err))
			return
		}

		response := ResponseToNewExpr{ID: id}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			log.Info("Не принято на вычисление: ошибка при записи id", sl.Err(err))
		}
		log.Info("Принято на вычисление")
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) AllExprRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "auth.AllExprRoot"
		log := s.log.With(slog.String("op", op))
		if r.Method != http.MethodGet {
			http.Error(w, "Метод не поддерживается", http.StatusInternalServerError)
			log.Info("Выражения не отданы: Метод не поддерживается")
			return
		}

		isValid, userID, err := s.validateJWTToken(r)
		if err != nil || !isValid {
			if err != nil {
				http.Error(w, "Ошибка валидации токена", http.StatusInternalServerError)
				log.Info("Выражения не отданы: ошибка при валидации токена", sl.Err(err))
				return
			}
			http.Error(w, "Ошибка валидации токена", http.StatusInternalServerError)
			log.Info("Выражения не отданы: токен не валиден", sl.Err(err))
			return
		}

		exprs, err := s.UsrStorage.GetAll(context.TODO(), userID)

		if err != nil {
			http.Error(w, "Ошибка при обращении к бд", http.StatusInternalServerError)
			log.Info("Выражения не отданы: ошибка при обращении к бд", sl.Err(err))
			return
		}

		ans := ResponseToGiveAllExpr{Exprs: exprs}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(ans); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			log.Info("Выражения не отданы: ошибка при записи id", sl.Err(err))
		}
		log.Info("Выражения отданы", slog.Any("exprs", exprs))
		w.WriteHeader(http.StatusOK)

	}
}

func (s *Server) ExprByIdRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "auth.ExprByIdRoot"
		log := s.log.With(slog.String("op", op))
		if r.Method != http.MethodGet {
			http.Error(w, "Метод не поддерживается", http.StatusInternalServerError)
			log.Info("Выражения не отданы: Метод не поддерживается")
			return
		}

		isValid, _, err := s.validateJWTToken(r)
		if err != nil || !isValid {
			if err != nil {
				http.Error(w, "Ошибка валидации токена", http.StatusInternalServerError)
				log.Info("Выражения не отданы: ошибка при валидации токена", sl.Err(err))
				return
			}
			http.Error(w, "Ошибка валидации токена", http.StatusInternalServerError)
			log.Info("Выражения не отданы: токен не валиден", sl.Err(err))
			return
		}

		queryParams := r.URL.Query()
		id, err := strconv.Atoi(queryParams.Get("id"))
		if err != nil {
			http.Error(w, "ошибка получения id", http.StatusInternalServerError)
			log.Info("Выражения не отданы: ошибка при получении id", sl.Err(err))
			return
		}

		expr, err := s.UsrStorage.GetById(context.TODO(), int64(id))

		if err != nil {
			http.Error(w, "Ошибка при обращении к бд", http.StatusInternalServerError)
			log.Info("Выражения не отданы: ошибка при обращении к бд", sl.Err(err))
			return
		}

		ans := ResponseToGiveAllExpr{Exprs: []Expr{expr}}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(ans); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			log.Info("Выражения не отданы: ошибка при записи id", sl.Err(err))
		}
		log.Info("Выражения отданы", slog.Any("expr", expr))
		w.WriteHeader(http.StatusOK)
	}
}

// MustRun - метод для запуска сервера
func (s *Server) MustRun() {
	s.SetupRoutes()
	s.log.Info("Starting server on port", slog.String("port", s.Port))
	if err := http.ListenAndServe(s.Port, s.router); err != nil {
		s.log.Error("http server is not setuped", sl.Err(err))
		panic("http server is not setuped")
	}
}

func (s *Server) validateJWTToken(r *http.Request) (bool, int64, error) {
	const op = "auth.validateJWTToken"
	log := s.log.With(slog.String("op", op))
	tokenString := getTokenFromHeader(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(hmacSampleSecret), nil
	})

	if err != nil {
		log.Info("не удалось проверить токен", sl.Err(err))
		return false, 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println("User name: ", claims["name"])
		return true, claims["id"].(int64), nil
	}

	return false, 0, nil
}
func getTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}
