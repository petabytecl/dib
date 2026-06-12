package config_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/petabytecl/dib/config"
)

func TestSensitiveDefaultDiagnosticsDoNotExposeFakeCorpus(t *testing.T) {
	t.Parallel()

	corpus := []string{
		"dib_fake_secret_value",
		"dib_fake_password_value",
		"dib_fake_token_value",
	}

	for _, raw := range corpus {
		t.Run(raw, func(t *testing.T) {
			_, err := config.NewSet(config.Define("token", config.KindInt, "token", config.Default(raw), config.Sensitive()))
			if err == nil {
				t.Fatal("NewSet returned nil error")
			}
			if !errors.Is(err, config.ErrInvalidDefault) {
				t.Fatalf("errors.Is(err, ErrInvalidDefault) = false; err=%v", err)
			}
			if strings.Contains(err.Error(), raw) {
				t.Fatalf("sensitive diagnostic leaked raw corpus value: %v", err)
			}
			if !strings.Contains(err.Error(), "token") {
				t.Fatalf("sensitive diagnostic omitted key context: %v", err)
			}
			if !strings.Contains(err.Error(), config.SourceDefault) {
				t.Fatalf("sensitive diagnostic omitted default provenance: %v", err)
			}
		})
	}
}

func TestNonSensitiveDefaultDiagnosticsMayExposeOrdinaryInvalidValues(t *testing.T) {
	t.Parallel()

	const ordinaryInvalid = "not-an-int"
	_, err := config.NewSet(config.Define("workers", config.KindInt, "workers", config.Default(ordinaryInvalid)))
	if err == nil {
		t.Fatal("NewSet returned nil error")
	}
	if !errors.Is(err, config.ErrInvalidDefault) {
		t.Fatalf("errors.Is(err, ErrInvalidDefault) = false; err=%v", err)
	}
	if !strings.Contains(err.Error(), ordinaryInvalid) {
		t.Fatalf("non-sensitive diagnostic omitted ordinary invalid value: %v", err)
	}
}
