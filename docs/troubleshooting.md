# Troubleshooting Guide

This guide covers common issues and their solutions when using SceneShift.

## General Issues

### SceneShift Won't Start

**Symptom**: Double-clicking the executable does nothing or shows an error.

**Common Causes and Solutions**:

1. **Not running as Administrator**
   - Right-click SceneShift.exe
   - Select "Run as administrator"
   - Configure your shortcut to always run as admin (see Installation guide)

2. **Missing dependencies**
   - Install Visual C++ Redistributable (x64)
   - Download from Microsoft's website
   - Restart computer after installation

3. **Antivirus blocking execution**
   - Check your antivirus quarantine
   - Add SceneShift.exe to the exclusion list
   - Unsigned executables may trigger false positives

4. **Corrupted download**
   - Re-download the executable
   - Verify file size matches the release page
   - Extract again if using zip file

### Configuration File Errors

**Symptom**: Error message about config.yaml on startup.

**Solutions**:

1. **Corrupted configuration**
   - Delete config.yaml
   - SceneShift will regenerate it on next launch
   - Your apps and presets will be lost (back up config.yaml regularly)

2. **Invalid YAML syntax**
   - Open config.yaml in a text editor
   - Check for proper indentation (use spaces, not tabs)
   - Ensure colons have a space after them
   - Verify quotes are balanced

3. **Permission issues**
   - Ensure SceneShift folder is not in a protected location
   - Move to Documents folder if in Program Files
   - Check folder permissions allow write access

## Process Management Issues

### Process Won't Terminate

**Symptom**: Attempting to kill a process shows an error or does nothing.

**Common Causes**:

1. **Process is protected**
   - Check for shield icon next to process name
   - Protected processes cannot be terminated for safety
   - Remove from exclusion list only if you understand the risk

2. **Insufficient privileges**
   - Verify SceneShift is running as Administrator
   - Some system services require additional permissions
   - Try suspend mode instead of kill mode

3. **Process name mismatch**
   - Verify the process name exactly matches Task Manager
   - Process names are case-insensitive but must be spelled correctly
   - Some apps have multiple processes (check all variants)

4. **Protected by another application**
   - Antivirus software may protect certain processes
   - Game anti-cheat systems prevent process termination
   - Temporarily disable protection to manage these processes

### Suspend Operation Fails

**Symptom**: Attempting to suspend a process shows an error.

**Solutions**:

1. **Modern Windows app (UWP)**
   - UWP apps cannot be suspended by third-party tools
   - These apps use different process architecture
   - Use kill mode instead for UWP applications

2. **System service**
   - Windows services require special handling
   - SceneShift cannot suspend most system services
   - Check if process is in exclusion list

3. **Already suspended**
   - Check status indicator
   - If already showing suspended, no action needed
   - Use resume mode to unsuspend first

### Resume Operation Fails

**Symptom**: Resume doesn't work or resumes wrong instance.

**Common Causes**:

1. **Process no longer exists**
   - Original process may have been terminated
   - Windows may have reused the Process ID
   - SceneShift validates PID before resuming

2. **Multiple instances running**
   - SceneShift tracks specific PIDs
   - Newly launched instances won't be affected
   - This is intentional behavior to prevent mistakes

3. **Process already running**
   - Check status indicator
   - If showing running, already resumed
   - May need to suspend again first

### Restore Operation Fails

**Symptom**: Applications don't launch when restored.

**Common Causes**:

1. **Executable moved or deleted**
   - Verify the executable path in app configuration
   - Edit the app entry and update the path
   - Use Ctrl+F to search for the current location

2. **Insufficient permissions**
   - Some applications require specific user context
   - Try running SceneShift as the user who installed the app
   - Administrator privileges may not be enough for some apps

3. **Application requires specific launch arguments**
   - SceneShift launches without arguments
   - Add a launcher script if arguments are needed
   - Point the executable path to the script

4. **Application prevents multiple instances**
   - Some apps only allow one running instance
   - If already running, restore will fail
   - Check Task Manager before attempting restore

## Interface Issues

### Text Appears Garbled or Unreadable

**Symptom**: Strange characters or missing text in the interface.

**Solutions**:

1. **Terminal encoding issue**
   - Run SceneShift in Windows Terminal instead of CMD
   - Windows Terminal has better Unicode support
   - PowerShell also provides better rendering

2. **Font doesn't support icons**
   - Install a font with icon support (Cascadia Code, Fira Code)
   - Configure your terminal to use the new font
   - Restart SceneShift after font installation

