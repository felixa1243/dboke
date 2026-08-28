package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	pluginName := "visual_query"
	pluginDir := filepath.Join(".", pluginName)
	
	fmt.Println("1. Compiling Go backend...")
	// Compile backend.exe from backend/main.go
	cmd := exec.Command("go", "build", "-o", filepath.Join(pluginDir, "backend.exe"), filepath.Join(pluginDir, "backend", "main.go"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to compile: %v\n", err)
		return
	}

	zipName := pluginName + ".zip"
	fmt.Printf("2. Creating bundle: %s\n", zipName)
	outFile, err := os.Create(zipName)
	if err != nil {
		fmt.Printf("Failed to create zip: %v\n", err)
		return
	}
	defer outFile.Close()

	w := zip.NewWriter(outFile)
	defer w.Close()

	// Add backend.exe to zip
	addFileToZip(w, filepath.Join(pluginDir, "backend.exe"), "backend.exe")
	
	// Add frontend/page.tsx to zip
	addFileToZip(w, filepath.Join(pluginDir, "frontend", "page.tsx"), "frontend/page.tsx")
	
	// Cleanup compiled backend.exe so the source directory stays clean
	os.Remove(filepath.Join(pluginDir, "backend.exe"))
	
	fmt.Println("Done! You can now upload", zipName, "in Dboke.")
}

func addFileToZip(w *zip.Writer, srcPath string, zipPath string) {
	f, err := os.Open(srcPath)
	if err != nil {
		fmt.Printf("Warning: Could not open %s\n", srcPath)
		return
	}
	defer f.Close()

	info, _ := f.Stat()
	header, _ := zip.FileInfoHeader(info)
	header.Name = zipPath
	header.Method = zip.Deflate

	writer, err := w.CreateHeader(header)
	if err != nil {
		fmt.Printf("Warning: Could not create header for %s\n", zipPath)
		return
	}

	io.Copy(writer, f)
}
