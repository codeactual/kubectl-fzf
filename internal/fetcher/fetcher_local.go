package fetcher

import (
	"os"
	"time"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
	log "github.com/bonnefoa/kubectl-fzf/v3/internal/logger"
)

func (f *Fetcher) checkLocalFiles(r resources.ResourceType) (map[string]resources.K8sResource, error) {
	resourceStorePath := f.GetResourceStorePath(r)
	finfo, err := os.Stat(resourceStorePath)
	if err != nil {
		return nil, nil
	}

	deltaMod := time.Now().Sub(finfo.ModTime())
	log.Infof("%s found, using resources from file", resourceStorePath)
	if deltaMod >= time.Hour {
		log.Warnf("%s was not modified for more than one hour", resourceStorePath)
	}
	resources, err := loadResourceFromFile(resourceStorePath)
	return resources, err
}
