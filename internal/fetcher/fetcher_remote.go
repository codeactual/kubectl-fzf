package fetcher

import (
	"fmt"
	"path"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
	log "github.com/bonnefoa/kubectl-fzf/v3/internal/logger"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"
	"github.com/pkg/errors"
)

func (f *Fetcher) loadResourceFromHttpServer(endpoint string, r resources.ResourceType) (map[string]resources.K8sResource, error) {
	resources, err := f.checkHttpCache(endpoint, r)
	if err != nil {
		log.Infof("Error getting resources from cache: %s", err)
	}
	if resources != nil {
		log.Infof("Returning %s resources from cache", r.String())
		return resources, nil
	}
	log.Debugf("Loading from %s", endpoint)
	resourcePath := f.getResourceHttpPath(endpoint, r)
	headers, body, err := util.GetFromHttpServer(resourcePath)
	if err != nil {
		return nil, errors.Wrap(err, "error reading body content")
	}
	err = f.writeResourceToCache(headers, body, r)
	if err != nil {
		return nil, errors.Wrap(err, "error writing fetcher cache")
	}
	util.DecodeGob(&resources, body)
	return resources, err
}

func (f *Fetcher) getResourceHttpPath(host string, r resources.ResourceType) string {
	fullPath := path.Join("k8s", "resources", r.String())
	return fmt.Sprintf("http://%s/%s", host, fullPath)
}
