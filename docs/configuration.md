# Configuration

SceneShift uses two YAML files to keep your logic and visuals separate. These are generated automatically on the first run.

## 1. `config.yaml`
This file controls **logic**: keybindings, presets, and the applications you want to manage.

```yaml
hotkeys:
  up:
    - up
    - k
  down:
    - down
    - j
  toggle:
    - ' '
  select_all:
    - a
  deselect_all:
    - x
  kill_mode:
    - K
  restore_mode:
    - R
  quit:
    - q
    - ctrl+c
  help:
    - '?'
presets: []
apps: []
```

## 2. `theme.yaml`
This file controls visuals. You can change colors here without affecting functionality.
**Default Theme: Rose Pine Moon**

```yaml
name: Rose Pine Moon
base: '#232136'
surface: '#2a273f'
text: '#e0def4'
highlight: '#3e8fb0'
select: '#c4a7e7'
kill: '#eb6f92'
restore: '#9ccfd8'
warn: '#f6c177'
```
