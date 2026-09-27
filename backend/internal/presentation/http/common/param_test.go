package common_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
)

func TestParseID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		raw     string
		want    int64
		wantErr bool
	}{
		{name: "正の整数は有効", raw: "42", want: 42},
		{name: "1は有効", raw: "1", want: 1},
		{name: "0は400", raw: "0", wantErr: true},
		{name: "負の値は400", raw: "-1", wantErr: true},
		{name: "数値以外は400", raw: "abc", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := echo.New().NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
			c.SetParamNames("id")
			c.SetParamValues(tt.raw)

			got, err := common.ParseID(c, "id")
			if tt.wantErr {
				var he *echo.HTTPError
				if !errors.As(err, &he) || he.Code != http.StatusBadRequest {
					t.Fatalf("err = %v, want a 400 HTTPError", err)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("got (%d, %v), want (%d, nil)", got, err, tt.want)
			}
		})
	}
}
