package completion

import (
	"context"
	"testing"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/fetcher/fetchertest"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
)

func TestTagLabel(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	labelMap, err := getTagResourceOccurrences(context.Background(), resources.ResourceTypePod, nil, fetchConfig, TagTypeLabel)
	if err != nil {
		t.Fatalf("getTagResourceOccurrences() error = %v", err)
	}
	t.Log(labelMap)

	if _, ok := labelMap[TagResourceKey{"kube-system", "k8s-app=kube-dns"}]; !ok {
		t.Fatalf("expected kube-dns label to be present")
	}
	if count, ok := labelMap[TagResourceKey{"kube-system", "tier=control-plane"}]; !ok {
		t.Fatalf("expected control-plane label to be present")
	} else if count != 4 {
		t.Fatalf("expected control-plane label count to be 4, got %d", count)
	}
}

func TestLabelNamespaceFiltering(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	namespace := "default"
	labelMap, err := getTagResourceOccurrences(context.Background(), resources.ResourceTypePod, &namespace, fetchConfig, TagTypeLabel)
	if err != nil {
		t.Fatalf("getTagResourceOccurrences() error = %v", err)
	}
	if len(labelMap) != 0 {
		t.Fatalf("expected no labels for namespace %q, got %d", namespace, len(labelMap))
	}
}

func TestLabelCompletionPod(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	labelHeader, labelComps, err := GetTagResourceCompletion(context.Background(), resources.ResourceTypePod, nil, fetchConfig, TagTypeLabel)
	if err != nil {
		t.Fatalf("GetTagResourceCompletion() error = %v", err)
	}
	if len(labelComps) != 12 {
		t.Fatalf("expected 12 label completions, got %d", len(labelComps))
	}

	t.Log(labelComps)
	if labelHeader != "Namespace\tLabel\tOccurrences" {
		t.Fatalf("unexpected label header: %s", labelHeader)
	}
	if labelComps[0] != "kube-system\ttier=control-plane\t4" {
		t.Fatalf("unexpected first label completion: %s", labelComps[0])
	}
	if labelComps[1] != "kube-system\taddonmanager.kubernetes.io/mode=Reconcile\t1" {
		t.Fatalf("unexpected second label completion: %s", labelComps[1])
	}
}

func TestLabelCompletionNode(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	labelHeader, labelComps, err := GetTagResourceCompletion(context.Background(), resources.ResourceTypeNode, nil, fetchConfig, TagTypeLabel)
	if err != nil {
		t.Fatalf("GetTagResourceCompletion() error = %v", err)
	}
	if len(labelComps) != 12 {
		t.Fatalf("expected 12 label completions, got %d", len(labelComps))
	}

	t.Log(labelComps)
	if labelHeader != "Label\tOccurrences" {
		t.Fatalf("unexpected label header: %s", labelHeader)
	}
	if labelComps[0] != "beta.kubernetes.io/arch=amd64\t1" {
		t.Fatalf("unexpected first label completion: %s", labelComps[0])
	}
	if labelComps[1] != "beta.kubernetes.io/os=linux\t1" {
		t.Fatalf("unexpected second label completion: %s", labelComps[1])
	}
}

func TestGetFieldSelector(t *testing.T) {
	fetchConfig := fetchertest.GetTestFetcherWithDefaults(t)
	fieldSelectorOccurrences, err := getTagResourceOccurrences(context.Background(), resources.ResourceTypePod, nil, fetchConfig, TagTypeFieldSelector)
	if err != nil {
		t.Fatalf("getTagResourceOccurrences() error = %v", err)
	}

	if count, ok := fieldSelectorOccurrences[TagResourceKey{"kube-system", "spec.nodeName=minikube"}]; !ok {
		t.Fatalf("expected field selector occurrence to be present")
	} else if count != 7 {
		t.Fatalf("expected field selector count to be 7, got %d", count)
	}
}
