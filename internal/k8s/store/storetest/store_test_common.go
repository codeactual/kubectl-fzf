package storetest

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/clusterconfig"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/store"
	log "github.com/bonnefoa/kubectl-fzf/v3/internal/logger"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)
	code := m.Run()
	os.Exit(code)
}

func podResource(name string, ns string, labels map[string]string) corev1.Pod {
	meta := corev1.Pod{
		TypeMeta: metav1.TypeMeta{Kind: "Pod"},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         ns,
			Labels:            labels,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec:   corev1.PodSpec{},
		Status: corev1.PodStatus{},
	}
	return meta
}

func GetTestPodStore(t *testing.T) (string, *store.Store) {
	tempDir, err := ioutil.TempDir("/tmp/", "cacheTest")
	if err != nil {
		t.Fatalf("TempDir() error = %v", err)
	}
	storeConfigCli := &store.StoreConfigCli{
		ClusterConfigCli: &clusterconfig.ClusterConfigCli{
			ClusterName: "test", CacheDir: tempDir},
		TimeBetweenFullDump: 500 * time.Millisecond}
	storeConfig := store.NewStoreConfig(storeConfigCli)
	err = storeConfig.CreateDestDir()
	if err != nil {
		t.Fatalf("CreateDestDir() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctorConfig := resources.CtorConfig{}
	k8sStore := store.NewStore(ctx, storeConfig, ctorConfig, resources.ResourceTypePod)
	pods := []corev1.Pod{
		podResource("Test1", "ns1", map[string]string{"app": "app1"}),
		podResource("Test2", "ns2", map[string]string{"app": "app2"}),
		podResource("Test3", "ns2", map[string]string{"app": "app2"}),
		podResource("Test4", "aaa", map[string]string{"app": "app3"}),
	}
	for _, pod := range pods {
		k8sStore.AddResource(&pod)
	}
	return tempDir, k8sStore
}

func TestDumpPodFullState(t *testing.T) {
	tempDir, k := GetTestPodStore(t)
	defer util.RemoveTempDir(tempDir)

	err := k.DumpFullState()
	if err != nil {
		t.Fatalf("DumpFullState() error = %v", err)
	}
	podFilePath := path.Join(tempDir, "test", "pods")
	if _, statErr := os.Stat(podFilePath); statErr != nil {
		t.Fatalf("expected pod dump file to exist: %v", statErr)
	}

	pods := map[string]resources.K8sResource{}
	err = util.LoadGobFromFile(&pods, podFilePath)
	if err != nil {
		t.Fatalf("LoadGobFromFile() error = %v", err)
	}

	if len(pods) != 4 {
		t.Fatalf("expected 4 pods, got %d", len(pods))
	}
	for _, key := range []string{"ns1_Test1", "ns2_Test2", "ns2_Test3", "aaa_Test4"} {
		if _, ok := pods[key]; !ok {
			t.Fatalf("expected pod key %q in dump", key)
		}
	}
}

func TestTickerPodDumpFullState(t *testing.T) {
	tempDir, s := GetTestPodStore(t)
	defer util.RemoveTempDir(tempDir)

	time.Sleep(1000 * time.Millisecond)
	podFilePath := path.Join(tempDir, "test", "pods")
	if _, err := os.Stat(podFilePath); err != nil {
		t.Fatalf("expected pod dump file to exist: %v", err)
	}
	pods := map[string]resources.K8sResource{}
	err := util.LoadGobFromFile(&pods, podFilePath)
	if err != nil {
		t.Fatalf("LoadGobFromFile() error = %v", err)
	}
	if len(pods) != 4 {
		t.Fatalf("expected 4 pods, got %d", len(pods))
	}

	pod := podResource("Test1", "ns1", map[string]string{"app": "app1"})
	s.AddResource(&pod)
	fileInfoBefore, err := os.Stat(podFilePath)
	if err != nil {
		t.Fatalf("os.Stat() before error = %v", err)
	}
	time.Sleep(1000 * time.Millisecond)
	fileInfoAfter, err := os.Stat(podFilePath)
	if err != nil {
		t.Fatalf("os.Stat() after error = %v", err)
	}
	if fileInfoBefore.ModTime().Before(fileInfoAfter.ModTime()) {
		t.Fatalf("expected file modification time to be non-increasing")
	}
}
