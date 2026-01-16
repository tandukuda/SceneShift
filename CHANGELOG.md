# Changelog

All notable changes to SceneShift will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [2.1.1] - 2026-01-16

### 🎯 Overview
A patch release focused on clarity improvements and enhanced safety features based on community feedback. This version improves terminology, adds curated safe-to-kill process lists, and introduces visual safety indicators.

### ✨ Added
- **Safety Level Indicators**: Visual icons in the app list showing process safety status:
  - 🛡️ **Protected** - Critical system processes that cannot be modified
  - ✓ **Safe to Kill** - Community-curated safe-to-terminate processes
  - ⚠ **Use Caution** - Processes that might affect system functionality
- **Curated Safe-to-Kill Lists**: Pre-configured lists of processes safe to terminate:
  - **Bloatware**: OneDrive, Microsoft Edge, Cortana, Widget Service, etc.
  - **Chat Apps**: Discord, Slack, Teams, Zoom, Skype, Telegram
  - **Game Launchers**: Steam, Epic Games, Battle.net, Origin, GOG Galaxy
  - **Utilities**: Spotify, iTunes, Adobe Creative Cloud, RGB software
- **Auto-Detection**: Automatic safety level detection for newly added apps
- **Legend Display**: Helper text showing what each safety icon means
- **Enhanced Protection**: Expanded default exclusion list with additional critical Windows processes

### 🔄 Changed
- **BREAKING (with auto-migration)**: Renamed "Safelist" → "Exclusion List" for clarity
  - Old `safelist` field → `protection.exclusion_list`
  - UI text updated from "Safelist Manager" → "Exclusion List Manager"
  - Keybinding help text updated to match new terminology
- **Config Structure**: Reorganized configuration for better clarity:
  ```yaml
  # Old (v2.1.0)
  safelist: [...]
  
  # New (v2.1.1)
  protection:
    exclusion_list: [...]
  safe_to_kill:
    bloatware: [...]
    chat_apps: [...]
    game_launchers: [...]
    utilities: [...]
  ```
- **Enhanced Error Messages**: Protection violations now show 🛡️ icon for clarity

### 🔧 Fixed
- Improved terminology consistency across all UI elements
- Better protection messaging when attempting to modify protected processes
- Config validation to prevent empty exclusion lists

### 🔄 Migration
- **Automatic**: Existing v2.1.0 configs are automatically migrated on first run
- **Backward Compatible**: No manual intervention required
- Your existing apps and presets are preserved
- Old safelist entries are moved to `protection.exclusion_list`
- Apps without safety levels are auto-detected and assigned

### 🙏 Contributors
Special thanks to:
- **[@vasudev-gm](https://github.com/vasudev-gm)** for:
  - Identifying the confusing "Safelist" terminology
  - Providing detailed process lists from [Registry-Tweaks-Scripts](https://github.com/vasudev-gm/Registry-Tweaks-Scripts)
  - Suggesting the "Exclusion List" naming improvement
  - Contributing feedback on safety categorization

---

## [2.1.0] - 2026-01-11

### 🎯 Overview
Major feature release introducing process suspension/resume capabilities and safelist protection.

### ✨ Added
- **Suspend/Resume Feature**: Freeze processes without killing them
  - Use `S` key to suspend selected processes
  - Use `U` key to resume suspended processes
  - Maintains process state and memory while freeing CPU cycles
  - Perfect for keeping Chrome tabs alive while gaming
- **Safelist Protection**: Prevents accidental termination of critical Windows processes
  - Default safelist includes essential system processes
  - Customizable via Safelist Manager (`W` key)
  - Visual feedback when attempting to modify protected processes
- **Safelist Manager UI**: Dedicated interface for managing protected processes
  - Add processes to safelist with Enter key
  - Delete with `d` key
  - Navigate with arrow keys
- **Enhanced Theme Support**: Added Suspend color to theme configuration
- **Better Error Handling**: Improved error messages for suspend/resume operations

### 🔄 Changed
- Updated hotkey configuration to include suspend and resume modes
- Enhanced process operation flow to include suspend/resume options
- Improved countdown screen to show different colors for kill/suspend/resume
- Updated help text and keybindings display

### 🔧 Fixed
- Process detection reliability improvements
- Better handling of access-denied scenarios for system processes

---

## [2.0.0] - 2026-01-03

### 🎯 Overview
Complete rewrite with new TUI framework and major feature additions.

### ✨ Added
- **Preset System**: Create and manage app groups with hotkey triggers
  - Define presets in config (e.g., "Gaming Mode", "Work Mode")
  - Trigger with single keypress (configurable keys)
  - Visual preset hints in main menu
- **Preset App Picker**: Interactive selection UI for building presets
  - Press Ctrl+F in preset editor to open picker
  - Multi-select apps with spacebar
  - Maintains app order from config
- **Theme System**: Multiple built-in themes with custom editor
  - 5 pre-configured themes (Rose Pine Moon, Dracula, Nord, Gruvbox, Cyberpunk)
  - Live theme preview in theme picker
  - Custom theme editor with hex color support
  - Separate theme.yaml file for persistence
- **Process Search**: Real-time process discovery and filtering
  - Ctrl+F to search running processes when editing apps
  - Auto-populate app details from running processes
  - Fuzzy search filtering
- **Enhanced UI/UX**:
  - Vim-style navigation (j/k keys)
  - Full keyboard navigation
  - Context-aware help system
  - Better visual feedback and styling
  - Countdown timer before executing actions

### 🔄 Changed
- Migrated from Bubble Tea v0.x to v1.x
- Improved state management with clearer state machine
- Better config structure with nested YAML
- Enhanced error handling and user feedback

### 🔧 Fixed
- Memory leaks in process enumeration
- Config file corruption on write failures
- Unicode rendering issues in terminal

---

## [1.0.0] - 2025-12-20

### 🎯 Overview
Initial release of SceneShift.

### ✨ Added
- Basic process kill functionality
- Simple TUI with arrow key navigation
- YAML configuration support
- App entry management (add, edit, delete)
- Process launching/restore capability
- RAM usage tracking
- Progress bar for batch operations
- Basic theming support

### 🔧 Core Features
- Kill multiple processes with single action
- Restore (relaunch) terminated processes
- Multi-select with spacebar
- Select all / Deselect all
- Process name and executable path tracking
- Administrator privilege requirements
- Windows-specific process management

---

## Legend

- ✨ **Added**: New features
- 🔄 **Changed**: Changes in existing functionality
- 🔧 **Fixed**: Bug fixes
- 🗑️ **Removed**: Removed features
- 🔒 **Security**: Security improvements
- 📝 **Deprecated**: Soon-to-be removed features

---

## Links

- [Unreleased Changes](https://github.com/tandukuda/SceneShift/compare/v2.1.1...HEAD)
- [v2.1.1 Release](https://github.com/tandukuda/SceneShift/releases/tag/v2.1.1)
- [v2.1.0 Release](https://github.com/tandukuda/SceneShift/releases/tag/v2.1.0)
- [v2.0.0 Release](https://github.com/tandukuda/SceneShift/releases/tag/v2.0.0)
- [v1.0.0 Release](https://github.com/tandukuda/SceneShift/releases/tag/v1.0.0)
