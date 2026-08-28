package handler

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"io/ioutil"

	"dboke-api/internal/core/services"
	"dboke-api/internal/pkg/contextkeys"
	"dboke-api/internal/pkg/response"
	"dboke-api/internal/pkg/taskqueue"
	"github.com/go-chi/chi/v5"
	"log/slog"
)

type PluginHandler struct {
	pluginManager *services.PluginManager
	dbService     *services.DBService
}

func NewPluginHandler(pluginManager *services.PluginManager, dbService *services.DBService) *PluginHandler {
	return &PluginHandler{
		pluginManager: pluginManager,
		dbService:     dbService,
	}
}

func (h *PluginHandler) UploadPlugin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		response.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "File too large or invalid multipart data")
		return
	}

	file, header, err := r.FormFile("executable")
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Zip file is required")
		return
	}
	defer file.Close()

	cwd, _ := os.Getwd()
	// Target directory: apps/plugins
	pluginsDir := filepath.Join(cwd, "..", "plugins")
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create plugins directory")
		return
	}

	tempZipPath := filepath.Join(pluginsDir, header.Filename)
	dst, err := os.Create(tempZipPath)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create temp zip file")
		return
	}
	
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to write temp zip data")
		return
	}
	dst.Close()

	// Extract meta.json
	rZip, err := zip.OpenReader(tempZipPath)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid zip file")
		return
	}
	
	var meta struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	metaFound := false
	for _, f := range rZip.File {
		if f.Name == "meta.json" {
			rc, _ := f.Open()
			json.NewDecoder(rc).Decode(&meta)
			rc.Close()
			metaFound = true
			break
		}
	}
	rZip.Close()

	if !metaFound || meta.ID == "" {
		os.Remove(tempZipPath)
		response.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "meta.json with 'id' is required in the plugin bundle")
		return
	}

	folderName := meta.ID
	pluginInstallDir := filepath.Join(pluginsDir, folderName)

	taskId := taskqueue.EnqueueTask("Starting installation...", func(taskId string) {
		defer os.Remove(tempZipPath) // Always clean up the temp zip

		taskqueue.UpdateTask(taskId, taskqueue.StatusProcessing, 10, "Extracting plugin files...", "")
		
		err = unzipFile(tempZipPath, pluginInstallDir)
		if err != nil {
			fmt.Printf("Plugin Install Failed for %s: %v\n", meta.Name, err)
			taskqueue.UpdateTask(taskId, taskqueue.StatusFailed, 10, "Failed to extract plugin", err.Error())
			return
		}
		
		taskqueue.UpdateTask(taskId, taskqueue.StatusProcessing, 50, "Copying frontend assets...", "")

		// Dynamic routing hack for Next.js App Router!
		// Move frontend folder directly to Next.js app directory to hot-reload the React pages
		frontendSrc := filepath.Join(pluginInstallDir, "frontend")
		if stat, err := os.Stat(frontendSrc); err == nil && stat.IsDir() {
			nextjsDir := filepath.Join(cwd, "..", "..", "apps", "web", "app", "databases", "plugins", folderName)
			os.MkdirAll(nextjsDir, 0755)
			
			// Copy files
			copyDir(frontendSrc, nextjsDir)
			
			// Optional: Remove frontendSrc from plugins folder to save space
			os.RemoveAll(frontendSrc)
		}

		taskqueue.UpdateTask(taskId, taskqueue.StatusCompleted, 100, "Installation complete!", "")
		fmt.Printf("Successfully installed plugin %s to %s\n", meta.Name, pluginInstallDir)
	})

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Plugin zip uploaded successfully. Installation started.",
		"task_id": taskId,
	})
}

