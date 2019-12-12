package registry_test

import (
	"context"
	"github.com/joshrwolf/kamp/pkg/registry"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunDockerRegistry(t *testing.T) {
	ctx := context.Background()

	cwd, _ := os.Getwd()

	opts := registry.RegistryOpts{
		ImageName: "docker.io/registry:2",
		HostName:  "localhost:5000",
		FileMount: filepath.Join(cwd, "mounted"),
	}

	r := registry.NewDocker(ctx, opts)
	r.Run()

	time.Sleep(10 * time.Second)

	r.StopAndRemove()
}
