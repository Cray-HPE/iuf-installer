/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/Cray-HPE/iuf-installer/internal"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "iuf-installer",
	Short: "Create a k3d cluster for IUF",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.iuf-installer.yaml)")
	rootCmd.PersistentFlags().StringVarP(&internal.AppConfig.Tarball, "tarfile", "t", "", "Path to tarfile")
	rootCmd.MarkPersistentFlagRequired("tarfile")

	rootCmd.PersistentFlags().BoolVarP(&internal.AppConfig.Force, "force", "f", false, "delete previous install and start over")
}
