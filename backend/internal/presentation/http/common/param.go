package common

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// ParseID は正の整数の path パラメータを読む。それ以外は UseCase を呼ぶ前に 400 にする。
func ParseID(c echo.Context, name string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid "+name)
	}
	return id, nil
}
