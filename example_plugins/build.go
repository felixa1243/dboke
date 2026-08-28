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
	pluginName := "data-seeder" // This should match what Dboke expects (kebab-case folder structure)
	sourceDir := "data_seeder" // Our local folder name
	
	fmt.Println("1. Compiling Go backend...")
	backendSrc := filepath.Join(sourceDir, "backend", "main.go")
	backendExec := filepath.Join(sourceDir, "backend.exe")
	
	cmd := exec.Command("go", "build", "-o", backendExec, backendSrc)
	cmd.Stdout = os.Stdout
	cmd.Env = append(os.Environ(), "GOWORK=off")
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
	addFileToZip(w, backendExec, "backend.exe")
	
	// Add frontend/page.tsx to zip
	addFileToZip(w, filepath.Join(sourceDir, "frontend", "page.tsx"), "frontend/page.tsx")
	
	// Add meta.json to zip
	addFileToZip(w, filepath.Join(sourceDir, "meta.json"), "meta.json")
	
	// Cleanup compiled backend.exe so the source directory stays clean
	os.Remove(backendExec)
	
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
