package fetcher

import (
	"flag"
	"time"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/clusterconfig"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util/config"
)

type FetcherCli struct {
	*clusterconfig.ClusterConfigCli
	HttpEndpoint     string
	FetcherCachePath string
	MinimumCache     time.Duration
}

func SetFetchConfigFlags(fs *flag.FlagSet) {
	clusterconfig.SetClusterConfigCli(fs)
	fs.String("http-endpoint", "", "Force completion to fetch data from a specific http endpoint.")
	fs.String("fetcher-cache-path", "/tmp/kubectl_fzf_cache/fetcher_cache", "Location of cached resources fetched from a remote kubectl-fzf instance.")
	fs.Duration("minimum-cache", 5*time.Second, "The minimum duration after which the http endpoint will be queried to check for resource modification.")
}

func NewFetcherCli(store *config.Store) FetcherCli {
	return FetcherCli{
		ClusterConfigCli: clusterconfig.NewClusterConfigCli(store),
		FetcherCachePath: store.GetString("fetcher-cache-path", "/tmp/kubectl_fzf_cache/fetcher_cache"),
		HttpEndpoint:     store.GetString("http-endpoint", ""),
		MinimumCache:     store.GetDuration("minimum-cache", 5*time.Second),
	}
}
