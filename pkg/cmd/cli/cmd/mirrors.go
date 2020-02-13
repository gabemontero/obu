package cmd

import (
	"bytes"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/BurntSushi/toml"
	"github.com/gabemontero/obu/pkg/api"
	"github.com/gabemontero/obu/pkg/util"

	"github.com/containers/image/pkg/sysregistriesv2"
	rutils "github.com/openshift/runtime-utils/pkg/registries"

	operatorv1alpha1 "github.com/openshift/api/operator/v1alpha1"
	configv1 "github.com/openshift/api/config/v1"

	"github.com/spf13/cobra"

)

func NewCmdMirrorRegistryConf(cfg *api.Config) *cobra.Command {
	regCmd := &cobra.Command{
		Use:     "mirror [<options>]",
		Short:   "Access OpenShift mirrors registry configuration.",
		Long:    "Pull various elements from OpenShift's mirrors registry configuration.",
		Example: `
# Print ca.crt file content for accessing the OpenShift mirror registry over HTTPS.  This will inspect
# config maps in the openshift-config namespace.
$ obu mirror --ca-data

# Print Docker config file content for authenticating with the OpenShift mirror registry.
$ obu mirror --docker-cfg-file
`,
		Run: func(cmd *cobra.Command, args []string) {
			kubeconfig, err := util.GetConfig(cfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: problem with kubeconfig: %v\n", err)
				return
			}
			imageConfigClient := util.GetImageConfigClient(kubeconfig)
			imageConfig, err := imageConfigClient.Get("cluster", metav1.GetOptions{})
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: problem getting global image config: %v\n", err)
				return
			}
			coreClient := util.GetCoreClient(kubeconfig)
			mirrorCA, err := coreClient.CoreV1().ConfigMaps("openshift-config").Get(
				imageConfig.Spec.AdditionalTrustedCA.Name, metav1.GetOptions{})
			if err != nil && errors.IsNotFound(err) {
				fmt.Fprintf(os.Stderr, "ERROR: problem getting mirror registry CAs: %v\n", err)
				return
			}
			mirrorClient := util.GetImageMirrorClient(kubeconfig)
			imageContentSourcePolicies, err := mirrorClient.List(
				metav1.ListOptions{LabelSelector: labels.Everything().String()})

			switch {
			case cfg.CADataOnly:
				for _, ca := range mirrorCA.Data {
					fmt.Fprintf(os.Stdout, ca)
				}
			case cfg.DockerConfigFile:
				polices := []*operatorv1alpha1.ImageContentSourcePolicy{}
				for _, policy := range imageContentSourcePolicies.Items {
					polices = append(polices, &policy)
				}
				content, err := createBuildRegistriesConfigData(imageConfig, polices)
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERROR: problem building registry config: %v\n", err)
					return
				}
				fmt.Fprintf(os.Stdout, content)
			default:
				util.DefaultMessage()
			}
		},
	}
	regCmd.Flags().BoolVar(&(cfg.CADataOnly), "ca-data", cfg.CADataOnly,
		"Only list the raw CA CRT data (ca.crt contents) for accessing the registry.")
	regCmd.Flags().BoolVar(&(cfg.DockerConfigFile), "docker-cfg-file", cfg.DockerConfigFile,
		"Only print the docker config file for pushing to/pulling from the image internal image registry)")
	return regCmd
}

func createBuildRegistriesConfigData(config *configv1.Image, policies []*operatorv1alpha1.ImageContentSourcePolicy) (string, error) {

	blockedRegs := []string{}
	insecureRegs := []string{}
	if config != nil {
		insecureRegs = config.Spec.RegistrySources.InsecureRegistries
		blockedRegs = config.Spec.RegistrySources.BlockedRegistries
	}
	if len(insecureRegs) == 0 && len(blockedRegs) == 0 && len(policies) == 0 {
		return "", nil
	}

	configObj := sysregistriesv2.V2RegistriesConf{}
	// docker.io must be the only entry in the registry search list
	// See https://github.com/openshift/builder/pull/40
	configObj.UnqualifiedSearchRegistries = []string{"docker.io"}
	err := rutils.EditRegistriesConfig(&configObj, insecureRegs, blockedRegs, policies)
	if err != nil {
		return "", err
	}

	var newData bytes.Buffer
	encoder := toml.NewEncoder(&newData)
	if err := encoder.Encode(configObj); err != nil {
		return "", err
	}

	if len(newData.Bytes()) == 0 {
		return "", nil
	}
	return string(newData.Bytes()), nil
}