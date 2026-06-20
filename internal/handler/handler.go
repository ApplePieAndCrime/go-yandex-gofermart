package handler

import (
	"io"
	"net/http"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/middleware"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handler struct {
	services *service.Service
	logger   *zap.SugaredLogger
}

func NewHandler(services *service.Service, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		services: services,
		logger:   logger.With("component", "handler"),
	}
}

func (h Handler) InitRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Get("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "Сервер упешно запущен")
	})

	r.Route("/api/user/", func(r chi.Router) {
		r.Post("/login", h.Login)       // POST /users/login - регистрация пользователя
		r.Post("/register", h.Register) // POST /users/register - аутентификация пользователя

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(h.logger))

			r.Route("/orders", func(r chi.Router) {
				r.Post("/", h.UploadOrder) // POST /users/orders - загрузка пользователем номера заказа для расчёта
				r.Get("/", h.GetOrders)    // GET /users/orders - получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях
			})

			r.Route("/balance", func(r chi.Router) {
				r.Get("/", h.GetBalance)          // GET /users/balance - получение текущего баланса счёта баллов лояльности пользователя
				r.Post("/withdrawal", h.Withdraw) // POST /users/balance/withdrawal - запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа
			})

			r.Get("/withdrawals", h.GetWithdrawals) // GET /users/withdrawals - получение информации о выводе средств с накопительного счёта пользователем
		})
	})

	return r
}
