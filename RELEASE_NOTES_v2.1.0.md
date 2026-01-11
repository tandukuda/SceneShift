# 🎉 SceneShift v2.1.0 - Suspend/Resume & Safelist Protection

## Major Features

### 🔄 Suspend/Resume Functionality

The most-requested feature from the PRD is finally here!

**What is it?**
- Press `S` to **suspend** processes (freeze them without terminating)
- Press `U` to **resume** them later
- Processes stay in memory but use 0% CPU
- All application state is preserved (Chrome tabs, Discord servers, etc.)

**Why is this awesome?**
- **For Gamers**: Suspend your browser before gaming, resume after with all tabs intact
- **For Developers**: Pause heavy IDEs/Docker during meetings, resume instantly
- **For Multitaskers**: Freeze background apps, free up resources, restore later

**Technical Implementation:**
- Uses native Windows API (`NtSuspendProcess`/`NtResumeProcess`)
- Process threads are frozen, not terminated
- Instant freeze and resume (no restart delay)

---

### 🛡️ Windows Process Safelist

SceneShift now protects you from accidentally breaking Windows!

**What is it?**
- Built-in list of 20+ critical Windows processes
- Press `W` to view and manage your safelist
- Protected processes cannot be killed or suspended
- Fully customizable - add your own protected processes

**Protected by default:**
```
explorer.exe, dwm.exe, csrss.exe, lsass.exe, services.exe,
winlogon.exe, svchost.exe, RuntimeBroker.exe, and more...
```

**Why you need this:**
- Prevents accidentally killing Windows Explorer
- Stops you from crashing your display manager
- Protects critical system services
- Peace of mind when batch-killing processes

---

## Additional Improvements

✨ **Enhanced User Experience**
- Clear visual distinction between Kill/Suspend/Resume/Launch modes
- Each mode has its own color scheme
- Better countdown screens with mode indicators

📝 **Better Error Messages**
```
Before: [ERR] Chrome: 0x80070005
After:  [ERR] Chrome: Access Denied (run as Admin)
```

🎨 **Theme Updates**
- All 5 themes updated with Suspend mode colors
- Consistent styling across all operation modes

⚙️ **Version & Help**
```bash
SceneShift.exe --version  # Show version, build date, commit
SceneShift.exe --help     # Show usage guide
```

---

## Breaking Changes

None! All existing configs are compatible.

**Automatic upgrades:**
- Your `config.yaml` will auto-add `safelist` on first run
- New hotkeys `S` and `U` are added to your config
- All existing apps, presets, and themes are preserved

---

## Installation

### New Users

**Download `SceneShift.exe`** from Assets below → Run as Administrator

### Upgrading from v2.0

Just replace your old `SceneShift.exe` with the new one. Your config survives!

---

## Quick Start

```
1. Run SceneShift.exe
2. Select apps with Space
3. Press 'S' to suspend (NEW!)
4. Press 'U' to resume (NEW!)
5. Press 'W' to manage safelist (NEW!)
```

**Full keybindings:**
- `K` - Kill selected processes
- `S` - Suspend processes (NEW!)
- `U` - Resume suspended processes (NEW!)
- `R` - Launch processes
- `W` - Manage safelist (NEW!)

---

## Example Workflows

### 🎮 Gaming Mode
```
Before game:
  1. Select Discord, Chrome, Spotify
  2. Press 'S' to suspend
  3. Enjoy max FPS

After game:
  1. Press 'U' to resume
  2. Everything's back instantly
  3. No lost tabs or disconnected servers
```

### 💼 Focus Mode
```
1. Create "Deep Work" preset
2. Add all distractions (Slack, Discord, Email)
3. Press preset hotkey
4. Kill distractions
5. Press 'R' to restore when done
```

---

## Changelog

**Added:**
- Suspend/Resume process functionality (`S`/`U` keys)
- Windows critical process safelist (`W` key to manage)
- Version information (`--version` flag)
- Help command (`--help` flag)
- Suspend color to all theme presets

**Changed:**
- Improved error messages with actionable descriptions
- Enhanced visual styling for operation modes
- Better state management

**Fixed:**
- Edge cases in process handling
- Config validation on load

See [CHANGELOG.md](https://github.com/tandukuda/sceneshift/blob/main/CHANGELOG.md) for full details.

---

## What's Next?

**v2.2 Roadmap:**
- Visual suspension indicators in menu (see which apps are frozen)
- PID-specific resume tracking (only resume what YOU suspended)
- Session history with Undo feature
- Process resource stats (CPU/RAM) in selection menu

**v3.0 Vision:**
- Linux support (SIGSTOP/SIGCONT)
- macOS support
- Auto-trigger on game launch
- Cloud config sync

---

## Support

- 🐛 **Bug Reports**: [Open an issue](https://github.com/tandukuda/sceneshift/issues)
- 💡 **Feature Requests**: [Start a discussion](https://github.com/tandukuda/sceneshift/discussions)
- ⭐ **Enjoying SceneShift?** Star the repo!

---

## Technical Details

**Requirements:**
- Windows 10/11
- Administrator privileges

**File Size:** ~8-10 MB

**Dependencies:** None (standalone executable)

**Verification:**
```bash
# Verify download integrity
certutil -hashfile SceneShift.exe SHA256
# Compare with checksums.txt
```

---

**Built with ❤️ by tandukuda**

Special thanks to everyone who requested suspend/resume!