// GetTaskStatus handles checking the status of a background task
func (h *PluginHandler) GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	taskId := chi.URLParam(r, "id")
	task, ok := taskqueue.GetTask(taskId)
	if !ok {
		response.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Task not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// Helper to extract zip files safely
func unzipFile(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	os.MkdirAll(dest, 0755)

	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		path := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, 0755)
		} else {
			os.MkdirAll(filepath.Dir(path), 0755)
			fOut, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer fOut.Close()
			if _, err = io.Copy(fOut, rc); err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *PluginHandler) ListPlugins(w http.ResponseWriter, r *http.Request) {
	cwd, _ := os.Getwd()
	// Read from apps/plugins
	pluginsDir := filepath.Join(cwd, "..", "plugins")

	type PluginInfo struct {
		Id          string `json:"id"`
		Name        string `json:"name"`
		Status      string `json:"status"`
		Description string `json:"description"`
		Type        string `json:"type"`
	}

	plugins := []PluginInfo{}

	files, err := os.ReadDir(pluginsDir)
	if err == nil {
		for _, file := range files {
			if file.IsDir() {
				// Default values
				displayName := strings.Title(strings.ReplaceAll(file.Name(), "-", " "))
				description := "Plugin bundle containing backend logic and frontend assets."
				
				// Try to read meta.json
				metaPath := filepath.Join(pluginsDir, file.Name(), "meta.json")
				if metaData, err := ioutil.ReadFile(metaPath); err == nil {
					var meta struct {
						Name        string `json:"name"`
						Description string `json:"description"`
					}
					if json.Unmarshal(metaData, &meta) == nil {
						if meta.Name != "" {
							displayName = meta.Name
						}
						if meta.Description != "" {
							description = meta.Description
						}
					}
				}

				status := "Active"
				if _, err := os.Stat(filepath.Join(pluginsDir, file.Name(), ".disabled")); err == nil {
					status = "Inactive"
				}

				plugins = append(plugins, PluginInfo{
					Id:          file.Name(),
					Name:        displayName,
					Status:      status,
					Description: description,
					Type:        "external",
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plugins)
}

func (h *PluginHandler) TogglePlugin(w http.ResponseWriter, r *http.Request) {
	// Get plugin ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		response.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid plugin ID")
		return
	}
	pluginId := parts[4]

	cwd, _ := os.Getwd()
	pluginDir := filepath.Join(cwd, "..", "plugins", pluginId)
	disabledFile := filepath.Join(pluginDir, ".disabled")

	activeFrontendDir := filepath.Join(cwd, "..", "..", "apps", "web", "app", "databases", "plugins", pluginId)
	inactiveFrontendDir := filepath.Join(cwd, "..", "..", "apps", "web", "app", "databases", "plugins", "_"+pluginId)

	if _, err := os.Stat(disabledFile); err == nil {
		// Reactivate the plugin
		os.Remove(disabledFile)
		os.Rename(inactiveFrontendDir, activeFrontendDir)
	} else {
		// Deactivate the plugin
		os.WriteFile(disabledFile, []byte(""), 0644)
		// Next.js ignores folders starting with an underscore, so this safely disables the route!
		os.Rename(activeFrontendDir, inactiveFrontendDir)
	}
	w.WriteHeader(http.StatusOK)
}

func (h *PluginHandler) DeletePlugin(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 { // /api/v1/plugins/{id}
		response.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid plugin ID")
		return
	}
	pluginId := parts[4]

	cwd, _ := os.Getwd()
	// Delete backend folder
	os.RemoveAll(filepath.Join(cwd, "..", "plugins", pluginId))
	// Delete frontend next.js injected folder
	os.RemoveAll(filepath.Join(cwd, "..", "..", "apps", "web", "app", "databases", "plugins", pluginId))
	// Delete any inactive frontend folder too just in case
	os.RemoveAll(filepath.Join(cwd, "..", "..", "apps", "web", "app", "databases", "plugins", "_"+pluginId))

	w.WriteHeader(http.StatusOK)
}

// copyDir copies a directory recursively
func copyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()
		
		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()
		
		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return err
		}
		return nil
	})
}

func (h *PluginHandler) ExecutePluginQuery(w http.ResponseWriter, r *http.Request) {
	pluginID := chi.URLParam(r, "id")
	database := chi.URLParam(r, "database")
	
	sessionID, ok := r.Context().Value(contextkeys.SessionIDKey).(string)
	if !ok || sessionID == "" {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}

	// 1. Get the Plugin Feature
	feature, err := h.pluginManager.GetFeature(pluginID)
	if err != nil {
		slog.Error("Failed to get plugin feature", slog.String("error", err.Error()))
		response.WriteError(w, http.StatusInternalServerError, "PLUGIN_ERROR", "Failed to load plugin logic")
		return
	}

	// 2. Ask Plugin to Build Query
	sqlQuery, err := feature.BuildQuery(string(body))
	if err != nil {
		slog.Error("Plugin failed to build query", slog.String("error", err.Error()))
		response.WriteError(w, http.StatusInternalServerError, "PLUGIN_ERROR", err.Error())
		return
	}

	// 3. Execute the built query against the database
	adapter, err := h.dbService.GetAdapter(r.Context(), sessionID, database)
	if err != nil {
		slog.Error("Failed to get adapter", slog.String("error", err.Error()))
		response.WriteError(w, http.StatusInternalServerError, "DB_ERROR", "Database connection failed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	rs, err := adapter.ExecuteRaw(ctx, sqlQuery)
	if err != nil {
		slog.Error("Failed to execute plugin query", slog.String("error", err.Error()))
		response.WriteError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rs)
}


