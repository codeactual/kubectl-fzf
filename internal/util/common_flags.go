package util

import (
	"flag"
	"os"
	"runtime/pprof"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/util/config"
)

func SetCommonCliFlags(fs *flag.FlagSet, defaultLogLevel string) {
	fs.String("log-level", defaultLogLevel, "Log level to use")
	fs.String("cpu-profile", "", "Destination file for cpu profiling")
	fs.String("mem-profile", "", "Destination file for memory profiling")
}

func CommonInitialization(store *config.Store) {
	configureLog(store)
	cpuProfile := store.GetString("cpu-profile", "")
	if cpuProfile == "" {
		return
	}
	f, err := os.Create(cpuProfile)
	if err != nil {
		FatalIf(err)
	}
	pprof.StartCPUProfile(f)
}
