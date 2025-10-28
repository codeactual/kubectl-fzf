package store

import (
	"flag"
	"time"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/clusterconfig"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util/config"
)

type StoreConfigCli struct {
	*clusterconfig.ClusterConfigCli
	TimeBetweenFullDump time.Duration
}

func SetStoreConfigCli(fs *flag.FlagSet) {
	clusterconfig.SetClusterConfigCli(fs)
	fs.Duration("time-between-full-dump", 10*time.Second, "Buffer changes and only do full dump every x secondes")
}

func NewStoreConfigCli(store *config.Store) StoreConfigCli {
	return StoreConfigCli{
		ClusterConfigCli:    clusterconfig.NewClusterConfigCli(store),
		TimeBetweenFullDump: store.GetDuration("time-between-full-dump", 10*time.Second),
	}
}
