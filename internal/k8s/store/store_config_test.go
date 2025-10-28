package store

import (
	"testing"
	"time"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/clusterconfig"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
)

func TestFileStoreExists(t *testing.T) {
	c := &StoreConfigCli{
		ClusterConfigCli: &clusterconfig.ClusterConfigCli{
			ClusterName: "minikube",
			CacheDir:    "./testdata",
		}, TimeBetweenFullDump: 1 * time.Second}
	s := NewStoreConfig(c)
	if !s.FileStoreExists(resources.ResourceTypePod) {
		t.Errorf("expected pod file store to exist")
	}
	if s.FileStoreExists(resources.ResourceTypeApiResource) {
		t.Errorf("expected api resource file store to not exist")
	}
}
