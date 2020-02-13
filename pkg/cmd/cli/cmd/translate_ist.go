package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gabemontero/obu/pkg/api"
	"github.com/gabemontero/obu/pkg/util"

	"github.com/openshift/library-go/pkg/image/imageutil"
	imagehelpers "github.com/openshift/oc/pkg/helpers/image"
)

func NewCmdTranslateIST(cfg *api.Config) *cobra.Command {
	translateCmd := &cobra.Command{
		Use:   "translate <imagestreamtag> [<options>]",
		Short: "Translate an image stream tag",
		Long:  "Translate an image stream reference to an image reference that can be pulled from an image registry.",
		Example: `
# Translate an image stream tag that exists in the current project/namespace
$ obu translate mystream:latest

# Translate an image stream tag that exists in another namespace
$ obu translate nodejs:12 -n openshift
`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Fprintf(os.Stderr, "ERROR: not enough arguments: %s\n", cmd.Use)
				return
			}
			istName := args[0]
			stream, tag, ok := imageutil.SplitImageStreamTag(istName)
			if !ok {
				fmt.Fprint(os.Stderr, "ERROR: invalid image stream tag reference (use '<stream>:<tag>'): %s\n", istName)
				return
			}
			kubeconfig, err := util.GetConfig(cfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: problem with kubeconfig: %v\n", err)
				return
			}
			imageClient := util.GetImageClient(kubeconfig)
			namespace := cfg.Namespace
			if len(namespace) == 0 {
				namespace = util.GetCurrentProject()
				if len(namespace) == 0 {
					return
				}
			}

			is, err := imageClient.ImageStreams(namespace).Get(stream, metav1.GetOptions{})
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: problem retrieving image stream %s: %v\n", stream, err)
				return
			}
			// use source tag regardless
			if cfg.OverrideLocal {
				_, tagRef, _, err := imagehelpers.FollowTagReference(is, tag)
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERROR: image stream tag %s had tag reference error: %v\n", istName, err)
					return
				}
				if tagRef == nil {
					fmt.Fprintf(os.Stderr, "ERROR: image stream tag %s has no tag references\n", istName)
				}

				if !cfg.SHA {
					fmt.Fprintf(os.Stdout, tagRef.From.Name)
					return
				}
				latestGen := int64(0)
				latestGenImage := ""
				for _, tagStatus := range is.Status.Tags {
					if tagStatus.Tag == tag {
						for _, item := range tagStatus.Items {
							if item.Generation > latestGen {
								latestGen = item.Generation
								latestGenImage = item.DockerImageReference
							}
						}
					}
				}
				fmt.Fprintf(os.Stdout, latestGenImage)
				return
			}
			// use local tag reference policy if available
			img, ok := imageutil.ResolveLatestTaggedImage(is, tag)
			if !ok {
				fmt.Fprintf(os.Stderr, "ERROR: unable to resolve image stream tag %s\n", istName)
				return
			}
			fmt.Fprintf(os.Stdout, img)
		},
	}
	translateCmd.Flags().BoolVar(&(cfg.OverrideLocal), "override-local", cfg.OverrideLocal,
		"Bypass local copy of image in OpenShift Internal registry and return external registry reference.")
	translateCmd.Flags().BoolVar(&(cfg.SHA), "sha-vs-tag", cfg.SHA,
		"End the translated image reference with the SHA instead of the tag name.")
	translateCmd.Flags().StringVarP(&(cfg.Namespace), "namespace", "n", "",
		"Specify the namespace the image stream is located in")

	return translateCmd
}
