package results

import (
	"fmt"
	"strings"

	"github.com/codeactual/kubectl-fzf/v4/internal/fetcher"
	"github.com/codeactual/kubectl-fzf/v4/internal/k8s/resources"
	log "github.com/codeactual/kubectl-fzf/v4/internal/logger"
	"github.com/codeactual/kubectl-fzf/v4/internal/parse"
	"github.com/codeactual/kubectl-fzf/v4/internal/util"
	"github.com/pkg/errors"
)

// ProcessResult handles fzf output and provides completion to use
// The fzfResult should have the first 3 columns of the fzf preview
func ProcessResult(cmdUse string, cmdArgs []string,
	f *fetcher.Fetcher, fzfResult string) (string, error) {
	log.Debugf("Processing fzf result %s", fzfResult)
	log.Debugf("Cmd command %s", cmdArgs)
	namespace, err := f.GetNamespace()
	if err != nil {
		return "", err
	}
	return processResultWithNamespace(cmdUse, cmdArgs, fzfResult, namespace)
}

func completingAfterDoubleDash(cmdArgs []string) bool {
	doubleDashIndex := -1
	for i, arg := range cmdArgs {
		if arg == "--" {
			doubleDashIndex = i
			break
		}
	}
	return doubleDashIndex >= 0 && doubleDashIndex < len(cmdArgs)-1
}

func parseNamespaceFlag(cmdArgs []string) (*string, error) {
	log.Debugf("Parsing namespace from %v", cmdArgs)
	var namespace *string
	for i := 0; i < len(cmdArgs); i++ {
		arg := cmdArgs[i]
		if arg == "--" {
			break
		}
		switch {
		case arg == "-n" || arg == "--namespace":
			if i+1 >= len(cmdArgs) {
				return nil, fmt.Errorf("flag needs an argument: %s", arg)
			}
			value := cmdArgs[i+1]
			namespace = &value
			i++
		case strings.HasPrefix(arg, "-n="):
			value := strings.TrimPrefix(arg, "-n=")
			namespace = &value
		case strings.HasPrefix(arg, "--namespace="):
			value := strings.TrimPrefix(arg, "--namespace=")
			namespace = &value
		case strings.HasPrefix(arg, "-n") && len(arg) > 2:
			value := arg[2:]
			namespace = &value
		}
	}
	return namespace, nil
}

func processResultWithNamespace(cmdUse string, cmdArgs []string, fzfResult string, currentNamespace string) (string, error) {
	// If apiresource:
	// 0 -> fullname, 1 -> shortname, 2 -> groupversion
	// If namespaceless resource:
	// 0 -> name, 1 -> age
	// Otherwise:
	// 0 -> namespace, 1 -> value
	resultFields := strings.Fields(fzfResult)
	if len(resultFields) < 2 {
		return "", fmt.Errorf("fzf result should have at least 3 elements, got %v", resultFields)
	}
	log.Debugf("Processing fzfResult '%s', cmdArgs '%s', current namespace '%s'", fzfResult, cmdArgs, currentNamespace)
	resourceType, flagCompletion, err := parse.ParseFlagAndResources(cmdUse, cmdArgs)
	if err != nil {
		return "", err
	}
	log.Debugf("Resource type %s, flagCompletion %s", resourceType, flagCompletion)

	if resourceType == resources.ResourceTypeApiResource {
		return resultFields[0], nil
	}

	// Generic resource
	resultNamespace := resultFields[0]
	resultValue := resultFields[1]
	if !resourceType.IsNamespaced() {
		resultValue = resultFields[0]
		resultNamespace = ""
	}

	if resourceType == resources.ResourceTypeNamespace {
		resultValue = resultFields[0]
	}

	log.Debugf("Result namespace: %s, resultValue: %s", resultNamespace, resultValue)

	var cmdNamespace *string
	if flagCompletion != parse.FlagNamespace {
		cmdNamespace, err = parseNamespaceFlag(cmdArgs)
		if err != nil {
			return "", errors.Wrapf(err, "Error parsing commands %s", cmdArgs)
		}
		if cmdNamespace != nil {
			log.Debugf("Namespace parsed: %s", *cmdNamespace)
		} else {
			log.Debugf("Namespace parsed: <none>")
		}
	}
	afterDoubleDash := completingAfterDoubleDash(cmdArgs)

	if len(cmdArgs) > 0 {
		lastWord := cmdArgs[len(cmdArgs)-1]
		// add flag to the completion
		lastFlags := []string{"-l=", "-l", "--field-selector=", "--selector=", "-n=", "--namespace=", "-n"}
		if util.IsStringIn(lastWord, lastFlags) {
			resultValue = fmt.Sprintf("%s%s", lastWord, resultValue)
		}
	}

	if afterDoubleDash {
		return resultValue, nil
	}

	if cmdNamespace != nil && *cmdNamespace == resultNamespace {
		return resultValue, nil
	}

	if resultNamespace != "" && resultNamespace != currentNamespace && flagCompletion != parse.FlagNamespace {
		completion := fmt.Sprintf("%s -n %s", resultValue, resultNamespace)
		return completion, nil
	}
	return resultValue, nil
}
