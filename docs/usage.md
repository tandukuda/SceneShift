# Usage Guide

SceneShift is designed for keyboard-driven operation. All functionality can be accessed without using a mouse.

## Basic Commands

### Running SceneShift
```powershell
# Launch the application
SceneShift.exe

# Show version information
SceneShift.exe --version

# Display help message
SceneShift.exe --help
```

Administrator privileges are required for all operations.

## Navigation

### Main Menu
- Up Arrow or 'k': Move cursor up
- Down Arrow or 'j': Move cursor down
- Spacebar: Toggle selection on current item
- 'a': Select all items
- 'x': Deselect all items

### Other Screens
Most interfaces support arrow key navigation. Press Escape to return to the previous screen.

## Process Management

### Kill Mode
Permanently terminates selected processes.

1. Select applications in the main menu
2. Press 'K'
3. Review the countdown screen
4. Press 'q' to cancel or wait for execution
5. View results showing which processes were terminated

RAM reclaimed is displayed after the operation completes.

### Suspend Mode
Freezes process threads without terminating the application.

1. Select applications in the main menu
2. Press 'S'
3. Confirm during countdown
4. Processes are frozen in place

Suspended processes show a pause indicator in the status column. They consume RAM but use minimal CPU.

### Resume Mode
Unfreezes previously suspended processes.

1. Navigate to applications showing suspended status
2. Press 'U'
3. Confirm during countdown
4. Processes continue execution

Only processes suspended by SceneShift can be resumed. The application tracks specific Process IDs to ensure accuracy.

### Restore Mode
Relaunches terminated applications.

1. Select terminated applications
2. Press 'R'
3. Confirm during countdown
4. Applications start using stored executable paths

If an executable has been moved or deleted, the restore operation will fail for that application.

## Application Management

### Adding Applications
1. Press 'n' in the main menu
2. Enter application name
3. Enter process name (e.g., Discord.exe)
4. Enter full executable path
5. Press Enter to save

Alternatively, press Ctrl+F to search running processes and auto-populate the fields.

### Editing Applications
1. Navigate to the application
2. Press 'e'
3. Modify fields as needed
4. Press Enter to save

### Deleting Applications
1. Navigate to the application
2. Press 'd'
3. The application is removed immediately

## Presets

Presets allow you to define groups of applications that can be activated with a single keypress.

### Creating a Preset
1. Press 'p' to open preset manager
2. Press 'n' to create new preset
3. Enter preset name
4. Assign a single-key trigger (e.g., '1', '2', 'g')
5. Press Ctrl+F to select applications or type comma-separated names
6. Press Enter to save

### Using Presets
In the main menu, press the assigned hotkey to activate the preset. All applications in that preset are automatically selected.

### Managing Presets
- Press 'e' to edit an existing preset
- Press 'd' to delete a preset
- Press Escape to return to main menu

## Session History and Undo

### Viewing History
1. Press 'h' in the main menu
2. Review operations performed this session
3. Navigate with arrow keys
4. Press Escape to return

History includes timestamps and details about which applications were affected.

### Undoing Operations
1. Press 'u' or Ctrl+Z in the main menu
2. Review the confirmation screen
3. Press Enter to execute undo or Escape to cancel

Undo reverses the most recent operation:
- Undo Kill: Restores terminated applications
- Undo Suspend: Resumes suspended processes
- Undo Resume: Re-suspends processes
- Undo Restore: Terminates launched applications

## Configuration Profiles

### Exporting Configuration
1. Press Ctrl+E in the main menu
2. Enter a description for this configuration
3. Optionally enter your name as author
4. Press Enter
5. Profile saved as sceneshift-profile-YYYY-MM-DD.json

### Importing Configuration
1. Press 'i' in the main menu
2. Select a profile from the list using arrow keys
3. Press Enter to import

Import merges the selected profile with your current configuration. Existing apps and presets are preserved. If you want to replace your entire configuration, manually delete your config.yaml first.

## Themes

### Changing Themes
1. Press 't' in the main menu
2. Navigate through available themes
3. Theme previews in real-time
4. Press Enter to apply
5. Press Escape to return without changing

### Customizing Themes
1. In the theme selector, press 'e'
2. Modify color hex codes
3. Changes preview immediately
4. Press Enter to save as custom theme
5. Press Escape to discard changes

## Safety Features

### Exclusion List
The exclusion list contains processes that cannot be killed or suspended. This prevents accidental termination of critical Windows components.

1. Press 'w' to open exclusion list manager
2. Review protected processes
3. Type process name and press Enter to add
4. Navigate to a process and press 'd' to remove

Default exclusions include explorer.exe, dwm.exe, and other essential Windows processes.

### Safety Indicators
Each application shows a safety indicator:
- Shield icon: Protected (cannot be modified)
- Checkmark: Safe to kill
- Warning sign: Use caution

These indicators help you make informed decisions about which processes to manage.

## Process Information

The main menu displays real-time information for each application:
- Status indicator (running, suspended, or not found)
- Safety level
- CPU usage percentage
- RAM consumption in megabytes

This information updates automatically while the interface is active.

## Keyboard Reference

### Main Menu Actions
- 'K': Kill selected processes
- 'S': Suspend selected processes
- 'U': Resume suspended processes
- 'R': Restore terminated processes

### Management
- 'n': New application
- 'e': Edit application
- 'd': Delete application
- 'p': Manage presets
- 't': Change theme
- 'w': Manage exclusion list

### Advanced Features
- 'h': View session history
- 'u' or Ctrl+Z: Undo last operation
- Ctrl+E: Export configuration
- 'i': Import configuration
- '?': Toggle help display
- 'q': Quit application

### Preset Triggers
Press any configured preset hotkey (1-9, letters) to activate that preset.

## Tips and Best Practices

### Before Critical Tasks
Test your presets in low-risk situations before relying on them for important work.

### Regular Backups
Export your configuration periodically to preserve your optimized setup.

### Reviewing Changes
Use session history to verify what operations were performed if unexpected behavior occurs.

### Suspension vs Termination
Use suspend mode when you want to preserve application state. Use kill mode when you need maximum resource reclamation.

### Resource Monitoring
Check the CPU and RAM columns to identify which applications are consuming the most resources before deciding what to manage.
