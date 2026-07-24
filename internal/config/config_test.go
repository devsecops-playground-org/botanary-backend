package config

import "testing"

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("PORT", "")

	cfg := Load()
	if cfg.Env != "development" || cfg.Port != "8080" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if cfg.IsProduction() {
		t.Fatal("development config reported itself as production")
	}
}

func TestMissingProductionSecrets(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want int
	}{
		{"development needs nothing", Config{Env: "development"}, 0},
		{"production fully configured", Config{Env: "production", JWTSecret: "s", DatabaseURL: "d"}, 0},
		{"production missing jwt", Config{Env: "production", DatabaseURL: "d"}, 1},
		{"production missing both", Config{Env: "production"}, 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := len(tc.cfg.MissingProductionSecrets()); got != tc.want {
				t.Fatalf("got %d missing secrets, want %d", got, tc.want)
			}
		})
	}
}

func TestGetenvPrefersTheEnvironment(t *testing.T) {
	t.Setenv("PORT", "9999")
	if got := Load().Port; got != "9999" {
		t.Fatalf("got %q, want 9999", got)
	}
}
