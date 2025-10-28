package resourcewatcher

import (
	"flag"
	"time"

	"github.com/codeactual/kubectl-fzf/v4/internal/util/config"
)

type ResourceWatcherCli struct {
	watchResources         []string
	excludResources        []string
	watchNamespaces        []string
	excludNamespaces       []string
	ignoreNodeRoles        []string
	nodePollingPeriod      time.Duration
	namespacePollingPeriod time.Duration
	exitOnUnauthorized     bool
}

func SetResourceWatcherCli(fs *flag.FlagSet) {
	fs.Var(config.NewStringSliceValue([]string{}), "watch-resources", "Resources to watch, separated by comma.")
	fs.Var(config.NewStringSliceValue([]string{}), "exclude-resources", "Resources to exclude, separated by comma. To exclude everything: pods,configmaps,services,serviceaccounts,replicasets,daemonsets,secrets,statefulsets,deployments,endpoints,ingresses,cronjobs,jobs,horizontalpodautoscalers,persistentvolumes,persistentvolumeclaims,nodes,namespaces.")
	fs.Var(config.NewStringSliceValue([]string{}), "watch-namespaces", "Namespace regexps to watch, separated by comma.")
	fs.Var(config.NewStringSliceValue([]string{}), "exclude-namespaces", "Namespace regexps to exclude, separated by comma.")
	fs.Var(config.NewStringSliceValue([]string{}), "ignore-node-roles", "List of node role to ommit in the dump. It won't appaear in the completion. Useful to save space and remove cluster for 'common' node role. Separated by comma.")
	fs.Duration("node-polling-period", 300*time.Second, "Polling period for nodes.")
	fs.Duration("namespace-polling-period", 600*time.Second, "Polling period for namespaces.")
	fs.Bool("exit-on-unauthorized", false, "Exit on unauthorized error.")
}

func NewResourceWatcherCli(store *config.Store) ResourceWatcherCli {
	return ResourceWatcherCli{
		watchResources:         store.GetStringSlice("watch-resources", []string{}),
		excludResources:        store.GetStringSlice("exclude-resources", []string{}),
		watchNamespaces:        store.GetStringSlice("watch-namespaces", []string{}),
		excludNamespaces:       store.GetStringSlice("exclude-namespaces", []string{}),
		ignoreNodeRoles:        store.GetStringSlice("ignore-node-roles", []string{}),
		nodePollingPeriod:      store.GetDuration("node-polling-period", 300*time.Second),
		namespacePollingPeriod: store.GetDuration("namespace-polling-period", 600*time.Second),
		exitOnUnauthorized:     store.GetBool("exit-on-unauthorized", false),
	}
}
