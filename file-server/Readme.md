
# FileServer

A simple, lightweight Go-based web server that serves files from a specified directory with a modern, responsive UI. It automatically opens the default browser to the server's URL, displays directory paths relative to the user's home directory, and logs requests to the console.

## Features

- **File and Directory Serving**: Serves files and directories from any specified path with a clean, table-based UI.
- **Responsive UI**: Built with [Tailwind CSS](https://tailwindcss.com/) for a modern, mobile-friendly design.
- **Typography**: Uses [Google Fonts' Roboto](https://fonts.google.com/specimen/Roboto) for professional typography.
- **Icons**: Displays SVG icons for files and folders to enhance usability.
- **Path Display**: Shows directory paths relative to the home directory (e.g., `~/Documents/folder`) or absolute paths for directories outside the home.
- **Request Logging**: Logs HTTP requests with timestamps, methods, and paths to the console.
- **Browser Auto-Open**: Automatically opens the default browser to the server's URL.
- **Security**: Prevents directory traversal attacks by restricting access to the specified directory.
- **Human-Readable File Sizes**: Displays file sizes in bytes, KB, MB, or GB.

## Requirements

- **Go**: Version 1.16 or later to compile the source code.
- **Git**: For cloning the repository (optional).
- **Browser**: A modern web browser (e.g., Chrome, Firefox, Safari) for viewing the UI.
- **Internet Access**: Required at runtime to load Tailwind CSS (via CDN) and Roboto font.

## Installation

1. **Clone the Repository** (optional, if hosted in a repo):
   ```bash
   git clone https://github.com/<your-username>/fileserver.git
   cd fileserver
   ```

2. **Install the Dependency**:
   The script uses the `github.com/pkg/browser` package to open the browser automatically.
   ```bash
   go get github.com/pkg/browser
   ```

3. **Compile the Binary**:
   Compile the `server.go` file into an executable binary named `fileserver`.
   ```bash
   go build -o fileserver server.go
   ```

4. **Make the Binary Globally Accessible**:
   To run `fileserver` from any directory, follow the instructions in [usage.md](usage.md) to move the binary to a directory in your system's PATH.

## Usage

Run the server by specifying a directory and optional port:
```bash
fileserver -dir=/path/to/your/folder -port=8080
```
- The browser will open to `http://localhost:8080`.
- The UI displays the directory listing with a path like `~/your/folder` (if within the home directory) or an absolute path (e.g., `/var/log`).
- Console logs show request details, e.g.:
  ```
  2025/11/07 17:09:45 Serving files from /path/to/your/folder on http://localhost:8080
  2025/11/07 17:09:45 Opened browser at http://localhost:8080
  [2025-11-07T17:09:47+0545] GET /
  ```

For detailed instructions on making the binary accessible from anywhere, see [usage.md](usage.md).

## Flags

- `-dir string`: Directory to serve files from (default: current directory `.`).
- `-port int`: Port to run the server on (default: `8080`).

Example:
```bash
fileserver -dir=~/Documents -port=9090
```

## Notes

- **Tailwind CSS**: Loaded via CDN for simplicity. For production, consider a local Tailwind setup to reduce latency.
- **Performance**: Large directories may load slowly. Consider adding pagination or filtering for such cases.
- **Security**: The server restricts access to the specified directory but does not include authentication or HTTPS. Use these for public-facing servers.
- **Cross-Platform**: The binary works on Linux, macOS, and Windows. Use `GOOS` and `GOARCH` for cross-compilation if needed (e.g., `GOOS=windows GOARCH=amd64 go build -o fileserver.exe server.go`).


## Contributing

Contributions are welcome! Please submit a pull request or open an issue on the repository for suggestions or bug reports.

## Acknowledgments

- [Tailwind CSS](https://tailwindcss.com/) for the responsive UI.
- [Google Fonts](https://fonts.google.com/) for the Roboto font.
- [pkg/browser](https://github.com/pkg/browser) for cross-platform browser opening.



