package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/browser"
)

// Custom handler to serve files with a styled UI
func fileServerWithUI(dir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request
		log.Printf("[%s] %s %s\n", time.Now().Format(time.RFC3339), r.Method, r.URL.Path)

		// Get the requested path
		path := filepath.Join(dir, r.URL.Path)
		path, err := filepath.Abs(path)
		if err != nil {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		// Ensure the path is within the directory
		if !isPathInDir(path, dir) {
			http.Error(w, "Access denied: Path outside directory", http.StatusForbidden)
			return
		}

		// Check if the path is a file or directory
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		if info.IsDir() {
			// Serve directory listing
			serveDirectoryListing(w, r, path, dir)
		} else {
			// Serve the file
			http.ServeFile(w, r, path)
		}
	})
}

// Check if a path is within the base directory
func isPathInDir(path, baseDir string) bool {
	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return false
	}
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	return filepath.HasPrefix(pathAbs, baseAbs)
}

// Format file size in a human-readable way
func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
}

// Get path relative to home directory
func getRelativePath(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path // Fallback to absolute path if home dir can't be determined
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	if strings.HasPrefix(absPath, homeDir) {
		relPath, err := filepath.Rel(homeDir, absPath)
		if err != nil {
			return absPath
		}
		if relPath == "." {
			return "~/"
		}
		return "~/" + relPath
	}
	return absPath
}

// Serve a styled directory listing with Tailwind CSS and Google Font
func serveDirectoryListing(w http.ResponseWriter, r *http.Request, path, baseDir string) {
	files, err := os.ReadDir(path)
	if err != nil {
		http.Error(w, "Unable to read directory", http.StatusInternalServerError)
		return
	}

	// Get relative path for display
	displayPath := getRelativePath(path)

	// Generate HTML for directory listing
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>File Server - %s</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Recursive:wght@300..1000&display=swap" rel="stylesheet">
<style>
*{
font-family: 'Recursive', sans-serif;
}
</style>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 text-gray-900 min-h-screen p-6">
    <div class="max-w-5xl mx-auto">
        <h1 class="text-3xl font-bold mb-6">Directory: %s</h1>
        <div class="bg-white shadow-md rounded-lg overflow-hidden">
            <table class="w-full">
                <thead class="bg-blue-600 text-white">
                    <tr>
                        <th class="py-3 px-4 text-left">Name</th>
                        <th class="py-3 px-4 text-left">Size</th>
                        <th class="py-3 px-4 text-left">Last Modified</th>
                    </tr>
                </thead>
                <tbody>`, r.URL.Path, displayPath)

	// Add parent directory link if not at root
	if r.URL.Path != "/" {
		parentPath := filepath.Dir(r.URL.Path)
		fmt.Fprintf(w, `<tr class="hover:bg-gray-100">
            <td class="py-2 px-4"><a href="%s" class="flex items-center text-blue-500 hover:underline"><svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"></path></svg>..</a></td>
            <td class="py-2 px-4">-</td>
            <td class="py-2 px-4">-</td>
        </tr>`, parentPath)
	}

	// List files and directories
	for _, file := range files {
		name := file.Name()
		info, err := file.Info()
		if err != nil {
			continue
		}
		size := "-"
		if !file.IsDir() {
			size = formatFileSize(info.Size())
		}
		modTime := info.ModTime().Format("2006-01-02 15:04:05")
		link := filepath.Join(r.URL.Path, name)
		icon := `<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"></path></svg>`
		if file.IsDir() {
			name += "/"
			icon = `<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"></path></svg>`
		}
		fmt.Fprintf(w, `<tr class="hover:bg-gray-100">
            <td class="py-2 px-4"><a href="%s" class="flex items-center text-blue-500 hover:underline">%s%s</a></td>
            <td class="py-2 px-4">%s</td>
            <td class="py-2 px-4">%s</td>
        </tr>`, link, icon, name, size, modTime)
	}

	fmt.Fprintf(w, `</tbody></table></div></div></body></html>`)
}

func main() {
	// Define command-line flag for directory
	dir := flag.String("dir", ".", "Directory to serve files from")
	port := flag.Int("port", 8080, "Port to run the server on")
	flag.Parse()

	// Resolve absolute path of the directory
	absDir, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatalf("Invalid directory: %v", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s", absDir)
	}

	// Create HTTP server
	addr := fmt.Sprintf(":%d", *port)
	mux := http.NewServeMux()
	mux.Handle("/", fileServerWithUI(absDir))

	// Start the server in a goroutine
	go func() {
		log.Printf("Serving files from %s on http://localhost%s", absDir, addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Open the browser
	url := fmt.Sprintf("http://localhost:%d", *port)
	if err := browser.OpenURL(url); err != nil {
		log.Printf("Failed to open browser: %v", err)
	} else {
		log.Printf("Opened browser at %s", url)
	}

	// Keep the program running
	select {}
}
