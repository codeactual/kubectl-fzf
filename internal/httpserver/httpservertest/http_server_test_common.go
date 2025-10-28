package httpservertest

import (
	"context"
	"testing"

	"github.com/codeactual/kubectl-fzf/v4/internal/httpserver"
	"github.com/codeactual/kubectl-fzf/v4/internal/k8s/clusterconfig"
	"github.com/codeactual/kubectl-fzf/v4/internal/k8s/store"
	"github.com/codeactual/kubectl-fzf/v4/internal/k8s/store/storetest"
)

func GetTestClusterConfigCli() *clusterconfig.ClusterConfigCli {
	return &clusterconfig.ClusterConfigCli{ClusterName: "minikube", CacheDir: "./testdata"}
}

func GetTestStoreConfigCli() *store.StoreConfigCli {
	return &store.StoreConfigCli{ClusterConfigCli: GetTestClusterConfigCli()}
}

func StartTestHttpServer(t *testing.T) *httpserver.FzfHttpServer {
	ctx := context.Background()
	storeConfigCli := GetTestStoreConfigCli()
	storeConfig := store.NewStoreConfig(storeConfigCli)
	_, podStore := storetest.GetTestPodStore(t)
	h := &httpserver.HttpServerConfigCli{ListenAddress: "localhost:0", Debug: false}
	fzfHttpServer, err := httpserver.StartHttpServer(ctx, h, storeConfig, []*store.Store{podStore})
	if err != nil {
		t.Fatalf("StartHttpServer() error = %v", err)
	}
	return fzfHttpServer
}
