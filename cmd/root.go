/*
Copyright © 2023 Davi Araújo <davi.araujo13356@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "secret-edit",
	Short: "A CLI to edit GCP Secret Manager secrets",
	Run: smEdit,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	rootCmd.PersistentFlags().StringP("project", "p", "", "GCP Project ID")
	rootCmd.Flags().Bool("list", false, "List all secrets instead of editing a specific one")
}
