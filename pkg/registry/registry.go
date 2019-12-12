package registry

import (
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
)

type Registry interface {
	Run()
}

type registry struct {
}

type RegistryOpts struct {
	ImageName string
	HostName  string
	FileMount string
	Port      int
}

type docker struct {
	cli         *client.Client
	ctx         context.Context
	containerID string
	RegistryOpts
}

const (
	registryDataPermission = 0755
)

// NewDockerRegistry creates a new registry with a given container runtime
func NewDocker(ctx context.Context, opts RegistryOpts) *docker {
	// Build a cli
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatalf("Failed to create a docker client: %v", err)
	}
	cli.NegotiateAPIVersion(ctx)

	return &docker{
		cli:          cli,
		ctx:          ctx,
		RegistryOpts: opts,
	}
}

func (d *docker) Run() {
	log.Infof("Starting a new local Docker registry using the Docker runtime at: %s", d.HostName)

	reader, err := d.cli.ImagePull(d.ctx, d.ImageName, types.ImagePullOptions{})
	if err != nil {
		log.Fatalf("Failed to pull registry image: %s, with error: %v", d.ImageName, err)
	}

	io.Copy(os.Stdout, reader)

	resp, err := d.cli.ContainerCreate(
		d.ctx,
		&container.Config{
			Image: d.ImageName,
			ExposedPorts: nat.PortSet{
				"5000/tcp": struct{}{},
			},
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				"5000/tcp": []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: "5000",
					},
				},
			},
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: d.FileMount,
					Target: "/var/lib/registry",
				},
			},
		}, nil, "")

	if err != nil {
		log.Fatalf("Failed to run registry with docker runtime: %v", err)
	}

	// Capture the running container ID
	d.containerID = resp.ID

	if err := d.cli.ContainerStart(d.ctx, d.containerID, types.ContainerStartOptions{}); err != nil {
		log.Fatalf("Registry failed to start: %v", err)
	}

	log.Infof("Registry is successfully running with ID: %s at %s mounted to '%s'", resp.ID, d.HostName, d.FileMount)

	// Capture sigterm for impatient users
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	statusCh, errCh := d.cli.ContainerWait(d.ctx, d.containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			log.Fatalf("Failed to do something? %v", err)
		}
	case <-statusCh:
		d.StopAndRemove()
		log.Infof("Registry terminated")
	case <-c:
		d.StopAndRemove()
		log.Infof("SIGTERM: Transient registry stopped and removed")
	default:
	}

	out, err := d.cli.ContainerLogs(d.ctx, d.containerID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		log.Fatalf("Failed something? %v", err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}

// StopAndRemove will stop the running container and remove it
func (d *docker) StopAndRemove() error {
	if err := d.cli.ContainerStop(d.ctx, d.containerID, nil); err != nil {
		return errors.New("Failed to stop container")
	}

	err := d.cli.ContainerRemove(d.ctx, d.containerID, types.ContainerRemoveOptions{})
	if err != nil {
		return errors.New("Failed to remove container")
	}

	log.Infof("Stopped and removed container: %s", d.containerID)
	return nil
}
