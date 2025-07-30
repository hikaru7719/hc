package main

import (
	"fmt"
	"os"

	"github.com/hc/hc/cmd"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hc",
	Short: "HTTP Client - A browser-based GUI HTTP client",
	Long: `HC (HTTP Client) is a CLI tool that launches a local server
and provides a browser-based GUI for making HTTP requests.`,
}

func init() {
	cmd.GetFrontendFS = GetFrontendFS
	cmd.AddToRoot(rootCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
