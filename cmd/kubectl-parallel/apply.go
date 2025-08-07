package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

// collectManifests reads all provided files, grouping the contained
// Kubernetes resources by label. Resources from multiple files are
// merged together so that none are lost.
func collectManifests(files []string, label string) (resourceGroups, error) {
	manifests := make(resourceGroups)

	for _, file := range files {
		var manifest io.Reader

		if strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://") {
			resp, err := http.Get(file)
			if err != nil {
				return nil, err
			}
			if resp.StatusCode < 200 || resp.StatusCode > 299 {
				return nil, fmt.Errorf("unable to read URL %s, server reported %d", file, resp.StatusCode)
			}

			defer resp.Body.Close()
			manifest = resp.Body
		} else if file == "-" {
			manifest = os.Stdin
		} else {
			var err error
			manifest, err = os.Open(file)
			if err != nil {
				return nil, err
			}
		}

		groups, err := groupManifests(manifest, label)
		if err != nil {
			return nil, err
		}

		for key, resources := range groups {
			for _, resource := range resources {
				manifests.insert(key, resource)
			}
		}
	}

	return manifests, nil
}

func NewApplyCommand(params *rootCommandParams) *cobra.Command {
	var files []string

	cmd := &cobra.Command{
		Use:   "apply [flags] -f pod.yaml",
		Short: fmt.Sprint("Apply resources in parallel using label."),
		RunE: func(cmd *cobra.Command, args []string) error {
			manifests, err := collectManifests(files, params.label)
			if err != nil {
				return err
			}

			g, _ := errgroup.WithContext(context.Background())

			for _, rawResources := range manifests {
				g.Go(func() error {
					resourceBuffer := bytes.NewBuffer(nil)

					for _, rawResource := range rawResources {
						_, err := resourceBuffer.Write(rawResource)
						if err != nil {
							return err
						}

						_, err = resourceBuffer.WriteString("\n---\n")
						if err != nil {
							return err
						}
					}

					kubectlArgs := append([]string{"apply", "-f", "-"}, args...)

					kubectlCommand := exec.Command("kubectl", kubectlArgs...)
					kubectlCommand.Stdin = resourceBuffer
					kubectlCommand.Stderr = os.Stderr
					kubectlCommand.Stdout = os.Stdout

					if err := kubectlCommand.Run(); err != nil {
						cmd.SilenceUsage = true
						cmd.SilenceErrors = true
						return err
					}

					return nil
				})
			}

			return g.Wait()
		},
	}

	cmd.PersistentFlags().StringSliceVarP(&files, "filename", "f", []string{}, "Filename or URL to files to use to create the resource (use - for STDIN)")

	return cmd
}
