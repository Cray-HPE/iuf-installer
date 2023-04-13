/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/Cray-HPE/iuf-installer/internal"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: casmInit,
}

func init() {
	//initCmd.Flags().AddFlagSet(k3dclustercmd.NewCmdClusterCreate().Flags())
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")
}

func casmInit(cmd *cobra.Command, args []string) error {
	// configure podman
	err := internal.PodmanServiceInstance.PodmanInit()
	if err != nil {
		return err
	}
	// load local image into podman
	err = internal.ImageServiceInstance.LoadImages()
	if err != nil {
		return err
	}
	// start k3d server
	// print out kubeconfig and kube info
	return nil
}
