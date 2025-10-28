package fetcher

import (
	"context"
	"fmt"
	"time"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/clusterconfig"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"
)

// Fetcher defines configuration to fetch completion datas
type Fetcher struct {
	clusterconfig.ClusterConfig
	fetcherCachePath string
	httpEndpoint     string
	minimumCache     time.Duration
	fetcherState     FetcherState
}

func NewFetcher(fetchConfigCli *FetcherCli) *Fetcher {
	f := Fetcher{
		ClusterConfig:    clusterconfig.NewClusterConfig(fetchConfigCli.ClusterConfigCli),
		httpEndpoint:     fetchConfigCli.HttpEndpoint,
		fetcherCachePath: fetchConfigCli.FetcherCachePath,
		minimumCache:     fetchConfigCli.MinimumCache,
		fetcherState:     *newFetcherState(fetchConfigCli.FetcherCachePath),
	}
	return &f
}

func (f *Fetcher) LoadFetcherState() error {
	err := f.LoadClusterConfig()
	if err != nil {
		return err
	}
	return f.fetcherState.loadStateFromDisk()
}

func (f *Fetcher) SaveFetcherState() error {
	return f.fetcherState.writeToDisk()
}

func loadResourceFromFile(filePath string) (map[string]resources.K8sResource, error) {
	resources := map[string]resources.K8sResource{}
	err := util.LoadGobFromFile(&resources, filePath)
	return resources, err
}

func (f *Fetcher) GetResources(ctx context.Context, r resources.ResourceType) (map[string]resources.K8sResource, error) {
	resources, err := f.checkLocalFiles(r)
	if resources != nil || err != nil {
		return resources, err
	}

	// Check for recent cache
	resources, err = f.checkRecentCache(r)
	if resources != nil || err != nil {
		return resources, err
	}

	// Fetch remote
	if util.IsAddressReachable(f.httpEndpoint) {
		return f.loadResourceFromHttpServer(f.httpEndpoint, r)
	}
	if f.httpEndpoint == "" {
		return nil, fmt.Errorf("http endpoint not configured; run kubectl-fzf-server locally or provide --http-endpoint")
	}
	return nil, fmt.Errorf("http endpoint %s is not reachable", f.httpEndpoint)
}
