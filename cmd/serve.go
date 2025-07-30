package cmd

import (
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/hc/hc/internal/logger"
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
		log := logger.Get()
		
		log.Info("Initializing database")
		db, err := storage.InitDB()
		if err != nil {
			log.Error("Failed to initialize database", slog.String("error", err.Error()))
			return err
		}
		defer db.Close()

		var frontendFS fs.FS
		if GetFrontendFS != nil {
			var err error
			frontendFS, err = GetFrontendFS()
			if err != nil {
				log.Warn("Frontend files not embedded, will serve from filesystem", slog.String("error", err.Error()))
				frontendFS = nil
			}
		} else {
			log.Warn("Frontend files not embedded, will serve from filesystem")
			frontendFS = nil
		}

		srv := server.New(port, db, frontendFS)

		log.Info("Starting HC server", 
			slog.Int("port", port),
			slog.String("url", fmt.Sprintf("http://localhost:%d", port)),
		)

		return srv.Start()
	},
}

func init() {
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the server on")
}

func AddToRoot(rootCmd *cobra.Command) {
	rootCmd.AddCommand(serveCmd)
}
