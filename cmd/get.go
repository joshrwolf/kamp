/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/joshrwolf/kamp/pkg/containers"
	"github.com/joshrwolf/kamp/pkg/registry"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		images := []string{"busybox:latest", "alpine:latest", "nginx:latest"}

		// Build the background context we'll use for the remainder
		ctx := context.Background()

		cwd, _ := os.Getwd()

		// Build and run a registry based on input
		opts := registry.RegistryOpts{
			ImageName: "docker.io/registry:2",
			HostName:  "localhost:5000",
			FileMount: filepath.Join(cwd, "mounted"),
		}

		r := registry.NewDocker(ctx, opts)
		r.Run()

		// Build copy options to be used by all copiers
		copyOpts := containers.CopyOpts{
			Quiet:            true,
			RemoveSignatures: true,

			// TODO: Sign images with GPG key
			SignByFingerprint: "",
		}

		// Build the wait group for containers copying
		var wg sync.WaitGroup
		wg.Add(len(images))

		for _, image := range images {
			go func(image string) error {
				// Create src and dest image containers
				srcContainer := containers.Container{
					Transport: "docker",
					Name:      image,
				}

				destContainer := containers.Container{
					Transport: "oci",
					Name:      image,
					// Registry:  "mounted",
					Backend: "mounted",
				}

				// Copy the image from remote to transient registry
				err := containers.Copy(ctx, srcContainer, destContainer, copyOpts)
				if err != nil {
					log.Fatalf("Failed to copy image: %v", err)
				}

				// Mark the copy as done (copy context is closed)
				wg.Done()
				return nil
			}(image)
		}

		// Wait for all threads to finish
		wg.Wait()

		// Cleanup transient registry
		r.StopAndRemove()
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
