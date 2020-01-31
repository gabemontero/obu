package cli

import (
	"github.com/gabemontero/obu/pkg/api"
	"github.com/gabemontero/obu/pkg/cmd/cli/cmd"
	"github.com/spf13/cobra"
)

func NewCmdCLI() *cobra.Command {
	cfg := &api.Config{}
	obu := &cobra.Command{
		Use: "obu",
		Long: "OpenShift Build Utilities (obu) is a tool that facilitate building images in a OpenShift cluster via\n" +
			" the Tekton framework.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	obu.PersistentFlags().StringVar(&(cfg.Kubeconfig), "kubeconfig", cfg.Kubeconfig,
		"Path to the kubeconfig file to use for CLI requests.")
	obu.AddCommand(cmd.NewCmdTranslateIST(cfg))

	return obu
}

func CommandFor() *cobra.Command {
	return NewCmdCLI()
}