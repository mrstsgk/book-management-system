package main

import "testing"

// run is unexported, so this test lives in package main.

func TestRun_FailsFastOnInvalidConfig(t *testing.T) {
	t.Setenv("STAGE", "local")
	t.Setenv("LOG_LEVEL", "verbose")

	if err := run(); err == nil {
		t.Fatal("expected an error for an invalid LOG_LEVEL")
	}
}

func TestRun_FailsWhenDBIsUnreachable(t *testing.T) {
	t.Setenv("STAGE", "local")
	t.Setenv("LOG_LEVEL", "error")
	t.Setenv("DB_HOST", "127.0.0.1")
	t.Setenv("DB_PORT", "1")

	if err := run(); err == nil {
		t.Fatal("expected an error when the DB is unreachable")
	}
}
