package jwt

import (
	"time"

	"github.com/kms-qwe/DAEC/internal/domain/models"
)

func NewToken(
	user models.User,
	duration time.Duration,
) (string, error) {
	panic("implement NewToken")
}
