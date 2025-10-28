package util

import (
	log "github.com/codeactual/kubectl-fzf/v4/internal/logger"
	"github.com/codeactual/kubectl-fzf/v4/internal/util/config"
)

type LogConf struct {
	LogLevel log.Level
}

func getLogConf(store *config.Store) LogConf {
	logLevelStr := store.GetString("log-level", "info")
	logLevel, err := log.ParseLevel(logLevelStr)
	FatalIf(err)

	return LogConf{LogLevel: logLevel}
}

func configureLog(store *config.Store) {
	logConf := getLogConf(store)
	log.Debugf("Setting log level %v", logConf.LogLevel)
	log.SetLevel(logConf.LogLevel)
}
