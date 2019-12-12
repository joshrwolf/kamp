package containers

import (
	"context"
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/types"
)

type CopyOpts struct {
	RemoveSignatures  bool
	SignByFingerprint string
	Quiet             bool
}

func Copy(ctx context.Context, srcContainer Container, destContainer Container, opts CopyOpts) error {
	// Build new context with a cancel that inherits from host
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Use insecure policy for now
	// TODO: provide more options
	policy := &signature.Policy{
		Default: []signature.PolicyRequirement{
			signature.NewPRInsecureAcceptAnything(),
		},
	}
	policyCtx, err := signature.NewPolicyContext(policy)
	if err != nil {
		log.Fatalf("Failed to get policy context: %v", err)
	}
	defer policyCtx.Destroy()

	// Get src and dest image refs and context
	srcRef := srcContainer.CreateRefs()
	srcCtx := srcContainer.CreateContext()

	destRef := destContainer.CreateRefs()
	destCtx := destContainer.CreateContext()

	// TODO: unauthenticated is good enough for now b/c all we support is transient registry
	// On a side note, this api is garbage...
	destCtx.DockerInsecureSkipTLSVerify = types.OptionalBoolTrue
	destCtx.DockerDaemonInsecureSkipTLSVerify = true

	var stdout io.Writer

	if !opts.Quiet {
		stdout = os.Stdout
	}

	log.Infof("Copying %s:%s --> %s:%s", srcRef.Transport().Name(), srcRef.StringWithinTransport(), destRef.Transport().Name(), destRef.StringWithinTransport())
	_, err = copy.Image(ctx, policyCtx, destRef, srcRef, &copy.Options{
		RemoveSignatures:      opts.RemoveSignatures,
		SignBy:                opts.SignByFingerprint,
		ReportWriter:          stdout,
		SourceCtx:             srcCtx,
		DestinationCtx:        destCtx,
		ForceManifestMIMEType: manifest.DockerV2Schema2MediaType,
	})

	if err != nil {
		log.Fatalf("Failed to copy %v", err)
	}

	log.Infof("Finished copying %s", destContainer.Name)
	return nil
}
