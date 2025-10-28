package httpservertest

import (
	"context"
	"os"
	"testing"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/fetcher/fetchertest"
	log "github.com/bonnefoa/kubectl-fzf/v3/internal/logger"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)
	code := m.Run()
	os.Exit(code)
}

func TestHttpServerApiCompletion(t *testing.T) {
	fzfHttpServer := StartTestHttpServer(t)
	f, _ := fetchertest.GetTestFetcher(t, "nothing", fzfHttpServer.Port)
	ctx := context.Background()
	s, err := f.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}
	if len(s) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(s))
	}
}
