package k8s

import (
	"bytes"
	log "github.com/sirupsen/logrus"

	kust "k8s.io/cli-runtime/pkg/kustomize"
	"sigs.k8s.io/kustomize/pkg/fs"
)

// Kustomize docs TODO
type Kustomize interface {
	Build() []Image
}

type kustomize struct {
	path string
	repo string
}

func (k *kustomize) Build() string {
	// Build a buffer to stream `kustomize build` to
	var buildOutput bytes.Buffer

	// Build using RunKustomizeBuild
	err := kust.RunKustomizeBuild(&buildOutput, fs.MakeFakeFS(), k.path)
	if err != nil {
		log.Fatalf("Failed to kustomize build on: %s, %v", k.path, err)
	}

	return buildOutput.String()
}
