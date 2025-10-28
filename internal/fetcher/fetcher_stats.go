package fetcher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/store"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"

	log "github.com/bonnefoa/kubectl-fzf/v3/internal/logger"
	"github.com/pkg/errors"
)

func (f *Fetcher) getStatsFromHttpServer(ctx context.Context, url string) ([]*store.Stats, error) {
	log.Debugf("Fetching stats from %s", url)
	_, body, err := util.GetFromHttpServer(url)
	if err != nil {
		return nil, errors.Wrap(err, "error on http get")
	}
	stats := make([]*store.Stats, 0)
	err = json.Unmarshal(body, &stats)
	return stats, err
}

func (f *Fetcher) GetStats(ctx context.Context) ([]*store.Stats, error) {
	// TODO Handle local file
	if util.IsAddressReachable(f.httpEndpoint) {
		url := fmt.Sprintf("http://%s/%s", f.httpEndpoint, "stats")
		return f.getStatsFromHttpServer(ctx, url)
	}
	if f.httpEndpoint == "" {
		return nil, fmt.Errorf("http endpoint not configured; run kubectl-fzf-server locally or provide --http-endpoint")
	}
	return nil, fmt.Errorf("http endpoint %s is not reachable", f.httpEndpoint)
}
