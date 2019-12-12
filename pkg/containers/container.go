package containers

import (
	"fmt"
	log "github.com/sirupsen/logrus"

	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	"github.com/joshrwolf/kamp/utils/k8s"
)

type Container struct {
	Image     k8s.Image
	Name      string
	Transport string
	Registry  string
	Backend   string
	TLSVerify bool
}

// CreateContext docs TODO
func (c *Container) CreateContext() *types.SystemContext {
	ctx := &types.SystemContext{
		OSChoice: "linux",
	}
	return ctx
}

// CreateRefs docs TODO
func (c *Container) CreateRefs() types.ImageReference {
	fullTransport := createFullTransport(c.Transport, c.Backend, c.Name)

	ref, err := alltransports.ParseImageName(fullTransport)
	if err != nil {
		log.Fatalf("Failed to parse image name: %s, %v", fullTransport, err)
	}

	return ref
}

func createFullTransport(transport string, backend string, imageName string) (fullTransport string) {
	// Build full transport uri depending on transport type and backend type
	if transport == "docker" {
		if backend == "" {
			// Use the default "docker.io" registry as the backend
			backend = "docker.io"
		}
		fullTransport = fmt.Sprintf("%s://%s/%s", transport, backend, imageName)
	} else {
		// Use some sort of file based storage transport
		fullTransport = fmt.Sprintf("%s:%s/%s", transport, backend, imageName)
	}

	return
}
