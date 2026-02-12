# Installation Guide

SceneShift is distributed as a standalone Windows executable. No installation wizard or registry modifications are required.

## System Requirements

- Operating System: Windows 10 or Windows 11
- Architecture: 64-bit (x64)
- Privileges: Administrator access required for process management
- Disk Space: Approximately 10 MB

## Installation Methods

### Method 1: Scoop Package Manager (Recommended)

Scoop is a command-line package manager for Windows that simplifies software installation and updates.

#### Installing Scoop
If you don't have Scoop installed, open PowerShell and run:
```powershell
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
irm get.scoop.sh | iex
```

#### Installing SceneShift via Scoop
```powershell
# Add the SceneShift bucket
scoop bucket add tandukuda https://github.com/tandukuda/scoop-bucket

# Install SceneShift
scoop install sceneshift
```

Scoop will:
- Download the latest version
- Place it in your PATH
- Create a shim for easy command-line access
- Handle updates automatically

#### Updating via Scoop
```powershell
scoop update sceneshift
```

### Method 2: Manual Installation

Manual installation gives you complete control over file placement.

#### Steps
1. Visit the [Releases page](https://github.com/tandukuda/SceneShift/releases)
2. Download the latest SceneShift-vX.X.X.zip file
3. Extract the archive to your preferred location
   - Example: `C:\Users\YourName\Documents\SceneShift`
   - Avoid placing in Program Files (requires elevated permissions for configuration files)
4. Navigate to the extracted folder
5. Right-click SceneShift.exe and select "Run as administrator"

#### Creating a Desktop Shortcut
1. Right-click SceneShift.exe
2. Select "Send to" then "Desktop (create shortcut)"
3. Right-click the desktop shortcut
4. Select "Properties"
5. Click "Advanced" button
6. Check "Run as administrator"
7. Click OK twice

Now you can launch SceneShift from the desktop with administrative privileges.

#### Adding to PATH (Optional)
To run SceneShift from any command prompt:

1. Copy the full path to the SceneShift folder
2. Open System Properties (Windows Key + Pause/Break)
3. Click "Advanced system settings"
4. Click "Environment Variables"
5. Under "User variables", select "Path" and click "Edit"
6. Click "New"
7. Paste the SceneShift folder path
8. Click OK on all dialogs

You can now run `SceneShift.exe` from any terminal.

### Method 3: Build from Source

Building from source requires development tools but ensures you have the absolute latest code.

#### Prerequisites
- Go 1.21 or newer
- Git
- rsrc tool for embedding Windows resources

#### Installing Prerequisites
```powershell
# Install Go from https://go.dev/dl/

# Install rsrc
go install github.com/akavel/rsrc@latest

# Verify installations
go version
rsrc -h
```

#### Build Steps
```bash
# Clone the repository
git clone https://github.com/tandukuda/sceneshift.git
cd sceneshift

# Run the build script
build-release.bat

# The executable will be in the release folder
```

The build script:
- Compiles resource files (icon, manifest)
- Builds the Go application
- Strips debug symbols for smaller binary size
- Embeds version information
- Creates a release package

#### Custom Build Options
For development builds:
```bash
go build -o SceneShift.exe
```

For optimized release builds:
```bash
go build -ldflags="-s -w" -trimpath -o SceneShift.exe
```

## First Run

### Initial Setup
1. Launch SceneShift.exe as Administrator
2. On first run, SceneShift will:
   - Create config.yaml in the executable directory
   - Create theme.yaml with default colors
   - Present a theme selection screen
3. Select a theme using arrow keys and press Enter
4. You will arrive at the main menu

### Configuration Files
SceneShift creates two files on first run:

**config.yaml**: Contains applications, presets, exclusion list, and keybindings
**theme.yaml**: Stores the active color theme

These files are in the same directory as SceneShift.exe. You can edit them manually or use the built-in interfaces.

### Verifying Installation
To verify SceneShift is working correctly:

```powershell
# Check version
SceneShift.exe --version

# Should display:
# SceneShift vX.X.X
# Built: YYYY-MM-DD
# Platform: Windows
```

## Troubleshooting Installation

### Error: "Windows protected your PC"
This is SmartScreen warning about an unsigned executable.

1. Click "More info"
2. Click "Run anyway"

This occurs because the executable is not code-signed. The application is safe and open-source.

### Error: "Administrator privileges required"
SceneShift needs admin access to manage processes.

1. Right-click SceneShift.exe
2. Select "Run as administrator"
3. Or follow the desktop shortcut steps above to always run as admin

### Error: "VCRUNTIME140.dll not found"
This indicates missing Visual C++ runtime libraries.

1. Download Visual C++ Redistributable from Microsoft
2. Install the x64 version
3. Restart your computer
4. Try running SceneShift again

### Scoop Installation Fails
If Scoop cannot find SceneShift:

1. Verify the bucket was added: `scoop bucket list`
2. Update Scoop: `scoop update`
3. Try installing again: `scoop install sceneshift`

### Build from Source Fails
Common build issues:

**Go version too old**: Update to Go 1.21 or newer

**rsrc not found**: Ensure GOPATH/bin is in your PATH

**Missing dependencies**: Run `go mod download` in the SceneShift directory

## Updating SceneShift

### Via Scoop
```powershell
scoop update sceneshift
```

### Manual Update
1. Download the new version from Releases
2. Close SceneShift if running
3. Replace SceneShift.exe with the new version
4. Your configuration files (config.yaml, theme.yaml) are preserved

### Checking for Updates
SceneShift does not auto-update. Check the Releases page periodically for new versions.

## Uninstallation

### Via Scoop
```powershell
scoop uninstall sceneshift
```

### Manual Uninstall
1. Close SceneShift if running
2. Delete the SceneShift folder
3. Remove desktop shortcut if created
4. Remove from PATH if added

Your configuration files are stored with the executable, so deleting the folder removes all traces of SceneShift.

## Portable Installation

SceneShift is fully portable. You can:
- Run it from a USB drive
- Copy the folder to another computer
- Move it between directories

Just ensure you run it as Administrator wherever you place it.

## Network and Firewall

SceneShift does not require network access and makes no outbound connections. It is safe to use on isolated or air-gapped systems.

Firewall configuration is not necessary.
