// Package apiready provides a boot-time apiserver-authorization gate. At system
// boot the default ClusterRoleBindings that grant the kubernetes-admin client
// cert (group system:masters) cluster-admin have not yet been reconciled, so
// every authorization check returns forbidden until the bootstrap completes.
// WaitAuthorized rides out that window with a SelfSubjectAccessReview poll.
//
// This is an intentionally-identical copy of svcgw's kube/ready.go helper:
// svcgw (module vibe) and kubectl-fzf (module
// github.com/codeactual/kubectl-fzf/v4) are separate Go modules that cannot
// import a shared package, so the WaitAuthorized body is duplicated verbatim.
package apiready

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff/v5"
	authzv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// WaitAuthorized polls a SelfSubjectAccessReview for attr with exponential
// backoff until the apiserver reports the action is Allowed, returning nil. It
// returns a wrapped error only when timeout elapses first (apiserver
// unreachable or the identity genuinely unauthorized for the full window).
//
// The SSAR is a create against authorization.k8s.io/v1 that answers "can the
// current identity do X?" with HTTP 200 and Status.Allowed; it needs no
// permission on the target resource itself, so it is a side-effect-free probe
// of RBAC-bootstrap completion. A single representative permission suffices
// because the system:masters bindings reconcile together.
func WaitAuthorized(ctx context.Context, core kubernetes.Interface, attr authzv1.ResourceAttributes, timeout time.Duration) error {
	_, err := backoff.Retry(ctx, func() (bool, error) {
		ssar := &authzv1.SelfSubjectAccessReview{
			Spec: authzv1.SelfSubjectAccessReviewSpec{ResourceAttributes: &attr},
		}
		out, err := core.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, ssar, metav1.CreateOptions{})
		if err != nil {
			return false, fmt.Errorf("create SelfSubjectAccessReview: %w", err)
		}
		if !out.Status.Allowed {
			return false, fmt.Errorf("apiserver not yet authorized: %s %s/%s in %q not allowed",
				attr.Verb, attr.Group, attr.Resource, attr.Namespace)
		}
		return true, nil
	}, backoff.WithBackOff(backoff.NewExponentialBackOff()), backoff.WithMaxElapsedTime(timeout),
		// Notify fires only between a failed probe and the next wait — never on a
		// first-try success (ready cluster: no line) and never on the final
		// timeout (the caller surfaces that). So the "waiting for apiserver
		// authorization" line appears exactly when the gate is riding out the
		// boot race.
		backoff.WithNotify(func(err error, d time.Duration) {
			fmt.Fprintf(os.Stderr, "waiting for apiserver authorization (%s %s/%s in %q); retrying in %s: %v\n",
				attr.Verb, attr.Group, attr.Resource, attr.Namespace, d.Round(time.Millisecond), err)
		}))
	if err != nil {
		return fmt.Errorf("wait for apiserver authorization (%s %s/%s in %q): %w",
			attr.Verb, attr.Group, attr.Resource, attr.Namespace, err)
	}
	return nil
}
