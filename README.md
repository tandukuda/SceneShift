<p align="center">
  <img src="assets/logo.png" width="512" alt="SceneShift Logo">
</p>

<h1 align="center">SceneShift 🎮🚀</h1>

[![Latest Release](https://img.shields.io/github/v/release/tandukuda/SceneShift?style=for-the-badge&color=blue)](https://github.com/tandukuda/SceneShift/releases/latest)
[![Go](https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Windows-blue?style=for-the-badge&logo=windows)](https://www.microsoft.com/windows)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

## What is SceneShift?

If you’re tired of manually closing dozens of apps before **gaming, rendering, or streaming**, SceneShift turns that ritual into a **single performance switch**.

**SceneShift** is a keyboard-first, terminal UI (TUI) process optimizer for Windows.  
It lets you **kill or suspend background applications** to instantly free CPU and RAM—clean, fast, and zero-bloat.

Think of it as switching your machine into a new **performance scene**.

<p align="center">
  <img src="assets/demo.gif" alt="SceneShift Demo">
</p>

---

## Why SceneShift Exists

Before SceneShift, performance prep meant:
- Opening Task Manager
- Hunting background apps
- Killing the wrong thing
- Repeating this every session

SceneShift removes that friction.

It was built to be:
- **Fast** — one keypress, one scene
- **Safe** — critical processes are protected
- **Opinionated** — curated lists, smart defaults
- **Professional** — no gimmicks, no bloat

---

## 🆕 What’s New in v2.1.1

This release focuses on **clarity, safety, and polish**, based directly on user feedback.

- 🏷️ Clearer terminology: *Safelist* → **Exclusion List**
- 📋 Expanded, curated safe‑to‑kill process lists
- 🛡️ Improved safety indicators for protected / caution processes
- 🔄 Automatic config migration from v2.1.0
- ✨ Expanded protection for Windows‑critical processes

Full changelog:  
👉 [https://github.com/tandukuda/SceneShift/releases/latest](https://github.com/tandukuda/SceneShift/releases/latest)

---

## As Seen In

SceneShift has been featured in:

* **[MajorGeeks](https://www.majorgeeks.com/files/details/sceneshift.html)** (Rated 5/5) — *"A Lean, Mean Process Killing Machine"*
* **[Neowin](https://www.neowin.net/software/sceneshift-quickly-close-multiple-apps-at-once/)** — *"Quickly close multiple apps at once"*
* **[Deskmodder](https://www.deskmodder.de/blog/2026/01/03/sceneshift-mehrere-programme-und-hintergrundanwendungen-auf-einmal-beenden/)** — *"End multiple programs and background applications at once"*
* **[YouTube](https://www.youtube.com/watch?v=VB9lv18yqAI)** — *Video Tutorial by Vasudev Menon*

---

## Quick Start (1 Minute)

### Option 1: Install via Scoop (Recommended)

```bash
scoop bucket add tandukuda https://github.com/tandukuda/scoop-bucket
scoop install sceneshift
```

### Option 2: Manual Download
1. Download the latest `.exe` from the Releases page  
2. Move it anywhere (e.g. `Documents/SceneShift`)
3. Run `SceneShift.exe`

➡️ SceneShift creates its config automatically on first run.

### Option 3: Build from Source

```bash
# Prerequisites: Go 1.21+, rsrc
go install github.com/akavel/rsrc@latest

# Clone and build
git clone https://github.com/tandukuda/sceneshift.git
cd sceneshift
build-release.bat
```

---

## Core Workflow

1. Launch SceneShift
2. Activate a preset
3. Kill or suspend background apps
4. Do your heavy work
5. Restore everything when you’re done

All without touching the mouse.

---

## Key Features

### 🔄 Kill, Suspend, Resume, Restore
- **Kill** — permanently terminate processes for max resources
- **Suspend** — freeze apps without closing them
- **Resume** — instantly unfreeze suspended apps
- **Restore** — relaunch apps after your session

---

### 🛡️ Smart Safety System

- 🛡️ **Protected** — critical Windows processes
- ✓ **Safe to Kill** — verified background apps
- ⚠ **Use Caution** — may affect system behavior

A confirmation countdown appears before any destructive action.

> Administrator permissions are required.

---

### ⚡ Preset System

Presets define **performance scenes**.
One keypress = one optimized system state.

---

## Documentation

👉 [https://tandukuda.github.io/SceneShift/](https://tandukuda.github.io/SceneShift/)

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

## 🤝 Community & Feedback

SceneShift grows through **real-world usage and feedback**.

Special thanks to contributors and users who helped shape v2.1.1:
- **[@vasudev-gm](https://github.com/vasudev-gm)** - Safe-to-kill process lists and naming improvements
- Community testers who reported edge cases and workflow issues

Have ideas or presets to share?
- Discussions: [https://github.com/tandukuda/SceneShift/discussions](https://github.com/tandukuda/SceneShift/discussions)
- Issues: [https://github.com/tandukuda/SceneShift/issues](https://github.com/tandukuda/SceneShift/issues)

---

## Support the Project

If SceneShift is part of your workflow, sponsorship helps keep it **fast, safe, and maintained**.

- ☕ Ko‑fi: [https://ko-fi.com/tandukuda](https://ko-fi.com/tandukuda)
- 💳 PayPal: [https://paypal.me/justbams](https://paypal.me/justbams)

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
<div align="center">

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) by Charm

Process lists curated by the community 

**Built with ❤️ by [tandukuda](https://github.com/tandukuda)**

[⭐ Star this repo](https://github.com/tandukuda/sceneshift) • [🐛 Report Bug](https://github.com/tandukuda/sceneshift/issues) • [💡 Request Feature](https://github.com/tandukuda/sceneshift/issues) [☕ Ko-Fi](https://ko-fi.com/tandukuda)
</div>
