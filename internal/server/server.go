package server

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/hc/hc/internal/logger"
	"github.com/hc/hc/internal/models"
	"github.com/hc/hc/internal/proxy"
	"github.com/hc/hc/internal/storage"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	port        int
	db          *storage.DB
	proxyClient *proxy.Client
	frontendFS  fs.FS
}

func New(port int, db *storage.DB, frontendFS fs.FS) *Server {
	return &Server{
		port:        port,
		db:          db,
		proxyClient: proxy.NewClient(),
		frontendFS:  frontendFS,
	}
}

func (s *Server) Start() error {
	e := echo.New()
	log := logger.Get()

	// Disable Echo's default logger
	e.HideBanner = true
	e.HidePort = true

	// Request logging middleware
	e.Use(logger.EchoMiddleware())

	// CORS middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{"Content-Type"},
	}))

	// API routes
	api := e.Group("/api")
	
	// Proxy request
	api.POST("/request", s.handleProxyRequest)
	
	// Requests endpoints
	api.GET("/requests", s.handleGetRequests)
	api.POST("/requests", s.handleCreateRequest)
	api.GET("/requests/:id", s.handleGetRequestByID)
	api.PUT("/requests/:id", s.handleUpdateRequestByID)
	api.DELETE("/requests/:id", s.handleDeleteRequestByID)
	
	// Folders endpoints
	api.GET("/folders", s.handleGetFolders)
	api.POST("/folders", s.handleCreateFolder)
	api.GET("/folders/:id", s.handleGetFolderByID)
	api.PUT("/folders/:id", s.handleUpdateFolderByID)
	api.DELETE("/folders/:id", s.handleDeleteFolderByID)

	// Static files
	e.GET("/*", s.handleStatic)

	addr := fmt.Sprintf(":%d", s.port)
	log.Info("Starting server", slog.String("address", addr))
	return e.Start(addr)
}

// Proxy handlers
func (s *Server) handleProxyRequest(c echo.Context) error {
	log := logger.Get()
	
	var proxyReq proxy.ProxyRequest
	if err := c.Bind(&proxyReq); err != nil {
		log.Error("Failed to bind proxy request", slog.String("error", err.Error()))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := proxy.ValidateURL(proxyReq.URL); err != nil {
		log.Error("Invalid URL", slog.String("url", proxyReq.URL), slog.String("error", err.Error()))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := proxy.ValidateMethod(proxyReq.Method); err != nil {
		log.Error("Invalid HTTP method", slog.String("method", proxyReq.Method), slog.String("error", err.Error()))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	log.Info("Proxying request", slog.String("method", proxyReq.Method), slog.String("url", proxyReq.URL))
	
	resp, err := s.proxyClient.ProxyRequest(&proxyReq)
	if err != nil {
		log.Error("Proxy request failed", slog.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to execute request: %v", err)})
	}

	return c.JSON(http.StatusOK, resp)
}

// Request handlers
func (s *Server) handleGetRequests(c echo.Context) error {
	requests, err := s.db.GetRequests()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get requests"})
	}

	// Ensure we return an empty array instead of null
	if requests == nil {
		requests = []models.Request{}
	}

	return c.JSON(http.StatusOK, requests)
}

func (s *Server) handleCreateRequest(c echo.Context) error {
	var request models.Request
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := s.db.CreateRequest(&request); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create request"})
	}

	return c.JSON(http.StatusCreated, request)
}

func (s *Server) handleGetRequestByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request ID"})
	}

	var request models.Request
	if err := s.db.GetRequest(id, &request); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Request not found"})
	}

	return c.JSON(http.StatusOK, request)
}

func (s *Server) handleUpdateRequestByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request ID"})
	}

	var request models.Request
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	request.ID = id
	if err := s.db.UpdateRequest(&request); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update request"})
	}

	return c.JSON(http.StatusOK, request)
}

func (s *Server) handleDeleteRequestByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request ID"})
	}

	if err := s.db.DeleteRequest(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete request"})
	}

	return c.NoContent(http.StatusNoContent)
}

// Folder handlers
func (s *Server) handleGetFolders(c echo.Context) error {
	folders, err := s.db.GetFolders()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get folders"})
	}

	// Ensure we return an empty array instead of null
	if folders == nil {
		folders = []models.Folder{}
	}

	return c.JSON(http.StatusOK, folders)
}

func (s *Server) handleCreateFolder(c echo.Context) error {
	var folder models.Folder
	if err := c.Bind(&folder); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := s.db.CreateFolder(&folder); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create folder"})
	}

	return c.JSON(http.StatusCreated, folder)
}

func (s *Server) handleGetFolderByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid folder ID"})
	}

	var folder models.Folder
	if err := s.db.GetFolder(id, &folder); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Folder not found"})
	}

	return c.JSON(http.StatusOK, folder)
}

func (s *Server) handleUpdateFolderByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid folder ID"})
	}

	var folder models.Folder
	if err := c.Bind(&folder); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	folder.ID = id
	if err := s.db.UpdateFolder(&folder); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update folder"})
	}

	return c.JSON(http.StatusOK, folder)
}

func (s *Server) handleDeleteFolderByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid folder ID"})
	}

	if err := s.db.DeleteFolder(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete folder"})
	}

	return c.NoContent(http.StatusNoContent)
}

// Static file handler
func (s *Server) handleStatic(c echo.Context) error {
	path := c.Request().URL.Path
	if path == "/" {
		path = "/index.html"
	}

	if s.frontendFS != nil {
		file, err := s.frontendFS.Open(strings.TrimPrefix(path, "/"))
		if err != nil {
			// If file not found, serve index.html for client-side routing
			file, err = s.frontendFS.Open("index.html")
			if err != nil {
				return c.NoContent(http.StatusNotFound)
			}
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		// Set appropriate content type based on file extension
		if strings.HasSuffix(path, ".css") {
			c.Response().Header().Set("Content-Type", "text/css")
		} else if strings.HasSuffix(path, ".js") {
			c.Response().Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(path, ".html") {
			c.Response().Header().Set("Content-Type", "text/html")
		} else if strings.HasSuffix(path, ".woff2") {
			c.Response().Header().Set("Content-Type", "font/woff2")
		}

		http.ServeContent(c.Response(), c.Request(), stat.Name(), stat.ModTime(), file.(io.ReadSeeker))
		return nil
	} else {
		return c.File("frontend/out" + path)
	}
}