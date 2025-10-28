package util

import (
	"os"
	"runtime"
	"runtime/pprof"

	log "github.com/codeactual/kubectl-fzf/v4/internal/logger"
	"github.com/codeactual/kubectl-fzf/v4/internal/util/config"
)

func DoMemoryProfile(store *config.Store) {
	memProfile := store.GetString("mem-profile", "")
	if memProfile == "" {
		return
	}
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
