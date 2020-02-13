package cmd

import (
	"fmt"
	"github.com/gabemontero/obu/pkg/api"
	"github.com/gabemontero/obu/pkg/util"
	buildv1 "github.com/openshift/api/build/v1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/kubernetes/pkg/credentialprovider"
	//credentialprovidersecrets "k8s.io/kubernetes/pkg/credentialprovider/secrets"
	kubeset "k8s.io/client-go/kubernetes"
	"os"
)

func NewCmdInternalRegistry(cfg *api.Config) *cobra.Command {
	regCmd := &cobra.Command{
		Use:     "registry [<options>]",
		Short:   "Access OpenShift internal registry configuration.",
		Long:    "Pull various elements from OpenShift's internal registry configuration.",
		Example: ``,
		Run: func(cmd *cobra.Command, args []string) {
			kubeconfig, err := util.GetConfig(cfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: problem with kubeconfig: %v\n", err)
				return
			}
			coreClient, err := util.GetCoreClient(kubeconfig)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: problem with k8s core client: %v\n", err)
				return
			}
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
				fmt.Fprintf(os.Stdout,registryCAData)
			default:
				util.DefaultMessage()
			}
			fmt.Fprintf(os.Stdout,registryCAData)
		},
	}
	regCmd.Flags().BoolVar(&(cfg.CADataOnly), "ca-data", cfg.CADataOnly,
		"Only list the raw CA CRT data (ca.crt contents) for accessing the registry.")
	return regCmd
}

// from build controller in OCM
func resolveImageSecretAsReference(cfg *api.Config, imageName, saName string, coreClient kubeset.Clientset) error {
	if len(saName) == 0 {
		saName = "builder"
	}
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
		secret, err := coreClient.CoreV1().Secrets(namespace).Get(ref.Namespace, metav1.GetOptions{})
		if err != nil {
			return err
		}
		secrets = append(secrets, *secret)
	}
	secret := &corev1.Secret{}
	if len(imageName) > 0 {
		/*emptyKeyring := credentialprovider.BasicDockerKeyring{}
		for _, s := range secrets {
			secretList := []corev1.Secret{s}
			keyring, err := credentialprovidersecrets.MakeDockerKeyring(secretList, &emptyKeyring)
			if err != nil {
				continue
			}
			if _, found := keyring.Lookup(image); found {
				secret = &s
				break
			}
		}*/
	}
	if secret == nil {
		for _, builderSecret := range secrets {
			if builderSecret.Type == corev1.SecretTypeDockercfg || builderSecret.Type == corev1.SecretTypeDockerConfigJson {
				secret = &builderSecret
				break
			}
		}
		if secret == nil {
			return fmt.Errorf("ERROR: No docker secrets associated with build service account %s", saName)
		}
	}
	return nil
}
