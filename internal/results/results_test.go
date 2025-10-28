package results

import (
	"os"
	"testing"

	log "github.com/bonnefoa/kubectl-fzf/v3/internal/logger"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)
	code := m.Run()
	os.Exit(code)
}

func TestParseNamespaceFlag(t *testing.T) {
	r, err := parseNamespaceFlag([]string{"get", "pods", "-ntest"})
	if err != nil {
		t.Fatalf("parseNamespaceFlag() error = %v", err)
	}
	if r == nil || *r != "test" {
		t.Fatalf("parseNamespaceFlag() = %v, want %q", r, "test")
	}

	r, err = parseNamespaceFlag([]string{"get", "pods", "--namespace", "kube-system"})
	if err != nil {
		t.Fatalf("parseNamespaceFlag() error = %v", err)
	}
	if r == nil || *r != "kube-system" {
		t.Fatalf("parseNamespaceFlag() = %v, want %q", r, "kube-system")
	}

	r, err = parseNamespaceFlag([]string{"get", "pods", "--context", "minikube", "--namespace", "kube-system"})
	if err != nil {
		t.Fatalf("parseNamespaceFlag() error = %v", err)
	}
	if r == nil || *r != "kube-system" {
		t.Fatalf("parseNamespaceFlag() = %v, want %q", r, "kube-system")
	}
}

func TestResult(t *testing.T) {
	testDatas := []struct {
		fzfResult        string
		cmdUse           string
		cmdArgs          []string
		currentNamespace string
		expectedResult   string
	}{
		{"kube-system kube-controller-manager-minikube", "get", []string{"pods", " "}, "kube-system", "kube-controller-manager-minikube"},
		{"kube-system coredns-64897985d-nrblm", "get", []string{"pods", "--context", "minikube", "--namespace", "kube-system", ""}, "default", "coredns-64897985d-nrblm"},
		{"kube-system kube-controller-manager-minikube", "get", []string{"pods", " "}, "default", "kube-controller-manager-minikube -n kube-system"},
		{"kube-system kube-controller-manager-minikube", "get", []string{"pods", "-nkube-system", " "}, "default", "kube-controller-manager-minikube"},

		{"kfzf kubectl-fzf-788969b7cb-vf85b", "exec", []string{"-ti", ""}, "default", "kubectl-fzf-788969b7cb-vf85b -n kfzf"},
		// Namespace
		{"default 30d kubernetes.io/metadata.name=default", "get", []string{"pods", "-n="}, "default", "-n=default"},
		{"default 30d kubernetes.io/metadata.name=default", "get", []string{"pods", "-n"}, "default", "-ndefault"},
		{"default 30d kubernetes.io/metadata.name=default", "get", []string{"pods", "-n", " "}, "default", "default"},
		// Label
		{"kube-system tier=control-plane", "get", []string{"pods", "-l="}, "default", "-l=tier=control-plane -n kube-system"},
		{"kube-system tier=control-plane", "get", []string{"pods", "-l", " "}, "default", "tier=control-plane -n kube-system"},
		{"kube-system tier=control-plane", "get", []string{"pods", "-l"}, "default", "-ltier=control-plane -n kube-system"},
		// Namespaceless label
		{"beta.kubernetes.io/arch=amd64 1", "get", []string{"nodes", "-l"}, "default", "-lbeta.kubernetes.io/arch=amd64"},
		// Field selector
		{"kube-system spec.nodeName=minikube", "get", []string{"pods", "--field-selector="}, "default", "--field-selector=spec.nodeName=minikube -n kube-system"},
		{"kube-system spec.nodeName=minikube", "get", []string{"pods", "--field-selector", " "}, "default", "spec.nodeName=minikube -n kube-system"},
		{"kube-system coredns-64897985d-nrblm", "get", []string{"pods", "c"}, "default", "coredns-64897985d-nrblm -n kube-system"},
		{"apiservices.apiregistration.k8s.io None apiregistration.k8s.io/v1", "get", []string{" "}, "default", "apiservices.apiregistration.k8s.io"},
	}
	for _, testData := range testDatas {
		res, err := processResultWithNamespace(testData.cmdUse, testData.cmdArgs, testData.fzfResult, testData.currentNamespace)
		if err != nil {
			t.Fatalf("processResultWithNamespace() error = %v", err)
		}
		if res != testData.expectedResult {
			t.Fatalf("processResultWithNamespace() = %q, want %q (fzf=%q, use=%q, args=%v, ns=%q)", res, testData.expectedResult, testData.fzfResult, testData.cmdUse, testData.cmdArgs, testData.currentNamespace)
		}
	}
}
