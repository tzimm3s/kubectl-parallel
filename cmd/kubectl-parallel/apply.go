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

func NewApplyCommand(params *rootCommandParams) *cobra.Command {
	var files []string

	cmd := &cobra.Command{
		Use:   "apply [flags] -f pod.yaml",
		Short: fmt.Sprint("Apply resources in parallel using label."),
		RunE: func(cmd *cobra.Command, args []string) error {
			var manifests resourceGroups

			for _, file := range files {
				var manifest io.Reader

				if strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://") {
					resp, err := http.Get(file)
					if err != nil {
						return err
					}
					if resp.StatusCode < 200 || resp.StatusCode > 299 {
						return fmt.Errorf("unable to read URL %s, server reported %d", file, resp.StatusCode)
					}

					defer resp.Body.Close()
					manifest = resp.Body
				} else if file == "-" {
					manifest = os.Stdin
				} else {
					var err error
					manifest, err = os.Open(file)
					if err != nil {
						return err
					}
				}

				var err error
				manifests, err = groupManifests(manifest, params.label)
				if err != nil {
					return err
				}
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
