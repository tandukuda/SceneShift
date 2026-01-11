# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.1.0] - 2026-01-11

### Added
- **Suspend/Resume Functionality**: Press `S` to suspend processes, `U` to resume them
  - Uses native Windows API (`NtSuspendProcess`/`NtResumeProcess`)
  - Perfect for keeping application state while freeing resources
  - Ideal for gamers who want to suspend browsers without losing tabs
- **Windows Process Safelist**: Built-in protection against killing critical system processes
  - Press `W` to manage safelist (add/remove protected processes)
  - Default list includes 20+ Windows essentials (explorer.exe, dwm.exe, services.exe, etc.)
  - Prevents accidental system crashes
  - Fully customizable
- **Version Information**: Added `--version` and `--help` CLI flags
- **Suspend mode color** to all theme presets

### Changed
- Improved error messages with specific, actionable descriptions
- Enhanced visual distinction between Kill/Suspend/Resume/Launch modes
- Better countdown and processing state handling
- Updated theme system to include suspend styling

### Fixed
- Process handling edge cases
- Config validation on load
- State management consistency

## [2.0.0] - 2025-12-27

### Added
- Initial public release
- Terminal User Interface with Bubble Tea
- Kill and Restore (Launch) process modes
- Preset system with customizable hotkeys
- 5 built-in themes (Rose Pine Moon, Dracula, Nord, Gruvbox Dark, Cyberpunk)
- YAML-based configuration
- Vim-style keybindings (j/k navigation)
- Multi-select support
- Real-time RAM usage tracking
- Process search functionality
- Theme editor and customization

[2.1.0]: https://github.com/tandukuda/sceneshift/compare/v2.0.0...v2.1.0
[2.0.0]: https://github.com/tandukuda/sceneshift/releases/tag/v2.0.0
