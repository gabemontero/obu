package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gabemontero/obu/pkg/api"
	"github.com/gabemontero/obu/pkg/util"
	buildv1 "github.com/openshift/api/build/v1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeset "k8s.io/client-go/kubernetes"
)

func NewCmdInternalRegistry(cfg *api.Config) *cobra.Command {
	regCmd := &cobra.Command{
		Use:     "registry [<options>]",
		Short:   "Access OpenShift internal registry configuration.",
		Long:    "Pull various elements from OpenShift's internal registry configuration.",
		Example: `
# Print ca.crt file content for accessing the OpenShift internal registry over HTTPS.  This will inspect
# config maps in the openshift-controller-manager namespace.
$ obu registry --ca-data

# Print Docker config file content for authenticating with the OpenShift internal registry.
$ obu registry --docker-cfg-file
`,
		Run: func(cmd *cobra.Command, args []string) {
			kubeconfig, err := util.GetConfig(cfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: problem with kubeconfig: %v\n", err)
				return
			}
			coreClient := util.GetCoreClient(kubeconfig)
			registryCAMap, err := coreClient.CoreV1().ConfigMaps("openshift-controller-manager").Get("openshift-service-ca", metav1.GetOptions{})
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: problem retrieving registry CA config map: %v\n", err)
				return
			}
			if registryCAMap == nil || len(registryCAMap.Data) == 0 {
				fmt.Fprintf(os.Stderr, "ERROR: registry CA data is not available\n")
				return
			}
			registryCAData, exists := registryCAMap.Data[buildv1.ServiceCAKey]
			if !exists {
				fmt.Fprintf(os.Stderr, "ERROR: registry CA data has not been set\n")
				return
			}
			switch {
			case cfg.CADataOnly:
				fmt.Fprintf(os.Stdout, registryCAData)
			case cfg.DockerConfigFile:
				err := dumpBuilderDockerCfg(cfg, coreClient)
				if err != nil {
					fmt.Fprintf(os.Stderr, err.Error())
				}
			default:
				util.DefaultMessage()
			}
		},
	}
	regCmd.Flags().BoolVar(&(cfg.CADataOnly), "ca-data", cfg.CADataOnly,
		"Only list the raw CA CRT data (ca.crt contents) for accessing the registry.")
	regCmd.Flags().BoolVar(&(cfg.DockerConfigFile), "docker-cfg-file", cfg.DockerConfigFile,
		"Only print the docker config file for pushing to/pulling from the image internal image registry)")
	regCmd.Flags().StringVarP(&(cfg.Namespace), "namespace", "n", "",
		"Specify the namespace whose OpenShift builder service account should be inspected for docker authentication config")
	return regCmd
}

// elements from build controller in OCM
func dumpBuilderDockerCfg(cfg *api.Config, coreClient *kubeset.Clientset) error {
	saName := "builder"
	namespace := cfg.Namespace
	if len(namespace) == 0 {
		namespace = util.GetCurrentProject()
		if len(namespace) == 0 {
			return fmt.Errorf("ERROR: Need a namespace to fetch registry secrets")
		}
	}
	secrets := []corev1.Secret{}
	sa, err := coreClient.CoreV1().ServiceAccounts(namespace).Get(saName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	for _, ref := range sa.Secrets {
		secret, err := coreClient.CoreV1().Secrets(namespace).Get(ref.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		secrets = append(secrets, *secret)
	}
	for _, builderSecret := range secrets {
		if builderSecret.Type == corev1.SecretTypeDockercfg || builderSecret.Type == corev1.SecretTypeDockerConfigJson {
			contents, err := getDockerConfigFileStringForImageRegistryHost(&builderSecret)
			if err != nil {
				return err
			}
			if len(contents) == 0 {
				continue
			}
			fmt.Fprintf(os.Stdout, contents)
			return nil
		}
	}
	return fmt.Errorf("ERROR: No image registry docker secrets associated with build service account %s", saName)
}

func getDockerConfigFileStringForImageRegistryHost(secret *corev1.Secret) (string, error) {
	key := ""
	switch secret.Type {
	case corev1.SecretTypeDockercfg:
		key = corev1.DockerConfigKey
	case corev1.SecretTypeDockerConfigJson:
		key = corev1.DockerConfigJsonKey
	}
	secretEncodedData, exists := secret.Data[key]
	if !exists {
		return "", fmt.Errorf("ERROR: No data at key %s for secret %s", key, secret.Name)
	}
	secretDecodedData := []byte{}
	var err error
	switch secret.Type {
	case corev1.SecretTypeDockercfg:
		dockercfg := DockerConfig{}
		err = json.Unmarshal(secretEncodedData, &dockercfg)
		found := hasImageRegistryHost(dockercfg)
		if !found {
			return "", nil
		}
		secretDecodedData, err = json.MarshalIndent(dockercfg, "", "\t")
	case corev1.SecretTypeDockerConfigJson:
		dockercfgjson := DockerConfigJson{}
		err = json.Unmarshal(secretEncodedData, &dockercfgjson)
		found := hasImageRegistryHost(dockercfgjson.Auths)
		if !found {
			return "", nil
		}
		secretDecodedData, err = json.MarshalIndent(dockercfgjson, "", "\t")
	}
	if err != nil {
		return "", fmt.Errorf("ERROR: Problem decoding data at key %s for secret %s: %v", key, secret.Name, err)
	}
	return string(secretDecodedData), nil
}

func hasImageRegistryHost(auths DockerConfig) bool {
	for hostPort := range auths {
		if strings.HasPrefix(hostPort, "image-registry.openshift-image-registry") {
			return true
		}
	}
	return false
}

// similar structs in kubernetes/kubernetes, docker/docker, or containers/image, but either not public or too dicey to
// go.mod ... each of those three has its own copy of this
type AuthConfig struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Auth     string `json:"auth,omitempty"`

	// Email is an optional value associated with the username.
	// This field is deprecated and will be removed in a later
	// version of docker.
	Email string `json:"email,omitempty"`

	ServerAddress string `json:"serveraddress,omitempty"`

	// IdentityToken is used to authenticate the user and get
	// an access token for the registry.
	IdentityToken string `json:"identitytoken,omitempty"`

	// RegistryToken is a bearer token to be sent to a registry
	RegistryToken string `json:"registrytoken,omitempty"`
}

type DockerConfigJson struct {
	Auths DockerConfig `json:"auths"`
	// +optional
	HttpHeaders map[string]string `json:"HttpHeaders,omitempty"`
}

// DockerConfig represents the config file used by the docker CLI.
// This config that represents the credentials that should be used
// when pulling images from specific image repositories.
type DockerConfig map[string]DockerConfigEntry

type DockerConfigEntry struct {
	Username string               `json:"username"`
	Password string               `json:"password"`
	Email    string               `json:"email"`
	Provider DockerConfigProvider `json:"provider,omitempty"`
}
type DockerConfigProvider interface {
	// Enabled returns true if the config provider is enabled.
	// Implementations can be blocking - e.g. metadata server unavailable.
	Enabled() bool
	// Provide returns docker configuration.
	// Implementations can be blocking - e.g. metadata server unavailable.
	// The image is passed in as context in the event that the
	// implementation depends on information in the image name to return
	// credentials; implementations are safe to ignore the image.
	Provide(image string) DockerConfig
}