3. **Console window too small**
   - Maximize the terminal window
   - SceneShift requires minimum 80x24 character display
   - Resize manually if auto-resize doesn't work

### Colors Don't Display Correctly

**Symptom**: All text appears in the same color or colors are wrong.

**Solutions**:

1. **Terminal doesn't support colors**
   - Use Windows Terminal or PowerShell
   - CMD has limited color support
   - Consider switching terminals for better experience

2. **Theme issue**
   - Press 't' to change theme
   - Try a different theme to verify
   - Check theme.yaml for corrupted values

### UI Updates Slowly or Freezes

**Symptom**: Interface is unresponsive or laggy.

**Solutions**:

1. **Too many processes tracked**
   - SceneShift updates stats for all apps
   - Reduce number of configured apps
   - Stats are cached but updates still consume resources

2. **Terminal performance**
   - Windows Terminal generally performs better than CMD
   - Try a different terminal emulator
   - Close other terminal sessions

3. **Background CPU usage**
   - Other applications may be consuming resources
   - Check Task Manager for CPU bottlenecks
   - Close unnecessary background apps

## Feature-Specific Issues

### Presets Don't Work

**Symptom**: Pressing preset hotkey does nothing or selects wrong apps.

**Solutions**:

1. **Hotkey conflict**
   - Verify hotkey doesn't conflict with existing keybindings
   - Use single character or number
   - Avoid keys used by Windows or terminal

2. **App names don't match**
   - Preset uses app names, not process names
   - Check spelling in preset configuration
   - Names are case-sensitive in config.yaml

3. **Preset not saved**
   - Verify preset appears in preset manager
   - Check config.yaml contains the preset
   - Try recreating the preset

### Import Configuration Fails

**Symptom**: Attempting to import a profile shows an error.

**Solutions**:

1. **Invalid JSON file**
   - Verify file is valid JSON
   - Use a JSON validator online
   - Check for syntax errors in the file

2. **Version incompatibility**
   - Profile may be from newer SceneShift version
   - Update SceneShift to latest version
   - Check metadata.sceneshift_version in profile

3. **File not found**
   - Ensure profile file is in SceneShift directory
   - Check filename matches exactly (case-sensitive)
   - Verify file wasn't moved or deleted

4. **No profiles in list**
   - Files must be named sceneshift-profile-*.json
   - Check file extensions are .json not .txt
   - Export a test profile to verify filename format

### Undo Doesn't Work

**Symptom**: Undo operation fails or shows error.

**Solutions**:

1. **Nothing to undo**
   - Undo only works for most recent operation
   - History is cleared when SceneShift exits
   - Verify an operation was actually performed

2. **Executable paths missing**
   - Undo kill requires valid executable paths
   - Edit app configurations to add missing paths
   - Cannot restore apps without path information

3. **Processes no longer exist**
   - Undo suspend requires original PIDs to exist
   - If processes were terminated externally, undo fails
   - This is expected behavior

### Session History Empty

**Symptom**: Pressing 'h' shows no history.

**Solutions**:

1. **No operations performed**
   - History only tracks kill/suspend/resume/restore
   - App management doesn't create history entries
   - Perform an operation to populate history

2. **History cleared on exit**
   - Session history is not persistent
   - Intentional behavior to keep memory usage low
   - Export logs if you need permanent record

## Performance Issues

### High CPU Usage

**Symptom**: SceneShift itself consumes significant CPU.

**Solutions**:

1. **Stats collection overhead**
   - Stats update in background
   - Cached for 2 seconds to minimize impact
   - Consider reducing number of tracked apps

2. **Large process list**
   - SceneShift enumerates all processes periodically
   - This is normal behavior
   - CPU usage should be < 5% when idle

### High Memory Usage

**Symptom**: SceneShift uses more RAM than expected.

**Solutions**:

1. **Large configuration**
   - Many apps and presets increase memory use
   - This is normal and generally under 50 MB
   - Not significant compared to other applications

2. **Memory leak**
   - Restart SceneShift to clear
   - Report issue on GitHub if it persists
   - Monitor memory over time to verify

## Getting Additional Help

If your issue is not covered here:

1. Check the [GitHub Issues](https://github.com/tandukuda/SceneShift/issues) page
2. Search for similar problems
3. Create a new issue with:
   - SceneShift version
   - Windows version
   - Detailed description of the problem
   - Steps to reproduce
   - Error messages or screenshots

For configuration help, see the Configuration Guide in the documentation.

For general usage questions, see the Usage Guide.
