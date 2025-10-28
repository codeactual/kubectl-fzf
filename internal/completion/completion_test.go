package completion

import (
	"context"
	"errors"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/fetcher/fetchertest"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/httpserver/httpservertest"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
	log "github.com/bonnefoa/kubectl-fzf/v3/internal/logger"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/parse"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)
	code := m.Run()
	os.Exit(code)
}

type cmdArg struct {
	verb string
	args []string
}

func TestPrepareCmdArgs(t *testing.T) {
	testDatas := []struct {
		cmdArgs        []string
		expectedResult []string
	}{
		{[]string{"get pods"}, []string{"get", "pods"}},
		{[]string{"get pods "}, []string{"get", "pods", " "}},
	}
	for _, testData := range testDatas {
		cmdArgs := PrepareCmdArgs(testData.cmdArgs)
		if !reflect.DeepEqual(testData.expectedResult, cmdArgs) {
			t.Errorf("PrepareCmdArgs(%v) = %v, want %v", testData.cmdArgs, cmdArgs, testData.expectedResult)
		}
	}

}

func TestProcessResourceName(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	cmdArgs := []cmdArg{
		{"get", []string{"pods", ""}},
		{"get", []string{"po", ""}},
		{"logs", []string{""}},
		{"exec", []string{"-ti", ""}},
	}
	for _, cmdArg := range cmdArgs {
		completionResults, err := processCommandArgsWithFetchConfig(context.Background(), fetchConfig, cmdArg.verb, cmdArg.args)
		if err != nil {
			t.Fatalf("processCommandArgsWithFetchConfig() error = %v", err)
		}
		if len(completionResults.Completions) == 0 {
			t.Fatalf("expected completions for %v", cmdArg)
		}
		if !strings.Contains(completionResults.Completions[0], "kube-system\tcoredns-6d4b75cb6d-m6m4q\t172.17.0.3\t192.168.49.2\tminikube\tRunning\tBurstable\tcoredns\tCriticalAddonsOnly:,node-role.kubernetes.io/master:NoSchedule,node-role.kubernetes.io/control-plane:NoSchedule\tNone") {
			t.Fatalf("unexpected completion content: %s", completionResults.Completions[0])
		}
	}
}

func TestProcessNamespace(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	cmdArgs := []cmdArg{
		{"get", []string{"pods", "-n"}},
		{"get", []string{"pods", "-n", " "}},
		{"get", []string{"po", "-n="}},
		{"logs", []string{"--namespace", ""}},
		{"logs", []string{"--namespace="}},
	}
	for _, cmdArg := range cmdArgs {
		completionResults, err := processCommandArgsWithFetchConfig(context.Background(), fetchConfig, cmdArg.verb, cmdArg.args)
		if err != nil {
			t.Fatalf("processCommandArgsWithFetchConfig() error = %v", err)
		}
		if len(completionResults.Completions) == 0 {
			t.Fatalf("expected completions for %v", cmdArg)
		}
		if !strings.Contains(completionResults.Completions[0], "default\t") {
			t.Fatalf("expected namespace completion to contain default: %s", completionResults.Completions[0])
		}
	}
}

func TestProcessLabelCompletion(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	cmdArgs := []cmdArg{
		{"get", []string{"pods", "-l="}},
		{"get", []string{"pods", "-l"}},
		{"get", []string{"pods", "-l", ""}},
		{"get", []string{"pods", "--selector", ""}},
		{"get", []string{"pods", "--selector="}},
	}
	for _, cmdArg := range cmdArgs {
		completionResults, err := processCommandArgsWithFetchConfig(context.Background(), fetchConfig, cmdArg.verb, cmdArg.args)
		if err != nil {
			t.Fatalf("processCommandArgsWithFetchConfig() error = %v", err)
		}
		if len(completionResults.Completions) == 0 {
			t.Fatalf("expected label completions for %v", cmdArg)
		}
		if completionResults.Completions[0] != "kube-system\ttier=control-plane\t4" {
			t.Fatalf("unexpected label completion: %s", completionResults.Completions[0])
		}
		if len(completionResults.Completions) != 12 {
			t.Fatalf("expected 12 label completions, got %d", len(completionResults.Completions))
		}
	}
}

