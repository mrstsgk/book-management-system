package common

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// ParseID reads a positive integer path parameter; anything else is a 400 before any usecase runs.
func ParseID(c echo.Context, name string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid "+name)
	}
	return id, nil
}
