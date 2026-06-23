package model

import "time"

type User struct {
	ID             int       // идентификатор
	Login          string    // логин
	PasswordHash   string    // хеш пароля
	CurrentBalance float64   // текущий баланс баллов
	WithdrawnTotal float64   // суммарно списано за всё время
	CreatedAt      time.Time // дата регистрации
}
