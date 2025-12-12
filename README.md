<p align="center">
  <img src="assets/logo.png" width="512" alt="SceneShift Logo">
</p>

<h1 align="center">SceneShift 🎮🚀</h1>

![Language](https://img.shields.io/badge/language-Go-00ADD8.svg)  ![Platform](https://img.shields.io/badge/platform-Windows-blue)  ![License](https://img.shields.io/badge/license-MIT-green.svg) 

**SceneShift** is a terminal-first process optimizer built with Go and Bubble Tea.  
It lets you **kill or suspend background applications**, freeing CPU and RAM before gaming or rendering — clean, fast, and zero-bloat.

Think of it as switching into a new “performance scene” for your machine.

<p align="center">
  <img src="assets/demo.gif" alt="SceneShift Demo">
</p>

---

## Table of Contents
- [Features](#-features)
- [Quick Start (Recommended)](#-quick-start-recommended)
- [Build From Source](#-build-from-source)
- [Configuration (`configyaml`)](#️-configuration-configyaml)
- [How to Use](#-how-to-use)
- [Project Structure](#-project-structure)
- [Disclaimer](#-disclaimer)
- [Contributing](#-contributing)
- [License](#-license)

---

## ⚡ Features

- **Smart Process Control** — Kill or suspend multiple apps with one toggle  
- **Presets** — Gaming, Rendering, or your own presets via `config.yaml`  
- **Multi-Process App Handling** — Perfect for complex apps like Adobe Creative Cloud  
- **Live TUI Feedback** — Real-time logs, highlights, and progress visualization  
- **Admin Elevation** — Automatically requests permission to manage system processes  
- **Windows-Optimized** — Suspension features rely on Windows APIs  

---

## 📥 Quick Start (Recommended)

1. Go to the **Releases** page:  
   https://github.com/tandukuda/sceneshift/releases  
2. Download the latest `SceneShift.exe`  
3. Create a folder for it (e.g., Desktop/SceneShift)  
4. Place a `config.yaml` in the same folder (see example below)  
5. Double-click `SceneShift.exe` to launch  

> **Note:** Windows will prompt for Administrator access. This is required to terminate and restore processes.

---

## 🛠 Build From Source

### Prerequisites

- **Go 1.19+**
- **Windows OS**
- Admin privileges to build + run process-suspension features

### 1. Clone the Repository

```bash
git clone https://github.com/tandukuda/sceneshift.git
cd sceneshift
```

### 2. Install Icon & Manifest Tool

```bash
go install github.com/akavel/rsrc@latest
```

### 3. Build Resources + Executable

```bash
rsrc -manifest sceneshift.manifest -ico icon.ico -o sceneshift.syso
go build -o SceneShift.exe
```

<details>
<summary><b>Optional: Dependency List</b></summary>

```bash
rsrc -manifest sceneshift.manifest -ico icon.ico -o sceneshift.syso
go build -o SceneShift.exe
```
</details>

## ⚙️ Configuration (```config.yaml```)
Whether you download or build the app, you need a ```config.yaml``` file in the same folder as the ```.exe```.

<details>
<summary><b>Full Default Configuration</b></summary>

```yaml
# I'm using Rose Pine Moon for the default
theme:
  base: "#232136"
  surface: "#2a273f"
  text: "#e0def4"
  highlight: "#3e8fb0"
  select: "#c4a7e7"
  kill: "#eb6f92"
  restore: "#9ccfd8"
  warn: "#f6c177"

# Hotkey. You can change everything
hotkeys:
  up: ["up", "k"]
  down: ["down", "j"]
  toggle: ["s", " "]
  select_all: ["a"]
  deselect_all: ["x"]
  kill_mode: ["K"]
  restore_mode: ["R"]
  quit: ["q", "ctrl+c"]
  help: ["?"]

# Presets. If you want having a preset that auto-select different apps, you can add the name here
presets:
  - name: "Gaming"
    key: "1"
    apps:
      [
        "OneCommander",
        "Flow Launcher",
        "QuickLook",
        "Adobe Creative Cloud",
        "PowerToys",
        "Dropbox",
        "Google Drive",
      ]

  - name: "Rendering"
    key: "2"
    apps: ["OneCommander", "Flow Launcher", "QuickLook", "Adobe Creative Cloud"]

# The Master List (Everything)
# NOTE! Please change the exec_path to the apps, you can find the location of the apps by right-clicking on the apps > Open File Location > Open File Location (Again) > Copy as path. Make sure to add double "\\" to ensure the code runs properly
apps:
  - name: "OneCommander"
    process_name: "OneCommander.exe, OneCommanderConnector.exe"
    exec_path: "C:\\Program Files\\OneCommander\\OneCommander.exe"
  - name: "Flow Launcher"
    process_name: "Flow.Launcher.exe"
    exec_path: "C:\\Users\\user\\AppData\\Local\\FlowLauncher\\Flow.Launcher.exe"
  - name: "QuickLook"
    process_name: "QuickLook.exe"
    exec_path: "C:\\Users\\user\\AppData\\Local\\Programs\\QuickLook\\QuickLook.exe"
  - name: "Adobe Creative Cloud"
    process_name: "Creative Cloud.exe, Adobe Desktop Service.exe, CCXProcess.exe, CoreSync.exe, Creative Cloud Helper.exe, AdobeIPCBroker.exe, CCLibrary.exe, Creative Cloud UI Helper.exe, AdobeUpdateService.exe"
    exec_path: "C:\\Program Files\\Adobe\\Adobe Creative Cloud\\ACC\\Creative Cloud.exe"
  - name: "PowerToys"
    process_name: "PowerToys.exe, PowerToys.ColorPickerUI.exe"
    exec_path: "C:\\Users\\user\\AppData\\Local\\PowerToys\\PowerToys.exe"
  - name: "Dropbox"
    process_name: "Dropbox.exe"
    exec_path: "" # Leave it blank if you want to start the app manually
  - name: "Google Drive"
    process_name: "GoogleDriveFS.exe"
    exec_path: ""
```
</details>

---

## 🏃 How to Use

- **Space** → Toggle select  
- **A / X** → Select all / Deselect all  
- **K** → Kill Mode  
- **R** → Restore Mode  
- **1–9** → Apply presets  
- **Q** → Quit  
- **?** → Help  

---

## 📂 Project Structure

```
sceneshift/
├── main.go               # Core TUI logic
├── config.go             # YAML loader + validation
├── process.go            # Kill / suspend / restore
├── styles.go             # Lipgloss styling
├── sceneshift.manifest   # Elevation + DPI config
├── icon.ico
└── README.md
```

---

## 🧰 Built With

- **Go** — Core language powering SceneShift  
- **Bubble Tea** — Terminal UI framework  
- **Lipgloss** — Styling and layout  
- **Bubbles** — Progress bars, keymaps, help components  
- **gopsutil** — Process management engine  
- **YAML v3** — Configuration handling  
- **rsrc** — Icon & manifest embedding for Windows  

---

## ⚠️ Disclaimer

SceneShift interacts with system processes. Terminating essential apps may cause data loss or instability. Use responsibly and only kill processes you understand.

---

## 🤝 Contributing

PRs and issues are welcome! If you discover new presets or app lists that boost performance, share them.

---

## 📄 License

Distributed under the **MIT License**.
