# Changelog

All notable changes to SceneShift will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [2.2.0] - 2026-02-13

### Overview
Major feature release adding session management, undo functionality, real-time resource monitoring, and configuration profile import/export capabilities. This version provides complete transparency into system state and enables users to share optimized configurations.

### Added
- **Visual Status Indicators**: Real-time process state tracking
  - Running indicator: Shows actively executing processes
  - Suspended indicator: Displays frozen processes
  - Not found indicator: Marks processes that are not currently running
  - Status updates automatically when processes change state

- **PID-Specific Suspend/Resume Tracking**: Precise process management
  - Records exact Process IDs when suspending applications
  - Resume operations target only the specific PIDs that were suspended
  - Prevents accidental resumption of newly launched instances
  - Validates PID existence before attempting resume operations
  - Automatic cleanup of invalid PIDs from tracking

- **Real-time Resource Statistics**: CPU and RAM monitoring per process
  - Displays current CPU usage percentage for each application
  - Shows RAM consumption in megabytes
  - Stats cached for 2 seconds to minimize performance overhead
  - Updates automatically in background without blocking UI
  - Helps identify resource-intensive processes at a glance

- **Session History**: Complete operation tracking
  - Records all kill, suspend, resume, and restore operations
  - Displays timestamp for each operation
  - Shows which applications were affected
  - Lists success and failure counts
  - Accessible via 'h' hotkey
  - History clears on application exit (session-only)

- **Undo Functionality**: Reverse recent operations
  - One-click undo for the most recent operation
  - Works for kill (restores apps), suspend (resumes apps), resume (re-suspends apps), and restore (kills apps)
  - Displays confirmation dialog before executing
  - Shows detailed results after undo completes
  - Accessible via 'u' or 'Ctrl+Z' hotkeys
  - Removes operation from history after successful undo

- **Configuration Profile Export**: Share and backup configurations
  - Export complete configuration to JSON file
  - Includes apps, presets, themes, and protection lists
  - Metadata tracking: version, export date, description, author
  - Auto-generated filenames with date stamps
  - Accessible via 'Ctrl+E' hotkey

- **Configuration Profile Import**: Load saved configurations
  - Interactive file browser shows available profiles
  - Displays profile metadata (description, date) in selection list
  - Merge mode adds to existing configuration
  - Version compatibility warnings for profiles from newer versions
  - Validates JSON structure before importing
  - Accessible via 'i' hotkey

- **Enhanced Safety Indicators**: Improved process categorization
  - Protected indicator for critical system processes
  - Safe indicator for verified safe-to-terminate processes
  - Caution indicator for processes that may affect system functionality
  - Legend displayed in main menu for reference

- **Process State Legend**: User interface improvements
  - Clear explanation of all status symbols
  - Displayed at bottom of main menu
  - Helps users understand process states at a glance

### Changed
- **User Interface Cleanup**: Reduced visual clutter
  - Removed inline help text from main menu bottom
  - Moved all shortcuts to expandable help menu (press '?')
  - Cleaner, more focused main interface
  - Professional appearance with minimal distractions

- **Process List Enhancement**: Better information display
  - Now shows safety level, status, CPU, and RAM in one view
  - Color-coded for quick visual scanning
  - Aligned columns for better readability

- **History Display**: Improved presentation
  - Newest operations shown first
  - Concise format: timestamp, operation type, app count
  - Expandable details showing all affected applications
  - Navigate with arrow keys

- **Import User Experience**: Simplified file selection
  - No longer requires typing file paths
  - Shows selectable list of available profiles
  - Displays profile descriptions and dates
  - Navigate with arrow keys, select with Enter
  - Graceful handling when no profiles are available

### Fixed
- **Undo Operation Freeze**: Resolved UI blocking issue
  - Fixed undo command returning null instead of proper message
  - Undo operations now complete and return to menu properly
  - Progress updates correctly during undo execution

- **PID Reuse Edge Cases**: Improved process tracking reliability
  - Added validation to prevent resuming wrong process instances
  - Checks process executable path before resuming
  - Handles cases where PIDs are reused by the operating system

- **Profile Import Keybinding**: Terminal compatibility improvement
  - Changed from Ctrl+I (conflicts with Tab key) to simple 'i'
  - Works reliably across all terminal emulators
  - More intuitive single-key access

- **Window Resize Handling**: Better responsive layout
  - Profile list properly resizes with window
  - All UI elements adapt to terminal size changes
  - No clipping or overflow issues

### Technical Details
- Process stats collection runs in background thread
- Stats cached for 2 seconds to reduce CPU overhead
- Session history limited to 50 entries to prevent memory bloat
- Profile JSON files use pretty-printed format for readability
- PID validation uses Windows API to check process existence
- Undo operations execute synchronously to prevent UI blocking

### Migration Notes
- Existing v2.1.1 configurations work without modification
- No manual migration required
- Profile export creates new JSON format (not compatible with older versions)
- Stats caching is automatic and requires no configuration

### Contributors
Special thanks to users and testers who provided feedback during development:
- Community members who tested PID tracking edge cases
- Users who reported undo functionality issues
- Feedback on import/export user experience

---

