package apiready

import (
	"context"
	"fmt"
	"testing"
	"time"

	authzv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/runtime"
	corefake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

// ssarReactor returns a "create selfsubjectaccessreviews" reactor that replies
// with the given Allowed value.
func ssarReactor(allowed bool) k8stesting.ReactionFunc {
	return func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, &authzv1.SelfSubjectAccessReview{
			Status: authzv1.SubjectAccessReviewStatus{Allowed: allowed},
		}, nil
	}
}

// testAttr is kubectl-fzf's representative permission: list namespaces at
// cluster scope, the first forbidden op observed during the boot race.
var testAttr = authzv1.ResourceAttributes{
	Verb:     "list",
	Group:    "",
	Resource: "namespaces",
}

func TestWaitAuthorizedAllowed(t *testing.T) {
	cs := corefake.NewClientset()
	cs.PrependReactor("create", "selfsubjectaccessreviews", ssarReactor(true))

	if err := WaitAuthorized(context.Background(), cs, testAttr, time.Second); err != nil {
		t.Fatalf("WaitAuthorized with Allowed=true: %v", err)
	}
}

func TestWaitAuthorizedNotAllowed(t *testing.T) {
	cs := corefake.NewClientset()
	cs.PrependReactor("create", "selfsubjectaccessreviews", ssarReactor(false))

	// A sub-second cap means the first computed backoff already exceeds the
	// budget, so WaitAuthorized gives up after one failed probe instead of
	// blocking the full 3-minute timeout.
	err := WaitAuthorized(context.Background(), cs, testAttr, 50*time.Millisecond)
	if err == nil {
		t.Fatal("WaitAuthorized with Allowed=false: got nil error, want timeout error")
	}
}

func TestWaitAuthorizedRetriesAfterTransientError(t *testing.T) {
	cs := corefake.NewClientset()
	calls := 0
	cs.PrependReactor("create", "selfsubjectaccessreviews", func(k8stesting.Action) (bool, runtime.Object, error) {
		calls++
		if calls == 1 {
			return true, nil, fmt.Errorf("transient apiserver error")
		}
		return true, &authzv1.SelfSubjectAccessReview{
			Status: authzv1.SubjectAccessReviewStatus{Allowed: true},
		}, nil
	})

	if err := WaitAuthorized(context.Background(), cs, testAttr, time.Minute); err != nil {
		t.Fatalf("WaitAuthorized after transient error: %v", err)
	}
	if calls < 2 {
		t.Errorf("SSAR create called %d times, want >= 2 (one transient failure then success)", calls)
	}
}
