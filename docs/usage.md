# Usage Guide

SceneShift is keyboard-centric. You can navigate the interface entirely without a mouse.

### Basic Commands

```powershell
# Run the TUI
SceneShift.exe

# Show version
SceneShift.exe --version

# Show help
SceneShift.exe --help
```

### Keybindings

#### Main Menu
| Key | Action |
|-----|--------|
| `↑`/`k` | Move up |
| `↓`/`j` | Move down |
| `Space` | Toggle selection |
| `a` | Select all |
| `x` | Deselect all |

#### Actions
| Key | Action | Description |
|-----|--------|-------------|
| `K` | **Kill** | Terminate selected processes |
| `S` | **Suspend** | Freeze processes (keeps state) 🆕 |
| `U` | **Resume** | Unfreeze suspended processes 🆕 |
| `R` | **Launch** | Start/restart applications |

#### Management
| Key | Action |
|-----|--------|
| `n` | New app entry |
| `e` | Edit selected |
| `d` | Delete selected |
| `p` | Manage presets |
| `t` | Change theme |
| `w` | Manage safelist 🆕 |
| `Ctrl+F` | Search running processes |
| `?` | Toggle help |
| `q` | Quit |

## 💡 Example Workflows

### Gaming Mode
```
1. Select: Discord, Chrome, Spotify, Steam (when not gaming)
2. Press 'S' to suspend them all
3. Play your game with max performance
4. Press 'U' after gaming to resume
5. Everything's back exactly as you left it!
```

### Work Focus
```
1. Create preset: "Deep Work" (kills Slack, Discord, Email)
2. Assign hotkey: '1'
3. Press '1' in menu → Instant focus mode
4. Press 'R' when done → Restore communication apps
```

### Before Streaming
```
1. Preset: "Stream Mode"
2. Kills: Torrent clients, auto-updaters, notification daemons
3. One hotkey press → Clean system for streaming
```

---
