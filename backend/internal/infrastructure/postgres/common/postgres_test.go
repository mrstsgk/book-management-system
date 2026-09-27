package common_test

import (
	"strings"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/common"
)

func TestConnect_UnreachableDBFailsWithoutLeakingPassword(t *testing.T) {
	t.Parallel()
	_, err := common.Connect(common.Config{
		Host:     "127.0.0.1",
		Port:     "1",
		User:     "postgres",
		Password: "super-secret-password",
		DBName:   "book_management",
		SSLMode:  "disable",
	})
	if err == nil {
		t.Fatal("expected an error for an unreachable DB")
	}
	if strings.Contains(err.Error(), "super-secret-password") {
		t.Errorf("error leaks the password: %v", err)
	}
}
