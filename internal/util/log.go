package util

import (
	log "github.com/bonnefoa/kubectl-fzf/v3/internal/logger"
	"github.com/spf13/viper"
)

type LogConf struct {
	LogLevel log.Level
}

func getLogConf() LogConf {
	logLevelStr := viper.GetString("log-level")
	logLevel, err := log.ParseLevel(logLevelStr)
	FatalIf(err)

	l := LogConf{}
	l.LogLevel = logLevel
	return l
}

func configureLog() {
	logConf := getLogConf()
	log.Debugf("Setting log level %v", logConf.LogLevel)
	log.SetLevel(logConf.LogLevel)
}
