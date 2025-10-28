package fzf

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	log "github.com/codeactual/kubectl-fzf/v4/internal/logger"
)

type InterruptedCommandError string

func (u InterruptedCommandError) Error() string {
	return string(u)
}

func setCompsInStdin(cmd *exec.Cmd, comps string) error {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		_, err = io.Copy(stdin, strings.NewReader(comps))
		if err != nil {
			log.Error(err, "error copy stdin")
		}
		err = stdin.Close()
		if err != nil {
			log.Error(err, "error closing stdin")
		}
	}()
	return nil
}

func CallFzf(comps string, query string) (string, error) {
	var result strings.Builder
	header := strings.Split(comps, "\n")[1]
	// Leave an additional line for overflow
	numFields := len(strings.Fields(header)) + 1
	log.Debugf("header: %s, numFields: %d", header, numFields)
	previewWindow := fmt.Sprintf("--preview-window=down:%d", numFields)
	previewCmd := fmt.Sprintf("echo -e \"%s\n{}\" | sed -e \"s/'//g\" | awk '(NR==1){for (i=1; i<=NF; i++) a[i]=$i} (NR==2){for (i in a) {printf a[i] \": \" $i \"\\n\"} }' | column -t | fold -w $COLUMNS", header)

	// TODO Make fzf options configurable
	fzfArgs := []string{"-1", "--header-lines=2", "--layout", "reverse", "-e", "--no-hscroll", "--no-sort", "--cycle", "-q", query, previewWindow, "--preview", previewCmd}
	log.Infof("fzf args: %+v", fzfArgs)
	cmd := exec.Command("fzf", fzfArgs...)
	cmd.Stdout = &result
	cmd.Stderr = os.Stderr

	err := setCompsInStdin(cmd, comps)
	if err != nil {
		return "", err
	}

	log.Info("Start fzf command")
	err = cmd.Start()
	if err != nil {
		log.Fatalf("Error when running fzf: %s", err)
	}

	err = cmd.Wait()
	if err != nil {
		if cmd.ProcessState.ExitCode() == 130 {
			// Interrupted with C-c or ESC
			return "", InterruptedCommandError(err.Error())
		}
		return "", err
	}
	res := strings.TrimSpace(result.String())
	log.Infof("Fzf result: %s", res)
	return res, nil
}
