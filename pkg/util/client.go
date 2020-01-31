package util

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/gabemontero/obu/pkg/api"
	imageset "github.com/openshift/client-go/image/clientset/versioned"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"


	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"

)

func GetConfig(cfg *api.Config) (*rest.Config, error) {
	if len(cfg.Kubeconfig) > 0 {
		return clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
	}
	// If an env variable is specified with the config locaiton, use that
	if len(os.Getenv("KUBECONFIG")) > 0 {
		return clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	}
	// If no explicit location, try the in-cluster config
	if c, err := rest.InClusterConfig(); err == nil {
		return c, nil
	}
	// If no in-cluster config, try the default location in the user's home directory
	if usr, err := user.Current(); err == nil {
		if c, err := clientcmd.BuildConfigFromFlags(
			"", filepath.Join(usr.HomeDir, ".kube", "config")); err == nil {
			return c, nil
		}
	}

	return nil, fmt.Errorf("could not locate a kubeconfig")
}

func GetImageClient(cfg *rest.Config) (imagev1.ImageV1Interface, error) {
	client, err := imageset.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return client.ImageV1(), nil
}

func GetCurrentProject() string {
	kubeConfigFlags := genericclioptions.NewConfigFlags(true)
	matchVersionKubeConfigFlags := kcmdutil.NewMatchVersionFlags(kubeConfigFlags)
	f := kcmdutil.NewFactory(matchVersionKubeConfigFlags)
	cfg, err := f.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr,"ERROR: could not get default project: %v", err)
		return ""
	}
	currentProject := ""
	currentContext := cfg.Contexts[cfg.CurrentContext]
	if currentContext != nil {
		currentProject = currentContext.Namespace
	}
	return currentProject
}

