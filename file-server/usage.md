# Usage Instructions for FileServer

This guide explains how to make the `fileserver` binary accessible from anywhere on your system.

## Steps to Make FileServer Globally Accessible

1. **Compile the Binary**:
   In the directory containing `server.go`, compile the script into a binary:
   ```bash
   go build -o fileserver server.go
   ```

2. **Move the Binary to a PATH Directory**:
   Move the `fileserver` binary to a directory in your system's PATH.

   - **Linux/macOS**:
     ```bash
     sudo mv fileserver /usr/local/bin/
     ```
     Or, use `~/bin`:
     ```bash
     mkdir -p ~/bin
     mv fileserver ~/bin/
     ```
     Ensure `~/bin` is in your PATH by adding to `~/.bashrc` or `~/.zshrc`:
     ```bash
     export PATH="$HOME/bin:$PATH"
     ```
     Apply changes:
     ```bash
     source ~/.bashrc  # or ~/.zshrc
     ```

   - **Windows**:
     Move the binary to a directory like `C:\Tools`:
     ```cmd
     move fileserver.exe C:\Tools\
     ```
     Add `C:\Tools` to your PATH:
     1. Open "Edit the system environment variables" from the Start menu.
     2. Click "Environment Variables."
     3. Edit the `Path` variable under "System Variables" or "User Variables."
     4. Add `C:\Tools` and save.

3. **Verify Accessibility**:
   Open a new terminal or command prompt and run:
   ```bash
   fileserver --help
   ```
   You should see:
   ```
   Usage of fileserver:
     -dir string
           Directory to serve files from (default ".")
     -port int
           Port to run the server on (default 8080)
   ```

4. **Run Anywhere**:
   Use the command from any directory:
   ```bash
   fileserver -dir=~/Documents -port=8080
   ```

## Notes
- Ensure the binary has executable permissions on Linux/macOS:
  ```bash
  chmod +x /usr/local/bin/fileserver
  ```
- If the port is in use, choose a different one with the `-port` flag.
- On Windows, the binary is named `fileserver.exe`.
- If you want to open the port default port `8080` for your local network `sudo ufw allow from 192.168.1.0/24 to any port 8080` for linux systems using UFW firewall.It will only allow access from the local network. Adjust the IP range as needed .