func TestProcessFieldSelectorCompletion(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	cmdArgs := []cmdArg{
		{"get", []string{"pods", "--field-selector", ""}},
		{"get", []string{"pods", "--field-selector="}},
	}
	for _, cmdArg := range cmdArgs {
		completionResults, err := processCommandArgsWithFetchConfig(context.Background(), fetchConfig, cmdArg.verb, cmdArg.args)
		if err != nil {
			t.Fatalf("processCommandArgsWithFetchConfig() error = %v", err)
		}
		if len(completionResults.Completions) == 0 {
			t.Fatalf("expected field selector completions for %v", cmdArg)
		}
		if completionResults.Completions[0] != "kube-system\tspec.nodeName=minikube\t7" {
			t.Fatalf("unexpected field selector completion: %s", completionResults.Completions[0])
		}
	}
}

func TestUnmanagedCompletion(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	cmdArgs := []cmdArg{
		{"get", []string{"-t"}},
		{"get", []string{"-i"}},
		{"get", []string{"--field-selector"}},
		{"get", []string{"--selector"}},
		{"get", []string{"--all-namespaces"}},
		{"get", []string{"pods", "aPod", ">", "/tmp"}},
	}
	for _, cmdArg := range cmdArgs {
		_, err := processCommandArgsWithFetchConfig(context.Background(), fetchConfig, cmdArg.verb, cmdArg.args)
		if err == nil {
			t.Fatalf("expected unmanaged error for cmdArgs %v", cmdArg)
		}
		var unmanagedErr parse.UnmanagedFlagError
		if !errors.As(err, &unmanagedErr) {
			t.Fatalf("expected UnmanagedFlagError, got %T", err)
		}
	}
}

func TestManagedCompletion(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	cmdArgs := []cmdArg{
		{"get", []string{"pods", "--selector", ""}},
		{"get", []string{"pods", "--selector="}},
		{"get", []string{"pods", "--field-selector", ""}},
		{"get", []string{"pods", "--field-selector="}},
		{"get", []string{"pods", "-t", ""}},
		{"get", []string{"pods", "-i", ""}},
		{"get", []string{"pods", "-ti", ""}},
		{"get", []string{"pods", "-it", ""}},
		{"get", []string{"-n"}},
		{"get", []string{"-n", ""}},
		{"get", []string{"pods", "--all-namespaces", ""}},
	}
	for _, cmdArg := range cmdArgs {
		completionResults, err := processCommandArgsWithFetchConfig(context.Background(), fetchConfig, cmdArg.verb, cmdArg.args)
		if err != nil {
			t.Fatalf("processCommandArgsWithFetchConfig() error = %v", err)
		}
		if completionResults == nil {
			t.Fatalf("expected completion results for %v", cmdArg)
		}
	}
}

func TestPodCompletionFile(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	res, err := getResourceCompletion(context.Background(), resources.ResourceTypePod, nil, fetchConfig)
	if err != nil {
		t.Fatalf("getResourceCompletion() error = %v", err)
	}
	t.Log(res)
	if len(res) == 0 {
		t.Fatalf("expected pod completions")
	}
	if !strings.Contains(res[0], "kube-system\t") {
		t.Fatalf("expected first completion to contain kube-system: %s", res[0])
	}
	if len(res) != 7 {
		t.Fatalf("expected 7 pod completions, got %d", len(res))
	}
}

func TestNamespaceFilterFile(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)

	// everything is filtered
	namespace := "test"
	res, err := getResourceCompletion(context.Background(), resources.ResourceTypePod, &namespace, fetchConfig)
	if err != nil {
		t.Fatalf("getResourceCompletion() error = %v", err)
	}
	t.Log(res)
	if len(res) != 0 {
		t.Fatalf("expected no completions for namespace %q, got %d", namespace, len(res))
	}

	// all results match
	namespace = "kube-system"
	res, err = getResourceCompletion(context.Background(), resources.ResourceTypePod, &namespace, fetchConfig)
	if err != nil {
		t.Fatalf("getResourceCompletion() error = %v", err)
	}
	if len(res) != 7 {
		t.Fatalf("expected 7 completions for namespace %q, got %d", namespace, len(res))
	}
}

func TestApiResourcesFile(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	res, err := getResourceCompletion(context.Background(), resources.ResourceTypeApiResource, nil, fetchConfig)
	if err != nil {
		t.Fatalf("getResourceCompletion() error = %v", err)
	}
	sort.Strings(res)
	if len(res) == 0 {
		t.Fatalf("expected api resource completions")
	}
	if !strings.Contains(res[0], "apiservices\tNone\tapiregistration.k8s.io/v1\tfalse\tAPIService") {
		t.Fatalf("unexpected api resource completion: %s", res[0])
	}
}

