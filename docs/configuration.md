# Configuration

Config files are stored in the executable directory:

```
SceneShift/
├── SceneShift.exe
├── config.yaml          # Apps, presets, safelist, keybindings
└── theme.yaml           # Active theme colors
```

### Example config.yaml

```yaml
hotkeys:
  up: [up, k]
  down: [down, j]
  kill_mode: [K]
  suspend_mode: [S]      # NEW in v2.1
  resume_mode: [U]       # NEW in v2.1
  
apps:
  - name: Discord
    process_name: Discord.exe
    exec_path: C:\Users\...\Discord.exe
    selected: false

presets:
  - name: Gaming Mode
    key: "1"
    apps: [Discord, Chrome, Spotify]

safelist:                # NEW in v2.1
  - explorer.exe
  - dwm.exe
  - csrss.exe
  # ... prevents killing these
```

---

## 🎨 Themes

Built-in themes:
- **Rose Pine Moon** (default)
- **Dracula**
- **Nord**
- **Gruvbox Dark**
- **Cyberpunk**

Press `t` to switch, `e` to edit colors in real-time.
