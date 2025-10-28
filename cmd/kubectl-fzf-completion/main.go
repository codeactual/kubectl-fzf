package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/completion"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/fetcher"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/fzf"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/gencode"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/clusterconfig"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
	storepkg "github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/store"
	log "github.com/bonnefoa/kubectl-fzf/v3/internal/logger"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/parse"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/results"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"
	configstore "github.com/bonnefoa/kubectl-fzf/v3/internal/util/config"
)

const (
	FallbackExitCode = 6
	configFileName   = ".kubectl_fzf.json"
)

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

func completeFun(store *configstore.Store, cmdArgs []string) {
	args := completion.PrepareCmdArgs(cmdArgs)
	if args == nil {
		os.Exit(FallbackExitCode)
	}

	firstWord := args[0]
	verbs := []string{"get", "exec", "logs", "label", "describe", "delete", "annotate", "edit", "scale"}
	if !util.IsStringIn(firstWord, verbs) {
		os.Exit(FallbackExitCode)
	}
	args = args[1:]

	fetchConfigCli := fetcher.NewFetcherCli(store)
	f := fetcher.NewFetcher(&fetchConfigCli)
	err := f.LoadFetcherState()
	if err != nil {
		log.Warnf("Error loading fetcher state")
		os.Exit(FallbackExitCode)
	}

	completionResults, err := completion.ProcessCommandArgs(firstWord, args, f)
	if e, ok := err.(resources.UnknownResourceError); ok {
		log.Warnf("Unknown resource type: %s", e)
		os.Exit(FallbackExitCode)
	} else if e, ok := err.(parse.UnmanagedFlagError); ok {
		log.Warnf("Unmanaged flag: %s", e)
		os.Exit(FallbackExitCode)
	} else if err != nil {
		log.Warnf("Error during completion: %s", err)
		os.Exit(FallbackExitCode)
	}

	err = f.SaveFetcherState()
	if err != nil {
		log.Warnf("Error saving fetcher state: %s", err)
		os.Exit(FallbackExitCode)
	}

	if len(completionResults.Completions) == 0 {
		log.Warn("No completion found")
		os.Exit(5)
	}
	formattedComps := completionResults.GetFormattedOutput()

	query := completion.ExtractQueryFromArgs(args)
	fzfResult, err := fzf.CallFzf(formattedComps, query)
	if err != nil {
		if e, ok := err.(fzf.InterruptedCommandError); ok {
			log.Infof("Fzf was interrupted: %s", e)
			os.Exit(FallbackExitCode)
		}
		log.Fatalf("Call fzf error: %s", err)
	}
	res, err := results.ProcessResult(firstWord, args, f, fzfResult)
	if err != nil {
		log.Fatalf("Process result error: %s", err)
	}
	fmt.Print(res)
}

func statsFun(cfg *configstore.Store) {
	fetchConfigCli := fetcher.NewFetcherCli(cfg)
	f := fetcher.NewFetcher(&fetchConfigCli)
	ctx := context.Background()
	stats, err := f.GetStats(ctx)
	util.FatalIf(err)
	statsOutput := storepkg.GetStatsOutput(stats)
	fmt.Print(statsOutput)
}

func genFun(cfg *configstore.Store) {
	ctx := context.Background()
	err := gencode.GenerateResourceCode(ctx)
	util.FatalIf(err)
}

func runCompletionCommand(cfg *configstore.Store, args []string) {
	util.CommonInitialization(cfg)
	defer pprof.StopCPUProfile()
	defer util.DoMemoryProfile(cfg)
	completeFun(cfg, args)
}

func runStatsCommand(cfg *configstore.Store, args []string) {
	statsFlags := flag.NewFlagSet("stats", flag.ContinueOnError)
	statsFlags.SetOutput(os.Stdout)
	fetcher.SetFetchConfigFlags(statsFlags)
	if err := cfg.BindFlagSet(statsFlags); err != nil {
		util.FatalIf(err)
	}
	if err := statsFlags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return
		}
		util.FatalIf(err)
	}
	cfg.UpdateFromFlagSet(statsFlags)
	util.CommonInitialization(cfg)
	defer pprof.StopCPUProfile()
	defer util.DoMemoryProfile(cfg)
	statsFun(cfg)
}

func runGenerateCommand(cfg *configstore.Store, args []string) {
	genFlags := flag.NewFlagSet("generate", flag.ContinueOnError)
	genFlags.SetOutput(os.Stdout)
	clusterconfig.SetClusterConfigCli(genFlags)
	if err := cfg.BindFlagSet(genFlags); err != nil {
		util.FatalIf(err)
	}
	if err := genFlags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return
		}
		util.FatalIf(err)
	}
	cfg.UpdateFromFlagSet(genFlags)
	util.CommonInitialization(cfg)
	defer pprof.StopCPUProfile()
	defer util.DoMemoryProfile(cfg)
	genFun(cfg)
}

func main() {
	cfg := configstore.NewStore()
	homeDir, _ := os.UserHomeDir()
	_ = cfg.LoadConfigFile([]string{"/etc/kubectl_fzf", homeDir}, configFileName)
	cfg.ApplyEnv("KUBECTL_FZF")

	rootFlags := flag.NewFlagSet("kubectl_fzf_completion", flag.ContinueOnError)
	rootFlags.SetOutput(os.Stdout)
	util.SetCommonCliFlags(rootFlags, "error")
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
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "expected subcommand: k8s_completion, stats, generate, version")
		os.Exit(1)
	}

	switch args[0] {
	case "version":
		util.CommonInitialization(cfg)
		defer pprof.StopCPUProfile()
		defer util.DoMemoryProfile(cfg)
		versionFun()
	case "k8s_completion":
		runCompletionCommand(cfg, args[1:])
	case "stats":
		runStatsCommand(cfg, args[1:])
	case "generate":
		runGenerateCommand(cfg, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %s\n", args[0])
		os.Exit(1)
	}
}
