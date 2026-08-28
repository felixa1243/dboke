package handler

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type PluginHandler struct{}

func NewPluginHandler() *PluginHandler {
	return &PluginHandler{}
}

func (h *PluginHandler) UploadPlugin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, `{"error": "File too large or invalid multipart data"}`, http.StatusBadRequest)
		return
	}

	pluginName := r.FormValue("name")
	if pluginName == "" {
		http.Error(w, `{"error": "Plugin name is required"}`, http.StatusBadRequest)
		return
	}

	// Format plugin folder name (e.g., "Visual Query" -> "visual-query")
	folderName := strings.ToLower(strings.ReplaceAll(pluginName, " ", "-"))

	file, header, err := r.FormFile("executable") // Frontend still sends it under this key
	if err != nil {
		http.Error(w, `{"error": "Zip file is required"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	cwd, _ := os.Getwd()
	// Target directory: apps/plugins
	pluginsDir := filepath.Join(cwd, "..", "plugins")
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		http.Error(w, `{"error": "Failed to create plugins directory"}`, http.StatusInternalServerError)
		return
	}

	tempZipPath := filepath.Join(pluginsDir, header.Filename)
	dst, err := os.Create(tempZipPath)
	if err != nil {
		http.Error(w, `{"error": "Failed to create temp zip file"}`, http.StatusInternalServerError)
		return
	}
	
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		http.Error(w, `{"error": "Failed to write temp zip data"}`, http.StatusInternalServerError)
		return
	}
	dst.Close()

	pluginInstallDir := filepath.Join(pluginsDir, folderName)

	// Launch Background Task (Goroutine) to install
	go func() {
		defer os.Remove(tempZipPath) // Always clean up the temp zip

		// Unzip logic
		err := unzipFile(tempZipPath, pluginInstallDir)
		if err != nil {
			fmt.Printf("Background Plugin Install Failed for %s: %v\n", pluginName, err)
			return
		}
		
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

		fmt.Printf("Successfully installed plugin %s to %s\n", pluginName, pluginInstallDir)
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Plugin zip uploaded successfully. Installation started in the background.",
	})
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
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
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
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	
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

	var plugins []PluginInfo

	files, err := os.ReadDir(pluginsDir)
	if err == nil {
		for _, file := range files {
			if file.IsDir() {
				displayName := strings.Title(strings.ReplaceAll(file.Name(), "-", " "))
				
				status := "Active"
				if _, err := os.Stat(filepath.Join(pluginsDir, file.Name(), ".disabled")); err == nil {
					status = "Inactive"
				}

				plugins = append(plugins, PluginInfo{
					Id:          file.Name(),
					Name:        displayName,
					Status:      status,
					Description: "Plugin bundle containing backend logic and frontend assets.",
					Type:        "external",
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plugins)
}

func (h *PluginHandler) TogglePlugin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Get plugin ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid plugin ID", http.StatusBadRequest)
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
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 { // /api/v1/plugins/{id}
		http.Error(w, "Invalid plugin ID", http.StatusBadRequest)
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
		srcFile, _ := os.Open(path)
		defer srcFile.Close()
		dstFile, _ := os.Create(dstPath)
		defer dstFile.Close()
		io.Copy(dstFile, srcFile)
		return nil
	})
}


