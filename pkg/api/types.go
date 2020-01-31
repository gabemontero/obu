package api

type Config struct {
	Kubeconfig string

	OverrideLocal bool
	SHA bool

	Namespace string

}
