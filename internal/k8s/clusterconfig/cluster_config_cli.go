package clusterconfig

import (
	"flag"

	"github.com/codeactual/kubectl-fzf/v4/internal/util/config"
)

type ClusterConfigCli struct {
	ClusterName string
	CacheDir    string
}

func SetClusterConfigCli(fs *flag.FlagSet) {
	fs.String("cache-dir", "/tmp/kubectl_fzf_cache/", "Cache dir location.")
}

func NewClusterConfigCli(store *config.Store) *ClusterConfigCli {
	return &ClusterConfigCli{
		CacheDir: store.GetString("cache-dir", "/tmp/kubectl_fzf_cache/"),
	}
}
