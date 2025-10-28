package util

import (
	"os"

	log "github.com/codeactual/kubectl-fzf/v4/internal/logger"
)

func RemoveTempDir(tempDir string) {
	err := os.RemoveAll(tempDir)
	log.Warnf("Couldn't remove tempdir %s: %s", tempDir, err)
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
