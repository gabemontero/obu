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
	CADataOnly bool
	ENVVarsOnly bool

	Namespace string

}
