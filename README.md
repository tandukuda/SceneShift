<p align="center">
  <img src="assets/logo.png" width="512" alt="SceneShift Logo">
</p>

<h1 align="center">SceneShift</h1>

[![Latest Release](https://img.shields.io/github/v/release/tandukuda/SceneShift?style=for-the-badge&color=blue)](https://github.com/tandukuda/SceneShift/releases/latest)
[![Go](https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Windows-blue?style=for-the-badge&logo=windows)](https://www.microsoft.com/windows)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

## Overview

SceneShift is a terminal-based process optimizer for Windows that helps you quickly free system resources by managing background applications. Designed for gamers, developers, and content creators who need maximum performance on demand.

If you regularly close dozens of apps before gaming, rendering, or streaming, SceneShift turns that ritual into a single keyboard command.

## What SceneShift Does

SceneShift allows you to:

- Kill or suspend background applications to free RAM and CPU
- Create presets for different performance scenarios
- Restore applications after your session
- Track session history and undo operations
- Share configurations across machines

Built for users who want a lightweight, keyboard-driven alternative to Task Manager.

<p align="center">
  <img src="assets/demo.gif" alt="SceneShift Demo">
</p>

---

## Key Features

### Process Management
- **Kill Mode**: Permanently terminate processes for maximum resource reclamation
- **Suspend Mode**: Freeze processes without closing them (preserves state)
- **Resume Mode**: Unfreeze suspended processes
- **Restore Mode**: Relaunch terminated applications

### Safety System
- **Protected Processes**: Critical Windows processes cannot be modified
- **Safe to Kill Lists**: Community-curated lists of safe-to-terminate applications
- **Safety Indicators**: Visual markers showing process risk levels

### Productivity Tools
- **Presets**: Define performance scenes with single-key triggers
- **Session History**: Review all operations performed
- **Undo Functionality**: Reverse the last operation
- **Configuration Profiles**: Export and import your entire setup

### Resource Monitoring
- **Real-time Stats**: View CPU and RAM usage per process
- **Status Indicators**: See which processes are running, suspended, or terminated
- **RAM Tracking**: Monitor memory reclaimed after optimization

## Installation

### Option 1: Scoop (Recommended)

```powershell
scoop bucket add tandukuda https://github.com/tandukuda/scoop-bucket
scoop install sceneshift
```

### Option 2: Manual Download

1. Download the latest release from the [Releases](https://github.com/tandukuda/SceneShift/releases) page
2. Extract the archive to your preferred location
3. Run SceneShift.exe as Administrator

### Option 3: Build from Source

```bash
# Prerequisites: Go 1.21+, rsrc
go install github.com/akavel/rsrc@latest

# Clone and build
git clone https://github.com/tandukuda/sceneshift.git
cd sceneshift
build-release.bat
```

## Quick Start

### First Launch

On first launch, SceneShift will:
1. Create a default configuration file
2. Ask you to choose a color theme
3. Present the main menu

Administrator privileges are required for process management.

### Basic Workflow

1. **Add Applications**: Press `n` to add apps you want to manage
2. **Select Apps**: Use arrow keys and Space to select processes
3. **Choose Action**:
   - Press `K` to kill selected processes
   - Press `S` to suspend them
   - Press `U` to resume suspended processes
   - Press `R` to restore terminated processes
4. **Confirm**: Wait through the 5-second countdown or press `q` to cancel

### Creating Presets

Presets allow you to switch your system into predefined performance states with a single keypress.

1. Press `p` to open preset manager
2. Press `n` to create a new preset
3. Enter a name (e.g., "Gaming Mode")
4. Assign a hotkey (e.g., "1")
5. Select which apps to target
6. Press Enter to save

Now press `1` in the main menu to activate Gaming Mode.

### Importing and Exporting Configurations

Share your optimized setup across machines or with the community.

**Export**:
1. Press `Ctrl+E` in the main menu
2. Enter a description and optional author name
3. Press Enter
4. Profile saved as `sceneshift-profile-YYYY-MM-DD.json`

**Import**:
1. Press `i` in the main menu
2. Select a profile from the list
3. Press Enter to merge with current configuration

## Keybindings

### Navigation
- `↑/k`: Move up
- `↓/j`: Move down
- `Space`: Toggle selection
- `a`: Select all
- `x`: Deselect all

### Actions
- `K`: Kill selected processes
- `S`: Suspend selected processes
- `U`: Resume suspended processes
- `R`: Launch/restore processes

### Management
- `n`: New app entry
- `e`: Edit selected app
- `d`: Delete selected app
- `p`: Manage presets
- `t`: Change theme
- `w`: Manage exclusion list

### Advanced
- `h`: View session history
- `u` or `Ctrl+Z`: Undo last operation
- `Ctrl+E`: Export configuration
- `i`: Import configuration
- `?`: Toggle help
- `q`: Quit

## Configuration

Configuration files are stored in the executable directory:

```
SceneShift/
├── SceneShift.exe
├── config.yaml          # Apps, presets, exclusion list, keybindings
└── theme.yaml           # Active theme colors
```

### Example config.yaml

```yaml
apps:
  - name: Discord
    process_name: Discord.exe
    exec_path: C:\Users\...\Discord.exe
    selected: false
    safety_level: safe

presets:
  - name: Gaming Mode
    key: "1"
    apps: [Discord, Chrome, Spotify]

protection:
  exclusion_list:
    - explorer.exe
    - dwm.exe
    - csrss.exe

safe_to_kill:
  bloatware:
    - OneDrive.exe
    - msedge.exe
  chat_apps:
    - Discord.exe
    - Slack.exe
  game_launchers:
    - Steam.exe
    - EpicGamesLauncher.exe
```

## Themes

Built-in themes:
- Rose Pine Moon (default)
- Dracula
- Nord
- Gruvbox Dark
- Cyberpunk

Press `t` to switch themes. Press `e` in the theme menu to create custom colors.

## Safety and Responsibility

SceneShift is a power tool. Best practices:

- Save your work before switching scenes
- Test new presets carefully
- Review the exclusion list before modifying it
- Back up your configuration before major changes

SceneShift includes safeguards for critical Windows processes, but users are responsible for their actions.

## Use Cases

### Gaming
Suspend Discord, Chrome, and game launchers to reclaim RAM before launching AAA titles. Resume everything afterward with one command.

### Development
Kill notification daemons, chat apps, and update services before compiling large projects or running resource-intensive builds.

### Content Creation
Terminate background processes before recording or streaming to prevent frame drops and ensure consistent performance.

### 3D Rendering
Free maximum RAM by suspending all non-essential applications before starting long render jobs in Blender or Maya.

## Troubleshooting

### SceneShift won't start
- Ensure you are running as Administrator
- Check that all DLL dependencies are present
- Verify Windows version is 10 or newer

### Process won't terminate
- Process may be protected (check for shield icon)
- Try suspend mode instead of kill mode
- Verify process name matches exactly

### Apps don't restore
- Check that executable paths are correct
- Ensure files haven't been moved or deleted
- Verify you have permission to launch the application

For more help, see the [Documentation](https://tandukuda.github.io/SceneShift/) or open an issue on GitHub.

## Community and Contributions

SceneShift is shaped by real-world usage and feedback. Contributions welcome:

- Share your presets and configurations
- Report bugs and edge cases
- Improve documentation
- Suggest new features

Special thanks to contributors and users who helped shape v2.2.0:
- [@vasudev-gm](https://github.com/vasudev-gm) for safe-to-kill process lists and naming improvements

## Support the Project

If SceneShift is part of your workflow, consider supporting development:

- Ko-fi: [https://ko-fi.com/tandukuda](https://ko-fi.com/tandukuda)
- PayPal: [https://paypal.me/justbams](https://paypal.me/justbams)

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and release notes.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

## Links

- [Documentation](https://tandukuda.github.io/SceneShift/)
- [Releases](https://github.com/tandukuda/SceneShift/releases)
- [Issues](https://github.com/tandukuda/SceneShift/issues)
- [Discussions](https://github.com/tandukuda/SceneShift/discussions)

---
<div align="center">

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) by Charm

Process lists curated by the community 

**Built with ❤️ by [tandukuda](https://github.com/tandukuda)**

[⭐ Star this repo](https://github.com/tandukuda/sceneshift) • [🐛 Report Bug](https://github.com/tandukuda/sceneshift/issues) • [💡 Request Feature](https://github.com/tandukuda/sceneshift/issues) • [☕ Ko-Fi](https://ko-fi.com/tandukuda)
</div>
