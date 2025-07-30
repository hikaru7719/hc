package cmd

import (
	"fmt"
	"io/fs"
	"log"

	"github.com/hc/hc/internal/server"
	"github.com/hc/hc/internal/storage"
	"github.com/spf13/cobra"
)

var (
	port          int
	GetFrontendFS func() (fs.FS, error)
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP client server",
	Long:  `Start the local web server that hosts the HTTP client interface.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := storage.InitDB()
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		var frontendFS fs.FS
		if GetFrontendFS != nil {
			var err error
			frontendFS, err = GetFrontendFS()
			if err != nil {
				log.Printf("Warning: Frontend files not embedded, will serve from filesystem: %v", err)
				frontendFS = nil
			}
		} else {
			log.Printf("Warning: Frontend files not embedded, will serve from filesystem")
			frontendFS = nil
		}

		srv := server.New(port, db, frontendFS)

		log.Printf("Starting HC server on port %d", port)
		log.Printf("Open http://localhost:%d in your browser", port)

		return srv.Start()
	},
}

func init() {
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the server on")
}

func AddToRoot(rootCmd *cobra.Command) {
	rootCmd.AddCommand(serveCmd)
}
