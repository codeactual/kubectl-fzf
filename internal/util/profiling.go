package util

import (
	"os"
	"runtime"
	"runtime/pprof"

	log "github.com/bonnefoa/kubectl-fzf/v3/internal/logger"
	"github.com/spf13/viper"
)

func DoMemoryProfile() {
	memProfile := viper.GetString("mem-profile")
	if memProfile != "" {
		f, err := os.Create(memProfile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}
