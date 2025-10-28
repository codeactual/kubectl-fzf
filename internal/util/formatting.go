package util

import (
	"fmt"
	"strings"
	"text/tabwriter"

	log "github.com/bonnefoa/kubectl-fzf/v3/internal/logger"
)

func FormatCompletion(lines []string) string {
	log.Info("formating completion")
	b := new(strings.Builder)
	w := tabwriter.NewWriter(b, 0, 0, 1, ' ', tabwriter.StripEscape)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	w.Flush()
	return b.String()
}
