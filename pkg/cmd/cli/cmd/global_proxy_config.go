package cmd

import (
	"fmt"
	"github.com/gabemontero/obu/pkg/api"
	"github.com/gabemontero/obu/pkg/util"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

func NewCmdGlobalProxyConfig(cfg *api.Config) *cobra.Command {
	proxyCmd := &cobra.Command{
		Use:   "proxy [<options>]",
		Short: "Access OpenShift global proxy configuration.",
		Long:  "Pull the various elements from OpenShift's global proxy configuration support.",
		Example: `
# List all three proxy related settings as if you were setting environment variables, where the hosts
# are listed if they have been successfully contacted by the global proxy operator
$ obu proxy --env-vars

# List only the HTTPS proxy host if the global proxy operator was able to connect to it
$ obu proxy --https-proxy-only

# List only the HTTP proxy host if the global proxy operator was able to connect to it
$ obu proxy --http-proxy-only

# List only the no proxy host list
$ obu proxy --no-proxy-only
`,
		Run: func(cmd *cobra.Command, args []string) {

			kubeconfig, err := util.GetConfig(cfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: problem with kubeconfig: %v\n", err)
				return
			}
			proxyClient, err := util.GetProxyClient(kubeconfig)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: problem with openshift proxy client: %v\n", err)
				return
			}
			proxyCfg, err := proxyClient.Get("cluster", metav1.GetOptions{})
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: problem retrieving openshift global proxy config: %v\n", err)
				return
			}
			if proxyCfg == nil {
				fmt.Fprintf(os.Stderr, "ERROR: problem retrieving openshift global proxy config: no error but nil reference\n")
				return
			}

			coreClient, err := util.GetCoreClient(kubeconfig)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: problem with k8s core client: %v\n", err)
				return
			}
			//TODO alternatively, we could allow for a creation of a CM in the specified namespace that has
			// the label 'config.openshift.io/inject-trusted-cabundle: "true"'
			ocmProxyCM, err := coreClient.CoreV1().ConfigMaps("openshift-controller-manager").Get(
				"openshift-global-ca", metav1.GetOptions{})
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: problem retrieving OCM globly proxy CA config map: %v\n", err)
				return
			}
			if ocmProxyCM == nil || len(ocmProxyCM.Data) == 0 {
				fmt.Fprintf(os.Stderr, "ERROR: proxy CA data is not available\n")
				return
			}
			globalCAData, exists := ocmProxyCM.Data["ca-bundle.crt"]
			if !exists {
				fmt.Fprintf(os.Stderr, "ERROR: proxy CA data has not been set\n")
				return
			}

			switch {
			case cfg.HttpsProxyOnly:
				fmt.Fprintf(os.Stdout, proxyCfg.Status.HTTPSProxy)
			case cfg.HttpProxyOnly:
				fmt.Fprintf(os.Stdout, proxyCfg.Status.HTTPProxy)
			case cfg.NoProxyOnly:
				fmt.Fprintf(os.Stdout, proxyCfg.Status.NoProxy)
			case cfg.CADataOnly:
				fmt.Fprintf(os.Stdout, globalCAData)
			case cfg.ENVVarsOnly:
				fmt.Fprintf(os.Stdout, "HTTPS_PROXY=%s\n", proxyCfg.Status.HTTPSProxy)
				fmt.Fprintf(os.Stdout, "HTTP_PROXY=%s\n", proxyCfg.Status.HTTPProxy)
				fmt.Fprintf(os.Stdout, "NO_PROXY=%s\n", proxyCfg.Status.NoProxy)
				fmt.Fprintf(os.Stdout, "https_proxy=%s\n", proxyCfg.Status.HTTPSProxy)
				fmt.Fprintf(os.Stdout, "http_proxy=%s\n", proxyCfg.Status.HTTPProxy)
				fmt.Fprintf(os.Stdout, "no_proxy=%s\n", proxyCfg.Status.NoProxy)
			default:
				util.DefaultMessage()
			}
		},
	}

	proxyCmd.Flags().BoolVar(&(cfg.HttpProxyOnly), "http-proxy", cfg.HttpProxyOnly,
		"Only list the HTTP proxy host if it is available.")
	proxyCmd.Flags().BoolVar(&(cfg.HttpProxyOnly), "https-proxy", cfg.HttpProxyOnly,
		"Only list the HTTPS proxy host if it is available.")
	proxyCmd.Flags().BoolVar(&(cfg.NoProxyOnly), "no-proxy", cfg.NoProxyOnly,
		"Only list the no proxy list if it is available.")
	proxyCmd.Flags().BoolVar(&(cfg.ENVVarsOnly), "env-vars", cfg.ENVVarsOnly,
		"Prints out bash style environment variable setting syntax for the well known proxy environment variables, using any available values.")
	proxyCmd.Flags().BoolVar(&(cfg.CADataOnly), "ca-data", cfg.CADataOnly,
		"Only list the raw CA CRT data (ca.crt contents) for accessing the HTTPS proxy.")

	return proxyCmd
}
