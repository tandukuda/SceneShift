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

## ⚡ Features

### 🚀 Core Optimization
* **Smart Process Control:** Kill or suspend multiple apps with one toggle.
* **⏸️ Suspend Mode:** Pause background apps (like Spotify or Chrome) to free up resources without closing them completely.
* **📊 Snapshot RAM:** Get real-time feedback on exactly how much memory (MB/GB) was reclaimed after a kill command.

### 🎨 Customization
* **🎨 Modular Theming:** Visual settings are now separated into `theme.yaml`. Swap color schemes without breaking your app logic.
* **📁 Presets:** Switch between "Gaming", "Rendering", or custom profiles with a single key.
* **⌨️ Custom Hotkeys:** Fully remappable keybindings in `config.yaml`.

### 🛡️ System Integration
* **Admin Elevation:** Automatically requests permissions to manage system processes.
* **Windows-Native:** optimized specifically for Windows process management APIs.

---

## 📥 Quick Start

1.  **Download:** Get the latest `SceneShift.zip` from the [Releases Page](https://github.com/tandukuda/SceneShift/releases).
2.  **Install:** Extract the files (`SceneShift.exe`, `config.yaml`, `theme.yaml`) to a folder of your choice.
3.  **Run:** Double-click `SceneShift.exe`.
    * *Note: Windows will request Administrator access to manage processes.*

---

## ⚙️ Configuration

SceneShift now uses two YAML files to keep your logic and visuals separate.

### 1. `config.yaml` (Logic & Apps)
Define your hotkeys, presets, and the list of applications to manage.

```yaml
hotkeys:
  up:
    - up
    - k
  down:
    - down
    - j
  toggle:
    - ' '
    - ' '
  select_all:
    - a
  deselect_all:
    - x
  kill_mode:
    - K
  restore_mode:
    - R
  quit:
    - q
    - ctrl+c
  help:
    - '?'
presets: []
apps: []
```

### 2. `theme.yaml` (Visuals)
Customize the look and feel. Below is the default **Rose Pine Moon** configuration:
```yaml
name: Rose Pine Moon
base: '#232136'
surface: '#2a273f'
text: '#e0def4'
highlight: '#3e8fb0'
select: '#c4a7e7'
kill: '#eb6f92'
restore: '#9ccfd8'
warn: '#f6c177'
```

## 🏃 How to Use

| Action         | Input              | Description                                                      |
| -------------- | ------------------ | ---------------------------------------------------------------- |
| Navigate       | `↑` `↓` or `k` `j` | Scroll through the app list.                                     |
| Toggle         | `Space`            | Select/Deselect an app.                                          |
| Select         | `a`                | Select all visible apps.                                         |
| Deselect All   | `K`                | Clear all current selections.                                    |
| Kill Mode      | `K`                | Terminate selected apps. Now displays **"RAM Reclaimed" stats**. |
| Restore Mode   | `R`                | Relaunch apps (requires `exec_path`).                            |
| Presets        | `1`-`9`            | Apply a saved preset.                                            |
| Help           | `?`                | Toggle the help menu overlay.                                    |
| New App        | `n`                | Adding a new app.                                                |
| Edit App       | `e`                | Edit an existing app.                                            |
| Delete App     | `d`                | Delete an existing app.                                          |
| Theme Selector | `t`                | Choose or edit a theme.                                          |
| Presets Editor | `p`                | Add or edit a presets.                                           |
| Find an App    | `ctrl + f`         | Search a running apps inside Edit App.                           |

## 🛠 Build From Source
**Prerequisites:** Go 1.21+ and Windows OS.
```bash
# 1. Clone the repo
git clone https://github.com/tandukuda/SceneShift.git
cd SceneShift

# 2. Install resource tool (for icon)
go install https://github.com/akavel/rsrc@latest

# 3. Build with assets
rsrc -manifest sceneshift.manifest -ico assets/icon.ico -o sceneshift.syso
go build -ldflags "-s -w" -o SceneShift.exe
```

## ⚠️ Disclaimer
SceneShift interacts with system processes. While built with safety in mind, terminating essential system applications may cause instability. Always save your work before using Kill Mode.

## 🤝 Contributing
PRs are welcome! Check out the [Issues](https://github.com/tandukuda/SceneShift/issues) tab to see what we're working on.

## 📄 License
Distributed under the **MIT License**.
