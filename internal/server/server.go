package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"github.com/hc/hc/internal/models"
	"github.com/hc/hc/internal/proxy"
	"github.com/hc/hc/internal/storage"
)

type Server struct {
	port       int
	db         *storage.DB
	proxyClient *proxy.Client
	frontendFS fs.FS
}

func New(port int, db *storage.DB, frontendFS fs.FS) *Server {
	return &Server{
		port:       port,
		db:         db,
		proxyClient: proxy.NewClient(),
		frontendFS: frontendFS,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/request", s.corsMiddleware(s.handleProxyRequest))
	mux.HandleFunc("/api/requests", s.corsMiddleware(s.handleRequests))
	mux.HandleFunc("/api/requests/", s.corsMiddleware(s.handleRequestByID))
	mux.HandleFunc("/api/folders", s.corsMiddleware(s.handleFolders))
	mux.HandleFunc("/api/folders/", s.corsMiddleware(s.handleFolderByID))

	mux.HandleFunc("/", s.handleStatic)

	addr := fmt.Sprintf(":%d", s.port)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

func (s *Server) handleProxyRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var proxyReq proxy.ProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&proxyReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := proxy.ValidateURL(proxyReq.URL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := proxy.ValidateMethod(proxyReq.Method); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.proxyClient.ProxyRequest(&proxyReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute request: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleRequests(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		requests, err := s.db.GetRequests()
		if err != nil {
			http.Error(w, "Failed to get requests", http.StatusInternalServerError)
			return
		}

		// Ensure we return an empty array instead of null
		if requests == nil {
			requests = []models.Request{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(requests)

	case "POST":
		var request models.Request
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := s.db.CreateRequest(&request); err != nil {
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(request)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRequestByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/requests/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		var request models.Request
		if err := s.db.GetRequest(id, &request); err != nil {
			http.Error(w, "Request not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(request)

	case "PUT":
		var request models.Request
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		request.ID = id
		if err := s.db.UpdateRequest(&request); err != nil {
			http.Error(w, "Failed to update request", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(request)

	case "DELETE":
		if err := s.db.DeleteRequest(id); err != nil {
			http.Error(w, "Failed to delete request", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleFolders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		folders, err := s.db.GetFolders()
		if err != nil {
			http.Error(w, "Failed to get folders", http.StatusInternalServerError)
			return
		}

		// Ensure we return an empty array instead of null
		if folders == nil {
			folders = []models.Folder{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(folders)

	case "POST":
		var folder models.Folder
		if err := json.NewDecoder(r.Body).Decode(&folder); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := s.db.CreateFolder(&folder); err != nil {
			http.Error(w, "Failed to create folder", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(folder)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleFolderByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/folders/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid folder ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		var folder models.Folder
		if err := s.db.GetFolder(id, &folder); err != nil {
			http.Error(w, "Folder not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(folder)

	case "PUT":
		var folder models.Folder
		if err := json.NewDecoder(r.Body).Decode(&folder); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		folder.ID = id
		if err := s.db.UpdateFolder(&folder); err != nil {
			http.Error(w, "Failed to update folder", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(folder)

	case "DELETE":
		if err := s.db.DeleteFolder(id); err != nil {
			http.Error(w, "Failed to delete folder", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	if s.frontendFS != nil {
		file, err := s.frontendFS.Open(strings.TrimPrefix(path, "/"))
		if err != nil {
			// If file not found, serve index.html for client-side routing
			file, err = s.frontendFS.Open("index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set appropriate content type based on file extension
		if strings.HasSuffix(path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		} else if strings.HasSuffix(path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(path, ".html") {
			w.Header().Set("Content-Type", "text/html")
		} else if strings.HasSuffix(path, ".woff2") {
			w.Header().Set("Content-Type", "font/woff2")
		}

		http.ServeContent(w, r, stat.Name(), stat.ModTime(), file.(io.ReadSeeker))
	} else {
		http.ServeFile(w, r, "frontend/out"+path)
	}
}