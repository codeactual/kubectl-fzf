package util

import (
	"fmt"
	"net"
	"runtime/debug"
	"time"

	log "github.com/bonnefoa/kubectl-fzf/v3/internal/logger"
	"github.com/pkg/errors"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func IsAddressReachable(address string) bool {
	if address == "" {
		return false
	}
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		log.Infof("Couldn't connect to %s: %s", address, err)
		return false
	}
	conn.Close()
	return true
}

// FatalIf exits if the error is not nil
func FatalIf(err error) {
	if err != nil {
		if stackErr, ok := err.(stackTracer); ok {
			log.Errorf("stacktrace: %+v", stackErr.StackTrace())
		} else {
			debug.PrintStack()
		}
		log.Fatalf("Fatal error: %s", err)
	}
}

// TimeToAge converts a time to a age string
func TimeToAge(t time.Time) string {
	duration := time.Since(t)
	duration = duration.Round(time.Minute)
	if duration.Hours() > 30 {
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	}
	hour := duration / time.Hour
	duration -= hour * time.Hour
	minute := duration / time.Minute
	return fmt.Sprintf("%02d:%02d", hour, minute)
}
