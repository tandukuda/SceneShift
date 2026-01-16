<p align="center">
  <img src="assets/logo.png" width="512" alt="SceneShift Logo">
</p>

<h1 align="center">SceneShift 🎮🚀</h1>

[![Latest Release](https://img.shields.io/github/v/release/tandukuda/SceneShift?style=for-the-badge&color=blue)](https://github.com/tandukuda/SceneShift/releases/latest)
[![Go](https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Windows-blue?style=for-the-badge&logo=windows)](https://www.microsoft.com/windows)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

**SceneShift** is a terminal-first process optimizer built with Go and Bubble Tea. 

It lets you **kill or suspend background applications**, freeing CPU and RAM before gaming or rendering — clean, fast, and zero-bloat. Think of it as switching into a new “performance scene” for your machine.

<p align="center">
  <img src="assets/demo.gif" alt="SceneShift Demo">
</p>

---

## 🆕 What's New in v2.1.1

* **🏷️ Clearer Terminology**: "Safelist" → "Exclusion List" for better clarity
* **📋 Curated Safe-to-Kill Lists**: Pre-configured lists of bloatware, chat apps, and game launchers
* **🛡️ Safety Indicators**: Visual icons show which apps are protected (🛡️), safe to kill (✓), or require caution (⚠)
* **🔄 Auto-Migration**: Seamlessly upgrades v2.1.0 configs with backward compatibility
* **✨ Enhanced Protection**: Expanded list of Windows critical processes

[See full changelog →](CHANGELOG.md)

---

## 📖 Documentation

**Want to learn more?**
We have moved our detailed guides, configuration options, and advanced usage tricks to our documentation site.

👉 **[Click here to read the SceneShift Documentation](https://tandukuda.github.io/SceneShift/)**

---

## ⚡ Quick Start

You don't need to install Go or compile anything to get started.

### Option 1: via Scoop (Recommended)

The easiest way to install and stay updated.

```bash
scoop bucket add tandukuda https://github.com/tandukuda/scoop-bucket
scoop install sceneshift
```

### Option 2: Manual Download

If you prefer not to use a package manager:

1. **Download** the latest `.exe` from the **[Releases Page](https://github.com/tandukuda/SceneShift/releases)**.
2. **Move** the file to a folder (e.g., `Documents/SceneShift`).
3. **Run** `SceneShift.exe`.
   * *Note: It will request Administrator permissions to manage your processes.*

### Option 3: Build from Source

```bash
# Prerequisites: Go 1.21+, rsrc
go install github.com/akavel/rsrc@latest

# Clone and build
git clone https://github.com/tandukuda/sceneshift.git
cd sceneshift
build-release.bat
```

**Requirements:**

* Windows 10/11
* Administrator privileges (required for process management)

---

## 🎯 Key Features

### 🔄 Kill, Suspend, or Resume
- **Kill**: Permanently terminate processes to free maximum resources
- **Suspend**: Freeze processes without killing them (keep Chrome tabs alive!)
- **Resume**: Unfreeze suspended processes instantly
- **Restore**: Relaunch apps when you're done

### 🛡️ Smart Safety System
SceneShift v2.1.1 introduces intelligent safety indicators:

- **🛡️ Protected** - Critical Windows processes that cannot be touched
- **✓ Safe to Kill** - Bloatware and background apps confirmed safe to terminate
- **⚠ Use Caution** - Processes that might affect system functionality

The **Exclusion List** protects essential Windows processes like `explorer.exe`, `dwm.exe`, and `svchost.exe` from accidental termination.

### 📋 Pre-Configured Safe Lists

Out of the box, SceneShift includes curated lists of safe-to-terminate processes:

**Bloatware**
- OneDrive, Microsoft Edge, Cortana, Widget Service, etc.

**Chat Apps**
- Discord, Slack, Teams, Zoom, Skype, Telegram

**Game Launchers**
- Steam, Epic Games, Battle.net, Origin, GOG Galaxy

**Utilities**
- Spotify, iTunes, Adobe Creative Cloud, RGB software (iCUE, Razer)

### ⚡ Preset System
Create custom presets for different scenarios:
- **Gaming Mode**: Kill all launchers, chat apps, and browsers
- **Streaming Setup**: Suspend non-essential apps while keeping OBS running
- **Focus Mode**: Freeze distractions without losing your work

Activate any preset with a single keypress!

### 🎨 Customizable Themes
Choose from 5 beautiful built-in themes:
- Rose Pine Moon
- Dracula
- Nord
- Gruvbox Dark
- Cyberpunk

Or create your own custom theme with the built-in editor.

---

## ⌨️ Keybindings

### Process Management
- `K` - Kill selected processes
- `S` - Suspend selected processes  
- `U` - Resume suspended processes
- `R` - Restore/Launch processes

### Selection
- `Space` - Toggle selection
- `a` - Select all
- `x` - Deselect all
- `↑/k` - Move up
- `↓/j` - Move down

### Configuration
- `n` - New app entry
- `e` - Edit selected app
- `d` - Delete selected app
- `Ctrl+F` - Search running processes

### Menus
- `p` - Manage presets
- `t` - Change theme
- `w` - Manage exclusion list
- `?` - Toggle help
- `q` - Quit

---

## 🔧 Configuration

SceneShift uses YAML configuration files that are automatically created on first run:

### `config.yaml`
```yaml
protection:
  exclusion_list:
    - explorer.exe
    - dwm.exe
    - csrss.exe
    # ... more critical processes

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

apps:
  - name: "Discord"
    process_name: "Discord.exe"
    exec_path: "C:\\Users\\You\\AppData\\Local\\Discord\\app-1.0.0\\Discord.exe"
    safety_level: "safe"

presets:
  - name: "Gaming"
    key: "1"
    apps: ["Discord", "Steam", "Chrome"]
```

### `theme.yaml`
```yaml
name: "Rose Pine Moon"
base: "#232136"
surface: "#2a273f"
text: "#e0def4"
highlight: "#3e8fb0"
select: "#c4a7e7"
kill: "#eb6f92"
restore: "#9ccfd8"
suspend: "#f6c177"
warn: "#ea9a97"
```

---

## 🌍 As Seen In

SceneShift has been featured in:

* **[MajorGeeks](https://www.majorgeeks.com/files/details/sceneshift.html)** (Rated 5/5) — *"A Lean, Mean Process Killing Machine"*
* **[Neowin](https://www.neowin.net/software/sceneshift-quickly-close-multiple-apps-at-once/)** — *"Quickly close multiple apps at once"*
* **[Deskmodder](https://www.deskmodder.de/blog/2026/01/03/sceneshift-mehrere-programme-und-hintergrundanwendungen-auf-einmal-beenden/)** — *"End multiple programs and background applications at once"*
* **[YouTube](https://www.youtube.com/watch?v=VB9lv18yqAI)** — *Video Tutorial by Vasudev Menon*

---

## 🛡️ Safety & Disclaimer

SceneShift interacts with system processes and includes multiple safety layers:

✅ **Protected by Default**: Critical Windows processes are in the exclusion list  
✅ **Visual Indicators**: See at a glance which processes are safe to modify  
✅ **Confirmation Countdown**: 5-second warning before any action  
✅ **Suspend Instead of Kill**: Keep your work alive while freeing resources

**However**, always save your work before using process management tools. While SceneShift is designed with safety in mind, terminating the wrong process could cause data loss.

---

## 🤝 Contributing

Contributions are welcome! Special thanks to our community contributors:

- **[@vasudev-gm](https://github.com/vasudev-gm)** - Safe-to-kill process lists and naming improvements

### How to Contribute

1. Fork the repo
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add AmazingFeature'`)
4. Push to branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Development Setup

```bash
# Install dependencies
go mod download

# Build for development
build.bat

# Run
SceneShift.exe
```

---

## ☕ Support

If SceneShift helped you squeeze out those extra frames or finish a render faster, consider buying me a coffee!

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/tandukuda)

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

* Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) by Charm
* Process lists curated by the community
* Inspired by the need for keyboard-first system management
* Thanks to all contributors and users providing feedback

---

## 🗺️ Roadmap

### v2.2 (Planned)
* Visual suspension status indicators in menu
* PID-specific resume tracking
* Session history / Undo feature
* Process CPU/RAM stats in selection menu
* Export/Import configuration profiles

### v3.0 (Future)
* Linux support (SIGSTOP/SIGCONT)
* macOS support
* Auto-trigger on game launch detection
* Cloud config sync (optional)
* Process groups and dependencies

---

## 💬 Community & Support

- **[Discussions](https://github.com/tandukuda/SceneShift/discussions)** - Share ideas and ask questions
- **[Issues](https://github.com/tandukuda/SceneShift/issues)** - Report bugs or request features
- **[Documentation](https://tandukuda.github.io/SceneShift/)** - Full guides and tutorials

---

## 🎯 Use Cases

**For Gamers**: Free up RAM and CPU by killing Discord, browsers, and launchers before starting AAA games. Resume everything after with one keypress.

**For Streamers**: Suspend resource-heavy apps during streams without losing your session state.

**For Developers**: Clear background processes before compiling or running resource-intensive builds.

**For 3D Artists**: Maximum RAM availability for rendering without manually closing dozens of apps.

**For Power Users**: Full keyboard control over your system resources without touching the mouse.

---

<div align="center">

**Built with ❤️ by [tandukuda](https://github.com/tandukuda)**

[⭐ Star this repo](https://github.com/tandukuda/sceneshift) • [🐛 Report Bug](https://github.com/tandukuda/sceneshift/issues) • [💡 Request Feature](https://github.com/tandukuda/sceneshift/issues)

</div>