## [2.1.1] - 2026-01-16

### Overview
A patch release focused on clarity improvements and enhanced safety features based on community feedback. This version improves terminology, adds curated safe-to-kill process lists, and introduces visual safety indicators.

### Added
- **Safety Level Indicators**: Visual icons in the app list showing process safety status
  - Protected: Critical system processes that cannot be modified
  - Safe to Kill: Community-curated safe-to-terminate processes
  - Use Caution: Processes that might affect system functionality
  
- **Curated Safe-to-Kill Lists**: Pre-configured lists of processes safe to terminate
  - Bloatware: OneDrive, Microsoft Edge, Cortana, Widget Service, etc.
  - Chat Apps: Discord, Slack, Teams, Zoom, Skype, Telegram
  - Game Launchers: Steam, Epic Games, Battle.net, Origin, GOG Galaxy
  - Utilities: Spotify, iTunes, Adobe Creative Cloud, RGB software
  
- **Auto-Detection**: Automatic safety level detection for newly added apps
- **Legend Display**: Helper text showing what each safety icon means
- **Enhanced Protection**: Expanded default exclusion list with additional critical Windows processes

### Changed
- **BREAKING (with auto-migration)**: Renamed "Safelist" to "Exclusion List" for clarity
  - Old safelist field migrated to protection.exclusion_list
  - UI text updated from "Safelist Manager" to "Exclusion List Manager"
  - Keybinding help text updated to match new terminology
  
- **Config Structure**: Reorganized configuration for better clarity
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
  
- **Enhanced Error Messages**: Protection violations now show shield icon for clarity

### Fixed
- Improved terminology consistency across all UI elements
- Better protection messaging when attempting to modify protected processes
- Config validation to prevent empty exclusion lists

### Migration
- Automatic: Existing v2.1.0 configs are automatically migrated on first run
- Backward Compatible: No manual intervention required
- Your existing apps and presets are preserved
- Old safelist entries are moved to protection.exclusion_list
- Apps without safety levels are auto-detected and assigned

### Contributors
Special thanks to:
- [@vasudev-gm](https://github.com/vasudev-gm) for identifying confusing terminology, providing detailed process lists, suggesting naming improvements, and contributing feedback on safety categorization

---

## [2.1.0] - 2026-01-11

### Overview
Major feature release introducing process suspension/resume capabilities and exclusion list protection.

### Added
- **Suspend/Resume Feature**: Freeze processes without killing them
  - Use 'S' key to suspend selected processes
  - Use 'U' key to resume suspended processes
  - Maintains process state and memory while freeing CPU cycles
  - Perfect for keeping Chrome tabs alive while gaming
  
- **Exclusion List Protection**: Prevents accidental termination of critical Windows processes
  - Default exclusion list includes essential system processes
  - Customizable via Exclusion List Manager ('W' key)
  - Visual feedback when attempting to modify protected processes
  
- **Exclusion List Manager UI**: Dedicated interface for managing protected processes
  - Add processes to exclusion list with Enter key
  - Delete with 'd' key
  - Navigate with arrow keys
  
- **Enhanced Theme Support**: Added Suspend color to theme configuration
- **Better Error Handling**: Improved error messages for suspend/resume operations

### Changed
- Updated hotkey configuration to include suspend and resume modes
- Enhanced process operation flow to include suspend/resume options
- Improved countdown screen to show different colors for kill/suspend/resume
- Updated help text and keybindings display

### Fixed
- Process detection reliability improvements
- Better handling of access-denied scenarios for system processes

---

## [2.0.0] - 2026-01-03

### Overview
Complete rewrite with new TUI framework and major feature additions.

### Added
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

### Changed
- Migrated from Bubble Tea v0.x to v1.x
- Improved state management with clearer state machine
- Better config structure with nested YAML
- Enhanced error handling and user feedback

### Fixed
- Memory leaks in process enumeration
- Config file corruption on write failures
- Unicode rendering issues in terminal

---

## [1.0.0] - 2025-12-20

### Overview
Initial release of SceneShift.

### Added
- Basic process kill functionality
- Simple TUI with arrow key navigation
- YAML configuration support
- App entry management (add, edit, delete)
- Process launching/restore capability
- RAM usage tracking
- Progress bar for batch operations
- Basic theming support

### Core Features
- Kill multiple processes with single action
- Restore (relaunch) terminated processes
- Multi-select with spacebar
- Select all / Deselect all
- Process name and executable path tracking
- Administrator privilege requirements
- Windows-specific process management

---

## Links

- [Unreleased Changes](https://github.com/tandukuda/SceneShift/compare/v2.2.0...HEAD)
- [v2.2.0 Release](https://github.com/tandukuda/SceneShift/releases/tag/v2.2.0)
- [v2.1.1 Release](https://github.com/tandukuda/SceneShift/releases/tag/v2.1.1)
- [v2.1.0 Release](https://github.com/tandukuda/SceneShift/releases/tag/v2.1.0)
- [v2.0.0 Release](https://github.com/tandukuda/SceneShift/releases/tag/v2.0.0)
- [v1.0.0 Release](https://github.com/tandukuda/SceneShift/releases/tag/v1.0.0)
