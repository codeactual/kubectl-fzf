package httpserver

import (
	"flag"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/util/config"
)

type HttpServerConfigCli struct {
	ListenAddress   string
	HttpProfAddress string
	Debug           bool
}

func SetHttpServerConfigFlags(fs *flag.FlagSet) {
	fs.String("listen-address", "localhost:8080", "Listen address of the http server")
	fs.String("http-prof-address", "localhost:6060", "Listen address of the pprof endpoint")
	fs.Bool("http-debug", false, "Activate debug mode of the http server")
}

func NewHttpServerConfigCli(store *config.Store) HttpServerConfigCli {
	return HttpServerConfigCli{
		ListenAddress:   store.GetString("listen-address", "localhost:8080"),
		HttpProfAddress: store.GetString("http-prof-address", "localhost:6060"),
		Debug:           store.GetBool("http-debug", false),
	}
}