func TestHttpServerApiCompletion(t *testing.T) {
	fzfHttpServer := httpservertest.StartTestHttpServer(t)
	f, tempDir := fetchertest.GetTestFetcher(t, "nothing", fzfHttpServer.Port)
	res, err := getResourceCompletion(context.Background(), resources.ResourceTypeApiResource, nil, f)
	if err != nil {
		t.Fatalf("getResourceCompletion() error = %v", err)
	}
	sort.Strings(res)
	if len(res) == 0 {
		t.Fatalf("expected api resource completions")
	}
	if !strings.Contains(res[0], "apiservices\tNone\tapiregistration.k8s.io/v1\tfalse\tAPIService") {
		t.Fatalf("unexpected api resource completion: %s", res[0])
	}
	if len(res) != 56 {
		t.Fatalf("expected 56 api resource completions, got %d", len(res))
	}

	expectedPath := path.Join(tempDir, "nothing", resources.ResourceTypeApiResource.String())
	if _, statErr := os.Stat(expectedPath); statErr != nil {
		t.Fatalf("expected cache file to exist: %v", statErr)
	}
}

func TestHttpServerPodCompletion(t *testing.T) {
	fzfHttpServer := httpservertest.StartTestHttpServer(t)
	f, tempDir := fetchertest.GetTestFetcher(t, "nothing", fzfHttpServer.Port)
	res, err := getResourceCompletion(context.Background(), resources.ResourceTypePod, nil, f)
	if err != nil {
		t.Fatalf("getResourceCompletion() error = %v", err)
	}
	if len(res) == 0 {
		t.Fatalf("expected pod completions")
	}
	if !strings.Contains(res[0], "kube-system\t") {
		t.Fatalf("expected first completion to contain kube-system: %s", res[0])
	}
	if len(res) != 7 {
		t.Fatalf("expected 7 pod completions, got %d", len(res))
	}

	expectedPath := path.Join(tempDir, "nothing", resources.ResourceTypePod.String())
	if _, statErr := os.Stat(expectedPath); statErr != nil {
		t.Fatalf("expected cache file to exist: %v", statErr)
	}
}

func TestHttpUnknownResourceCompletion(t *testing.T) {
	fzfHttpServer := httpservertest.StartTestHttpServer(t)
	f, tempDir := fetchertest.GetTestFetcher(t, "nothing", fzfHttpServer.Port)
	_, err := getResourceCompletion(context.Background(), resources.ResourceTypePersistentVolume, nil, f)
	if err == nil {
		t.Fatalf("expected error for unknown resource type")
	}

	expectedPath := path.Join(tempDir, "nothing")
	if _, statErr := os.Stat(expectedPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected no cache directory, got err=%v", statErr)
	}
}

func TestHttpServerCachePod(t *testing.T) {
	fzfHttpServer := httpservertest.StartTestHttpServer(t)
	f, tempDir := fetchertest.GetTestFetcher(t, "nothing", fzfHttpServer.Port)
	res, err := getResourceCompletion(context.Background(), resources.ResourceTypePod, nil, f)
	if err != nil {
		t.Fatalf("getResourceCompletion() error = %v", err)
	}
	err = f.SaveFetcherState()
	if err != nil {
		t.Fatalf("SaveFetcherState() error = %v", err)
	}
	if len(res) != 7 {
		t.Fatalf("expected 7 pod completions, got %d", len(res))
	}

	podCache := path.Join(tempDir, "nothing", resources.ResourceTypePod.String())
	if _, statErr := os.Stat(podCache); statErr != nil {
		t.Fatalf("expected pod cache file to exist: %v", statErr)
	}
	if fzfHttpServer.ResourceHit != 1 {
		t.Fatalf("expected ResourceHit to be 1, got %d", fzfHttpServer.ResourceHit)
	}
	fetcher_state := path.Join(tempDir, "fetcher_state")
	if _, statErr := os.Stat(fetcher_state); statErr != nil {
		t.Fatalf("expected fetcher state file to exist: %v", statErr)
	}

	res, err = getResourceCompletion(context.Background(), resources.ResourceTypePod, nil, f)
	if err != nil {
		t.Fatalf("getResourceCompletion() error = %v", err)
	}
	_ = res
	if fzfHttpServer.ResourceHit != 1 {
		t.Fatalf("expected ResourceHit to remain 1, got %d", fzfHttpServer.ResourceHit)
	}
}
