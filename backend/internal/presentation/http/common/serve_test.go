package common_test

import (
	"net"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
)

func TestServe_ReturnsErrorWhenAddressIsInUse(t *testing.T) {
	silenceLog(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	if err := common.Serve(common.NewEcho(), ln.Addr().String()); err == nil {
		t.Fatal("expected an error for an address already in use")
	}
}
