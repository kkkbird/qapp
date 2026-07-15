package qapp

import (
	"path/filepath"
	"testing"
)

func TestDeriveOverrideConfigFile(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"default yml", "app.yml", "app.override.yml"},
		{"no extension", "app", "app.override"},
		{"with dir", "config/prod.yaml", filepath.Join("config", "prod.override.yaml")},
		{"dotfile only", ".env", ".env.override"},
		{"nested dir no ext", "configs/app", filepath.Join("configs", "app.override")},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := deriveOverrideConfigFile(c.in); got != c.want {
				t.Errorf("deriveOverrideConfigFile(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
