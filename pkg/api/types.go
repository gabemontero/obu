package api

type Config struct {
	Kubeconfig string

	// imagestream translate specific
	OverrideLocal bool
	SHA bool

	// proxy config specific
	HttpProxyOnly bool
	HttpsProxyOnly bool
	NoProxyOnly bool
	ENVVarsOnly bool

	// both proxy and image registry
	CADataOnly bool

	// image registry
	DockerConfigFile bool

	Namespace string

}
