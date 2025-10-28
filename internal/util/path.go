package util

import (
	"os"
	"path/filepath"
)

const CacheOverrideEnv = "KUBECTL_FZF_CACHE_DIR"

func DefaultCacheRoot() string {
	if override := os.Getenv(CacheOverrideEnv); override != "" {
		return override
	}
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "kubectl-fzf")
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".cache", "kubectl-fzf")
	}
	return filepath.Join(os.TempDir(), "kubectl-fzf")
}
