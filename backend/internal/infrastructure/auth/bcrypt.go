package auth

import (
	"golang.org/x/crypto/bcrypt"

	domainauth "github.com/mrstsgk/book-management-system/backend/internal/domain/auth"
)

// hashCost は bcrypt のコスト。既定（10）より高くして総当たりを遅くする。ログインは 1 人が 1 日に数回なので待ち時間は問題にならない。
const hashCost = 12

type bcryptVerifier struct{}

func NewBcryptVerifier() domainauth.PasswordVerifier {
	return bcryptVerifier{}
}

// Matches は bcrypt で照合する。壊れたハッシュはエラーになるが、それも「一致しない」として扱う。
func (bcryptVerifier) Matches(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// HashPassword は ADMIN_PASSWORD_HASH に置く値を作る（cmd/hashpw が使う）。
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), hashCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
