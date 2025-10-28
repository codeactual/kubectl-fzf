package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/httpserver"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resourcewatcher"
	storepkg "github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/store"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/kubectlfzfserver"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"
	configstore "github.com/bonnefoa/kubectl-fzf/v3/internal/util/config"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const configFileName = ".kubectl_fzf.json"

var (
	version   = "dev"
	gitCommit = "none"
	gitBranch = "unknown"
	goVersion = "unknown"
	buildDate = "unknown"
)

func versionFun() {
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Git hash: %s\n", gitCommit)
	fmt.Printf("Git branch: %s\n", gitBranch)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Go Version: %s\n", goVersion)
}

func runServer(cfg *configstore.Store) {
	util.CommonInitialization(cfg)
	defer pprof.StopCPUProfile()
	defer util.DoMemoryProfile(cfg)
	kubectlfzfserver.StartKubectlFzfServer(cfg)
}

func main() {
	cfg := configstore.NewStore()
	homeDir, _ := os.UserHomeDir()
	_ = cfg.LoadConfigFile([]string{"/etc/kubectl_fzf", homeDir}, configFileName)
	cfg.ApplyEnv("KUBECTL_FZF")

	rootFlags := flag.NewFlagSet("kubectl_fzf_server", flag.ContinueOnError)
	rootFlags.SetOutput(os.Stdout)
	storepkg.SetStoreConfigCli(rootFlags)
	httpserver.SetHttpServerConfigFlags(rootFlags)
	resourcewatcher.SetResourceWatcherCli(rootFlags)
	util.SetCommonCliFlags(rootFlags, "info")
	if err := cfg.BindFlagSet(rootFlags); err != nil {
		util.FatalIf(err)
	}
	if err := rootFlags.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			return
		}
		util.FatalIf(err)
	}
	cfg.UpdateFromFlagSet(rootFlags)

	args := rootFlags.Args()
	if len(args) > 0 && args[0] == "version" {
		util.CommonInitialization(cfg)
		defer pprof.StopCPUProfile()
		defer util.DoMemoryProfile(cfg)
		versionFun()
		return
	}

	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "unknown subcommand %s\n", args[0])
		os.Exit(1)
	}

	runServer(cfg)
}
