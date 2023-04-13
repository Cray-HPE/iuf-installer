/*
 MIT License

 (C) Copyright 2022 Hewlett Packard Enterprise Development LP

 Permission is hereby granted, free of charge, to any person obtaining a
 copy of this software and associated documentation files (the "Software"),
 to deal in the Software without restriction, including without limitation
 the rights to use, copy, modify, merge, publish, distribute, sublicense,
 and/or sell copies of the Software, and to permit persons to whom the
 Software is furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included
 in all copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 OTHER DEALINGS IN THE SOFTWARE.
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
	initCmd.PersistentFlags().StringVarP(&internal.AppConfig.Tarball, "tarfile", "t", "", "Path to tarfile")
	initCmd.MarkPersistentFlagRequired("tarfile")

	initCmd.PersistentFlags().BoolVarP(&internal.AppConfig.Force, "force", "f", false, "delete previous install and start over")
}

func casmInit(cmd *cobra.Command, args []string) error {
	PodmanServiceInstance := internal.NewPodmanService()
	ImageServiceInstance := internal.NewImageService(PodmanServiceInstance)

	// configure podman
	err := PodmanServiceInstance.PodmanInit()
	if err != nil {
		return err
	}
	// load local image into podman
	err = ImageServiceInstance.LoadImages()
	if err != nil {
		return err
	}
	// start k3d server
	// print out kubeconfig and kube info
	return nil
}
