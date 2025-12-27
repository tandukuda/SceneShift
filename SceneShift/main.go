package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"gopkg.in/yaml.v3"
)

// --- ASCII LOGO ---
const logoASCII = `
   _____                     _____ __    _______
  / ___/________  ____  ___ / ___// /_  (_) __/ /_
  \__ \/ ___/ _ \/ __ \/ _ \\__ \/ __ \/ / /_/ __/
 ___/ / /__/  __/ / / /  __/__/ / / / / / __/ /_
/____/\___/\___/_/ /_/\___/____/_/ /_/_/_/  \__/
`

// --- Configuration ---

type Config struct {
	Theme   ThemeConfig    `yaml:"-"`
	Hotkeys HotkeyConfig   `yaml:"hotkeys"`
	Presets []PresetConfig `yaml:"presets"`
	Apps    []AppEntry     `yaml:"apps"`
}

type PresetConfig struct {
	Name string   `yaml:"name"`
	Key  string   `yaml:"key"`
	Apps []string `yaml:"apps"`
}

type ThemeConfig struct {
	Name      string `yaml:"name,omitempty"`
	Base      string `yaml:"base"`
	Surface   string `yaml:"surface"`
	Text      string `yaml:"text"`
	Highlight string `yaml:"highlight"`
	Select    string `yaml:"select"`
	Kill      string `yaml:"kill"`
	Restore   string `yaml:"restore"`
	Warn      string `yaml:"warn"`
}

type HotkeyConfig struct {
	Up          []string `yaml:"up"`
	Down        []string `yaml:"down"`
	Toggle      []string `yaml:"toggle"`
	SelectAll   []string `yaml:"select_all"`
	DeselectAll []string `yaml:"deselect_all"`
	KillMode    []string `yaml:"kill_mode"`
	RestoreMode []string `yaml:"restore_mode"`
	Quit        []string `yaml:"quit"`
	Help        []string `yaml:"help"`
}

type AppEntry struct {
	Name        string `yaml:"name"`
	ProcessName string `yaml:"process_name"`
	ExecPath    string `yaml:"exec_path"`
	Selected    bool   `yaml:"selected"`
}

// --- Hardcoded Theme Presets ---
var themePresets = []ThemeConfig{
	{Name: "Rose Pine Moon", Base: "#232136", Surface: "#2a273f", Text: "#e0def4", Highlight: "#3e8fb0", Select: "#c4a7e7", Kill: "#eb6f92", Restore: "#9ccfd8", Warn: "#f6c177"},
	{Name: "Dracula", Base: "#282a36", Surface: "#44475a", Text: "#f8f8f2", Highlight: "#bd93f9", Select: "#50fa7b", Kill: "#ff5555", Restore: "#8be9fd", Warn: "#ffb86c"},
	{Name: "Nord", Base: "#2e3440", Surface: "#3b4252", Text: "#eceff4", Highlight: "#88c0d0", Select: "#81a1c1", Kill: "#bf616a", Restore: "#a3be8c", Warn: "#ebcb8b"},
	{Name: "Gruvbox Dark", Base: "#282828", Surface: "#3c3836", Text: "#ebdbb2", Highlight: "#458588", Select: "#d79921", Kill: "#cc241d", Restore: "#98971a", Warn: "#d65d0e"},
	{Name: "Cyberpunk", Base: "#000b1e", Surface: "#05162a", Text: "#00ff9f", Highlight: "#00b8ff", Select: "#fcee0a", Kill: "#ff003c", Restore: "#00ff9f", Warn: "#fcee0a"},
}

// --- KeyMap ---

type keyMap struct {
	Up          key.Binding
	Down        key.Binding
	Toggle      key.Binding
	SelectAll   key.Binding
	DeselectAll key.Binding
	Kill        key.Binding
	Restore     key.Binding
	Quit        key.Binding
	Help        key.Binding
	NewItem     key.Binding
	EditItem    key.Binding
	DeleteItem  key.Binding
	SearchProc  key.Binding
	ThemeMenu   key.Binding
	PresetMenu  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Kill, k.Restore, k.Quit, k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Toggle},
		{k.SelectAll, k.DeselectAll},
		{k.NewItem, k.EditItem, k.DeleteItem},
		{k.Kill, k.Restore, k.Quit},
		{k.ThemeMenu, k.PresetMenu, k.SearchProc},
	}
}

// --- List Items ---

type processItem struct {
	name string
	exe  string
	path string
}

func (i processItem) Title() string       { return i.name }
func (i processItem) Description() string { return i.exe }
func (i processItem) FilterValue() string { return i.name }

type themeItem struct {
	config ThemeConfig
}

func (i themeItem) Title() string       { return i.config.Name }
func (i themeItem) Description() string { return "Press Enter to apply, 'e' to edit" }
func (i themeItem) FilterValue() string { return i.config.Name }

// --- Model ---

type state int

const (
	stateMenu state = iota
	stateCountdown
	stateProcessing
	stateDone
	stateAppEdit
	stateProcessPicker
	stateThemePicker
	stateThemeEditor
	statePresetList
	statePresetEdit
	statePresetAppPicker // New State: Picker for Preset Apps
)

type tickMsg time.Time

type model struct {
	config   Config
	keys     keyMap
	help     help.Model
	progress progress.Model

	// App Logic
	cursor        int
	mode          string
	currentState  state
	isFirstLaunch bool

	// Editor Logic
	inputs     []textinput.Model
	focusIndex int
	isNewItem  bool

	// Preset Logic
	presetCursor     int
	presetPickCursor int
	tempPresetApps   map[string]bool // Tracks selection in the picker

	// List Logic
	procList    list.Model
	allProcs    []list.Item
	searchInput textinput.Model
	themeList   list.Model

	// Countdown & Progress
	countdown   int
	progPercent float64
	logs        []string

	// Stats
	startRAM uint64
	width    int
	height   int
}

// --- Init & Config Loading ---

func loadConfig() (Config, bool, error) {
	f, err := os.ReadFile("config.yaml")
	if errors.Is(err, os.ErrNotExist) {
		cfg, _ := createDefaultConfig()
		return cfg, true, nil
	} else if err != nil {
		return Config{}, false, fmt.Errorf("could not read config.yaml: %w", err)
	}

	var cfg Config
	err = yaml.Unmarshal(f, &cfg)
	if err != nil {
		return Config{}, false, fmt.Errorf("could not parse config.yaml: %w", err)
	}
	loadTheme(&cfg)
	return cfg, false, nil
}

func createDefaultConfig() (Config, error) {
	defaultCfg := Config{
		Hotkeys: HotkeyConfig{
			Up:          []string{"up", "k"},
			Down:        []string{"down", "j"},
			Toggle:      []string{"space", " "},
			SelectAll:   []string{"a"},
			DeselectAll: []string{"x"},
			KillMode:    []string{"K"},
			RestoreMode: []string{"R"},
			Quit:        []string{"q", "ctrl+c"},
			Help:        []string{"?"},
		},
		Presets: []PresetConfig{},
		Apps:    []AppEntry{},
	}

	f, err := os.Create("config.yaml")
	if err != nil {
		return defaultCfg, nil
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	_ = encoder.Encode(defaultCfg)
	loadTheme(&defaultCfg)
	return defaultCfg, nil
}

func loadTheme(cfg *Config) {
	cfg.Theme = themePresets[0]
	fTheme, err := os.ReadFile("theme.yaml")
	if err == nil {
		_ = yaml.Unmarshal(fTheme, &cfg.Theme)
	}
}

func (m model) saveConfig() {
	f, err := os.Create("config.yaml")
	if err == nil {
		defer f.Close()
		encoder := yaml.NewEncoder(f)
		encoder.SetIndent(2)
		_ = encoder.Encode(m.config)
	}
	fT, err := os.Create("theme.yaml")
	if err == nil {
		defer fT.Close()
		encT := yaml.NewEncoder(fT)
		encT.SetIndent(2)
		_ = encT.Encode(m.config.Theme)
	}
}

func initialModel() model {
	cfg, firstLaunch, err := loadConfig()
	if err != nil {
		cfg = Config{}
		loadTheme(&cfg)
	}

	toggleKeys := cfg.Hotkeys.Toggle
	for i, k := range toggleKeys {
		if k == "space" {
			toggleKeys[i] = " "
		}
	}

	keys := keyMap{
		Up:          key.NewBinding(key.WithKeys(cfg.Hotkeys.Up...), key.WithHelp("↑/k", "up")),
		Down:        key.NewBinding(key.WithKeys(cfg.Hotkeys.Down...), key.WithHelp("↓/j", "down")),
		Toggle:      key.NewBinding(key.WithKeys(toggleKeys...), key.WithHelp("Space", "toggle")),
		SelectAll:   key.NewBinding(key.WithKeys(cfg.Hotkeys.SelectAll...), key.WithHelp("a", "all")),
		DeselectAll: key.NewBinding(key.WithKeys(cfg.Hotkeys.DeselectAll...), key.WithHelp("x", "none")),
		Kill:        key.NewBinding(key.WithKeys(cfg.Hotkeys.KillMode...), key.WithHelp("K", "KILL")),
		Restore:     key.NewBinding(key.WithKeys(cfg.Hotkeys.RestoreMode...), key.WithHelp("R", "RESTORE")),
		Quit:        key.NewBinding(key.WithKeys(cfg.Hotkeys.Quit...), key.WithHelp("q", "quit")),
		Help:        key.NewBinding(key.WithKeys(cfg.Hotkeys.Help...), key.WithHelp("?", "help")),
		NewItem:     key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
		EditItem:    key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		DeleteItem:  key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		SearchProc:  key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("ctrl+f", "search running")),
		ThemeMenu:   key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "theme")),
		PresetMenu:  key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "presets")),
	}

	prog := progress.New(
		progress.WithGradient(cfg.Theme.Kill, cfg.Theme.Highlight),
		progress.WithWidth(40),
	)

	// App Inputs Init
	appInputs := make([]textinput.Model, 3)
	for i := range appInputs {
		appInputs[i] = textinput.New()
	}

	sInput := textinput.New()
	sInput.Prompt = "🔍 Search: "
	sInput.Placeholder = "Type to filter..."
	sInput.Focus()

	lProc := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	lProc.SetShowHelp(false)
	lProc.SetShowTitle(false)
	lProc.SetFilteringEnabled(false)
	lProc.DisableQuitKeybindings()

	themeItems := []list.Item{}
	for _, t := range themePresets {
		themeItems = append(themeItems, themeItem{config: t})
	}
	lTheme := list.New(themeItems, list.NewDefaultDelegate(), 0, 0)
	lTheme.Title = "Select Theme"
	lTheme.SetShowHelp(false)

	initialState := stateMenu
	if firstLaunch {
		initialState = stateThemePicker
	}

	return model{
		config:        cfg,
		keys:          keys,
		help:          help.New(),
		progress:      prog,
		currentState:  initialState,
		isFirstLaunch: firstLaunch,
		countdown:     5,
		inputs:        appInputs,
		searchInput:   sInput,
		procList:      lProc,
		themeList:     lTheme,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func getRAMUsageMB() uint64 {
	v, err := mem.VirtualMemory()
	if err != nil {
		return 0
	}
	return v.Used / 1024 / 1024
}

// --- Process Fetching ---
func fetchRunningProcesses() []list.Item {
	procs, _ := process.Processes()
	uniqueMap := make(map[string]processItem)
	for _, p := range procs {
		name, err := p.Name()
		if err != nil || name == "" {
			continue
		}
		path, _ := p.Exe()
		if _, exists := uniqueMap[name]; !exists {
			friendlyName := strings.TrimSuffix(name, filepath.Ext(name))
			friendlyName = strings.Title(friendlyName)
			uniqueMap[name] = processItem{name: friendlyName, exe: name, path: path}
		}
	}
	items := []list.Item{}
	for _, v := range uniqueMap {
		items = append(items, v)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].(processItem).name < items[j].(processItem).name
	})
	return items
}

func filterProcs(items []list.Item, query string) []list.Item {
	if query == "" {
		return items
	}
	var matches []list.Item
	for _, item := range items {
		i := item.(processItem)
		if strings.Contains(strings.ToLower(i.name), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(i.exe), strings.ToLower(query)) {
			matches = append(matches, item)
		}
	}
	return matches
}

// --- Update ---

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.procList.SetSize(msg.Width, msg.Height-6)
		m.themeList.SetSize(msg.Width, msg.Height-4)

	case tea.KeyMsg:
		// Global Quit (Context Aware)
		safeStates := []state{stateMenu, statePresetList, stateThemePicker}
		isSafe := false
		for _, s := range safeStates {
			if m.currentState == s {
				isSafe = true
				break
			}
		}
		if isSafe && key.Matches(msg, m.keys.Quit) {
			m.saveConfig()
			return m, tea.Quit
		}

		switch m.currentState {
		case stateMenu:
			for _, preset := range m.config.Presets {
				if msg.String() == preset.Key {
					m.applyPreset(preset)
					return m, nil
				}
			}

			switch {
			case key.Matches(msg, m.keys.Up):
				if m.cursor > 0 {
					m.cursor--
				}
			case key.Matches(msg, m.keys.Down):
				if m.cursor < len(m.config.Apps)-1 {
					m.cursor++
				}
			case key.Matches(msg, m.keys.Toggle):
				if len(m.config.Apps) > 0 {
					m.config.Apps[m.cursor].Selected = !m.config.Apps[m.cursor].Selected
				}
			case key.Matches(msg, m.keys.SelectAll):
				for i := range m.config.Apps {
					m.config.Apps[i].Selected = true
				}
			case key.Matches(msg, m.keys.DeselectAll):
				for i := range m.config.Apps {
					m.config.Apps[i].Selected = false
				}
			case key.Matches(msg, m.keys.Help):
				m.help.ShowAll = !m.help.ShowAll
			case key.Matches(msg, m.keys.ThemeMenu):
				m.currentState = stateThemePicker
				return m, nil
			case key.Matches(msg, m.keys.PresetMenu):
				m.currentState = statePresetList
				m.presetCursor = 0
				return m, nil

			case key.Matches(msg, m.keys.DeleteItem):
				if len(m.config.Apps) > 0 {
					m.config.Apps = append(m.config.Apps[:m.cursor], m.config.Apps[m.cursor+1:]...)
					if m.cursor >= len(m.config.Apps) && m.cursor > 0 {
						m.cursor--
					}
					m.saveConfig()
				}
			case key.Matches(msg, m.keys.NewItem):
				m.isNewItem = true
				m.focusIndex = 0
				m.setupAppInputs()
				m.currentState = stateAppEdit
				return m, nil
			case key.Matches(msg, m.keys.EditItem):
				if len(m.config.Apps) == 0 {
					return m, nil
				}
				m.isNewItem = false
				m.focusIndex = 0
				app := m.config.Apps[m.cursor]
				m.setupAppInputs()
				m.inputs[0].SetValue(app.Name)
				m.inputs[1].SetValue(app.ProcessName)
				m.inputs[2].SetValue(app.ExecPath)
				m.currentState = stateAppEdit
				return m, nil

			case key.Matches(msg, m.keys.Kill):
				m.mode = "kill"
				m.currentState = stateCountdown
				m.countdown = 5
				return m, tickCmd()
			case key.Matches(msg, m.keys.Restore):
				m.mode = "restore"
				m.currentState = stateCountdown
				m.countdown = 5
				m.progress = progress.New(
					progress.WithGradient(m.config.Theme.Restore, m.config.Theme.Highlight),
					progress.WithWidth(40),
				)
				return m, tickCmd()
			}

		case statePresetList:
			switch {
			case key.Matches(msg, m.keys.Quit), msg.String() == "esc":
				m.currentState = stateMenu
				return m, nil
			case key.Matches(msg, m.keys.Up):
				if m.presetCursor > 0 {
					m.presetCursor--
				}
			case key.Matches(msg, m.keys.Down):
				if m.presetCursor < len(m.config.Presets)-1 {
					m.presetCursor++
				}

			case key.Matches(msg, m.keys.NewItem):
				m.isNewItem = true
				m.setupPresetInputs()
				m.currentState = statePresetEdit
				return m, nil

			case key.Matches(msg, m.keys.EditItem):
				if len(m.config.Presets) == 0 {
					return m, nil
				}
				m.isNewItem = false
				m.setupPresetInputs()
				p := m.config.Presets[m.presetCursor]
				m.inputs[0].SetValue(p.Name)
				m.inputs[1].SetValue(p.Key)
				m.inputs[2].SetValue(strings.Join(p.Apps, ", "))
				m.currentState = statePresetEdit
				return m, nil

			case key.Matches(msg, m.keys.DeleteItem):
				if len(m.config.Presets) > 0 {
					m.config.Presets = append(m.config.Presets[:m.presetCursor], m.config.Presets[m.presetCursor+1:]...)
					if m.presetCursor >= len(m.config.Presets) && m.presetCursor > 0 {
						m.presetCursor--
					}
					m.saveConfig()
				}
			}

		case statePresetEdit:
			// Ctrl+F to open App Picker
			if key.Matches(msg, m.keys.SearchProc) {
				m.currentState = statePresetAppPicker
				m.presetPickCursor = 0
				m.tempPresetApps = make(map[string]bool)

				// Load current text into map
				currentText := m.inputs[2].Value()
				parts := strings.Split(currentText, ",")
				for _, p := range parts {
					clean := strings.TrimSpace(p)
					if clean != "" {
						m.tempPresetApps[clean] = true
					}
				}
				return m, nil
			}

			switch msg.String() {
			case "enter":
				// Parse and Save
				rawApps := m.inputs[2].Value()
				splitApps := strings.Split(rawApps, ",")
				cleanApps := []string{}
				for _, a := range splitApps {
					trimmed := strings.TrimSpace(a)
					if trimmed != "" {
						cleanApps = append(cleanApps, trimmed)
					}
				}

				newPreset := PresetConfig{
					Name: m.inputs[0].Value(),
					Key:  m.inputs[1].Value(),
					Apps: cleanApps,
				}

				if m.isNewItem {
					m.config.Presets = append(m.config.Presets, newPreset)
					m.presetCursor = len(m.config.Presets) - 1
				} else {
					m.config.Presets[m.presetCursor] = newPreset
				}
				m.saveConfig()
				m.currentState = statePresetList
				return m, nil

			case "esc":
				m.currentState = statePresetList
				return m, nil

			case "tab", "shift+tab":
				if msg.String() == "tab" {
					m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
				} else {
					m.focusIndex = (m.focusIndex - 1 + len(m.inputs)) % len(m.inputs)
				}
			}

			// Update Inputs
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				if i == m.focusIndex {
					cmdFocus := m.inputs[i].Focus()
					m.inputs[i], cmd = m.inputs[i].Update(msg)
					cmds[i] = tea.Batch(cmdFocus, cmd)
				} else {
					m.inputs[i].Blur()
				}
			}
			return m, tea.Batch(cmds...)

		case statePresetAppPicker:
			// Picker Logic
			switch msg.String() {
			case "esc":
				m.currentState = statePresetEdit
				return m, nil
			case "up":
				if m.presetPickCursor > 0 {
					m.presetPickCursor--
				}
			case "down":
				if m.presetPickCursor < len(m.config.Apps)-1 {
					m.presetPickCursor++
				}
			case " ": // Space to toggle
				if len(m.config.Apps) > 0 {
					appName := m.config.Apps[m.presetPickCursor].Name
					if m.tempPresetApps[appName] {
						delete(m.tempPresetApps, appName)
					} else {
						m.tempPresetApps[appName] = true
					}
				}
			case "enter":
				// Confirm selection
				var selectedNames []string
				// Maintain original order of Config.Apps
				for _, app := range m.config.Apps {
					if m.tempPresetApps[app.Name] {
						selectedNames = append(selectedNames, app.Name)
					}
				}
				// Also add any names that were in the text box but not in Config.Apps (orphans)
				// (Optional: skipped here to keep it clean, or we can merge them)

				m.inputs[2].SetValue(strings.Join(selectedNames, ", "))
				m.currentState = statePresetEdit
				return m, nil
			}

		case stateAppEdit:
			if key.Matches(msg, m.keys.SearchProc) {
				m.currentState = stateProcessPicker
				m.allProcs = fetchRunningProcesses()
				m.searchInput.SetValue("")
				m.procList.SetItems(m.allProcs)
				m.procList.ResetSelected()
				return m, nil
			}

			switch msg.String() {
			case "enter":
				newApp := AppEntry{
					Name:        m.inputs[0].Value(),
					ProcessName: m.inputs[1].Value(),
					ExecPath:    m.inputs[2].Value(),
					Selected:    true,
				}
				if m.isNewItem {
					m.config.Apps = append(m.config.Apps, newApp)
					m.cursor = len(m.config.Apps) - 1
				} else {
					currSel := m.config.Apps[m.cursor].Selected
					newApp.Selected = currSel
					m.config.Apps[m.cursor] = newApp
				}
				m.saveConfig()
				m.currentState = stateMenu
				return m, nil

			case "esc":
				m.currentState = stateMenu
				return m, nil

			case "tab", "shift+tab":
				if msg.String() == "tab" {
					m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
				} else {
					m.focusIndex = (m.focusIndex - 1 + len(m.inputs)) % len(m.inputs)
				}
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				if i == m.focusIndex {
					cmdFocus := m.inputs[i].Focus()
					m.inputs[i], cmd = m.inputs[i].Update(msg)
					cmds[i] = tea.Batch(cmdFocus, cmd)
				} else {
					m.inputs[i].Blur()
				}
			}
			return m, tea.Batch(cmds...)

		case stateProcessPicker:
			switch msg.String() {
			case "enter":
				if m.procList.SelectedItem() != nil {
					i, ok := m.procList.SelectedItem().(processItem)
					if ok {
						m.inputs[0].SetValue(i.name)
						m.inputs[1].SetValue(i.exe)
						m.inputs[2].SetValue(i.path)
					}
					m.currentState = stateAppEdit
					return m, nil
				}
			case "esc":
				m.currentState = stateAppEdit
				return m, nil
			case "up", "down", "pgup", "pgdown":
				m.procList, cmd = m.procList.Update(msg)
				return m, cmd
			}

			var inputCmd tea.Cmd
			m.searchInput, inputCmd = m.searchInput.Update(msg)
			filtered := filterProcs(m.allProcs, m.searchInput.Value())
			m.procList.SetItems(filtered)
			return m, inputCmd

		case stateThemePicker:
			switch msg.String() {
			case "enter":
				item, ok := m.themeList.SelectedItem().(themeItem)
				if ok {
					m.config.Theme = item.config
					m.saveConfig()
				}
				m.currentState = stateMenu
				return m, nil
			case "e":
				item, ok := m.themeList.SelectedItem().(themeItem)
				if ok {
					m.config.Theme = item.config
					m.setupThemeInputs()
					m.currentState = stateThemeEditor
					m.focusIndex = 0
				}
				return m, nil
			case "esc":
				if !m.isFirstLaunch {
					m.currentState = stateMenu
				}
				return m, nil
			}
			m.themeList, cmd = m.themeList.Update(msg)
			if item, ok := m.themeList.SelectedItem().(themeItem); ok {
				m.config.Theme = item.config
			}
			return m, cmd

		case stateThemeEditor:
			switch msg.String() {
			case "enter":
				m.config.Theme.Name = "Custom"
				m.config.Theme.Base = m.inputs[0].Value()
				m.config.Theme.Surface = m.inputs[1].Value()
				m.config.Theme.Text = m.inputs[2].Value()
				m.config.Theme.Highlight = m.inputs[3].Value()
				m.config.Theme.Select = m.inputs[4].Value()
				m.config.Theme.Kill = m.inputs[5].Value()
				m.config.Theme.Restore = m.inputs[6].Value()
				m.config.Theme.Warn = m.inputs[7].Value()
				m.saveConfig()
				m.currentState = stateMenu
				return m, nil
			case "esc":
				m.currentState = stateThemePicker
				return m, nil
			case "tab", "shift+tab":
				if msg.String() == "tab" {
					m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
				} else {
					m.focusIndex = (m.focusIndex - 1 + len(m.inputs)) % len(m.inputs)
				}
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				if i == m.focusIndex {
					cmdFocus := m.inputs[i].Focus()
					m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
					cmds[i] = tea.Batch(cmdFocus, cmds[i])

					if i == 0 {
						m.config.Theme.Base = m.inputs[i].Value()
					}
					if i == 5 {
						m.config.Theme.Kill = m.inputs[i].Value()
					}
				} else {
					m.inputs[i].Blur()
				}
			}
			return m, tea.Batch(cmds...)

		case stateDone:
			if key.Matches(msg, m.keys.Quit) {
				m.saveConfig()
				return m, tea.Quit
			}
			m.currentState = stateMenu
			m.logs = []string{}
			m.progPercent = 0
		}

	case tickMsg:
		if m.currentState == stateCountdown {
			if m.countdown > 0 {
				m.countdown--
				return m, tickCmd()
			}
			m.currentState = stateProcessing
			m.startRAM = getRAMUsageMB()
			return m, processCmd(m)
		}

	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		return m, cmd

	case processResultMsg:
		m.logs = append(m.logs, msg.message)
		m.progPercent = msg.percent
		cmd := m.progress.SetPercent(msg.percent)

		if msg.done {
			if m.mode == "kill" {
				endRAM := getRAMUsageMB()
				if m.startRAM > endRAM {
					freed := m.startRAM - endRAM
					statMsg := fmt.Sprintf("🚀 RAM Reclaimed: %d MB", freed)
					m.logs = append(m.logs, statMsg)
				} else {
					m.logs = append(m.logs, "✨ Process cleanup complete.")
				}
			}
			m.currentState = stateDone
			return m, cmd
		}
		return m, tea.Batch(cmd, waitForNextProcess(m, msg.index+1))
	}

	return m, nil
}

// --- Helpers ---

func (m *model) setupAppInputs() {
	m.inputs = make([]textinput.Model, 3)
	for i := range m.inputs {
		t := textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Highlight))
		m.inputs[i] = t
	}
	m.inputs[0].Prompt = "Name: "
	m.inputs[1].Prompt = "Process: "
	m.inputs[2].Prompt = "Path: "
}

func (m *model) setupPresetInputs() {
	m.inputs = make([]textinput.Model, 3)
	for i := range m.inputs {
		t := textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Highlight))
		m.inputs[i] = t
	}
	m.inputs[0].Prompt = "Name: "
	m.inputs[1].Prompt = "Hotkey: "
	m.inputs[2].Prompt = "Apps: "
	m.inputs[2].Placeholder = "App 1, App 2 (Comma separated)"
	m.focusIndex = 0
}

func (m *model) setupThemeInputs() {
	vals := []string{
		m.config.Theme.Base, m.config.Theme.Surface, m.config.Theme.Text,
		m.config.Theme.Highlight, m.config.Theme.Select, m.config.Theme.Kill,
		m.config.Theme.Restore, m.config.Theme.Warn,
	}
	prompts := []string{"Base: ", "Surface: ", "Text: ", "High: ", "Sel: ", "Kill: ", "Rest: ", "Warn: "}

	m.inputs = make([]textinput.Model, 8)
	for i := range m.inputs {
		t := textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(vals[i]))
		t.Prompt = prompts[i]
		t.SetValue(vals[i])
		t.Placeholder = "#000000"
		m.inputs[i] = t
	}
}

func (m *model) applyPreset(p PresetConfig) {
	for i := range m.config.Apps {
		m.config.Apps[i].Selected = false
	}
	for _, targetName := range p.Apps {
		for i, app := range m.config.Apps {
			if strings.EqualFold(app.Name, targetName) {
				m.config.Apps[i].Selected = true
			}
		}
	}
}

// --- Commands ---

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type processResultMsg struct {
	message string
	percent float64
	done    bool
	index   int
}

func processCmd(m model) tea.Cmd {
	return waitForNextProcess(m, 0)
}

func waitForNextProcess(m model, index int) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(300 * time.Millisecond)
		if index >= len(m.config.Apps) {
			return processResultMsg{message: "All tasks completed.", percent: 1.0, done: true, index: index}
		}

		app := m.config.Apps[index]
		percent := float64(index+1) / float64(len(m.config.Apps))

		if !app.Selected {
			return processResultMsg{
				message: fmt.Sprintf("[SKIP] %s", app.Name),
				percent: percent,
				done:    false,
				index:   index,
			}
		}

		var msg string
		if m.mode == "kill" {
			err := killProcess(app.ProcessName)
			if err != nil {
				msg = fmt.Sprintf("[ERR]  %v", err)
			} else {
				msg = fmt.Sprintf("[KILL] Terminated %s", app.Name)
			}
		} else {
			err := startProcess(app.ProcessName, app.ExecPath)
			if err != nil {
				msg = fmt.Sprintf("[ERR]  Could not start %s", app.Name)
			} else {
				msg = fmt.Sprintf("[RUN]  Launched %s", app.Name)
			}
		}
		return processResultMsg{message: msg, percent: percent, done: false, index: index}
	}
}

func killProcess(rawNames string) error {
	procs, err := process.Processes()
	if err != nil {
		return err
	}
	targets := strings.Split(rawNames, ",")
	for i := range targets {
		targets[i] = strings.TrimSpace(targets[i])
	}
	var lastErr error
	killedCount := 0
	for _, p := range procs {
		n, err := p.Name()
		if err != nil {
			continue
		}
		for _, t := range targets {
			if strings.EqualFold(n, t) {
				if err := p.Kill(); err != nil {
					lastErr = err
				} else {
					killedCount++
				}
			}
		}
	}
	if killedCount > 0 {
		return nil
	}
	return lastErr
}

func startProcess(rawNames, path string) error {
	var cmd *exec.Cmd
	if path != "" {
		cmd = exec.Command(path)
	} else {
		names := strings.Split(rawNames, ",")
		firstName := strings.TrimSpace(names[0])
		cmd = exec.Command(firstName)
	}
	return cmd.Start()
}

// --- View ---

func (m model) View() string {
	base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Text))
	presetStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warn)).Italic(true).MarginTop(1)
	selected := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Select)).Bold(true)
	unselected := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Text)).Faint(true)
	killStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Kill)).Bold(true)
	restoreStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Restore)).Bold(true)
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Highlight)).MarginBottom(1)
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Select)).Bold(true).Underline(true)

	var s string

	switch m.currentState {
	case stateMenu:
		logoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Kill)).Bold(true).MarginBottom(1)
		s += logoStyle.Render(logoASCII) + "\n"
		s += lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Text)).Faint(true).Render("  by tandukuda") + "\n\n"

		if len(m.config.Apps) == 0 {
			s += lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warn)).Render("No apps configured. Press 'n' to add one.") + "\n"
		} else {
			for i, app := range m.config.Apps {
				cursor := "  "
				if m.cursor == i {
					cursor = "> "
				}
				check := "[ ]"
				if app.Selected {
					check = "[x]"
				}
				label := fmt.Sprintf("%s %s %s", cursor, check, app.Name)

				if m.cursor == i {
					s += selected.Render(label) + "\n"
				} else {
					s += unselected.Render(label) + "\n"
				}
			}
		}

		var presetHints []string
		for _, p := range m.config.Presets {
			presetHints = append(presetHints, fmt.Sprintf("[%s] %s", p.Key, p.Name))
		}
		if len(presetHints) > 0 {
			s += presetStyle.Render("Presets: "+strings.Join(presetHints, "  ")) + "\n"
		}
		s += "\n" + m.help.View(m.keys)

	case statePresetList:
		s += titleStyle.Render("MANAGE PRESETS") + "\n\n"
		if len(m.config.Presets) == 0 {
			s += lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warn)).Render("No presets. Press 'n' to create one.") + "\n"
		} else {
			for i, p := range m.config.Presets {
				cursor := "  "
				if m.presetCursor == i {
					cursor = "> "
				}
				label := fmt.Sprintf("%s[%s] %s (%d apps)", cursor, p.Key, p.Name, len(p.Apps))

				if m.presetCursor == i {
					s += selected.Render(label) + "\n"
				} else {
					s += unselected.Render(label) + "\n"
				}
			}
		}
		s += "\n" + lipgloss.NewStyle().Faint(true).Render("n: new, e: edit, d: delete, esc: back")

	case statePresetEdit:
		title := "EDIT PRESET"
		if m.isNewItem {
			title = "NEW PRESET"
		}
		s += titleStyle.Render(title) + "\n\n"
		s += lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warn)).Render("Press Ctrl+F to select apps from list!") + "\n\n"
		for i := range m.inputs {
			s += inputStyle.Render(m.inputs[i].View()) + "\n"
		}
		s += lipgloss.NewStyle().Faint(true).Render("\n(Tab to Move, Enter to Save, Esc to Cancel)")

	case statePresetAppPicker:
		s += titleStyle.Render("SELECT APPS FOR PRESET") + "\n\n"
		if len(m.config.Apps) == 0 {
			s += lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warn)).Render("No apps configured yet!")
		} else {
			for i, app := range m.config.Apps {
				cursor := "  "
				if m.presetPickCursor == i {
					cursor = "> "
				}
				check := "[ ]"
				if m.tempPresetApps[app.Name] {
					check = "[x]"
				}
				label := fmt.Sprintf("%s %s %s", cursor, check, app.Name)

				if m.presetPickCursor == i {
					s += selected.Render(label) + "\n"
				} else {
					s += unselected.Render(label) + "\n"
				}
			}
		}
		s += lipgloss.NewStyle().Faint(true).Render("\n(Space to Toggle, Enter to Confirm, Esc to Cancel)")

	case stateAppEdit:
		title := "EDIT APP"
		if m.isNewItem {
			title = "NEW APP"
		}
		s += titleStyle.Render(title) + "\n\n"
		s += lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warn)).Render("Press Ctrl+F to search running processes!") + "\n\n"
		for i := range m.inputs {
			s += inputStyle.Render(m.inputs[i].View()) + "\n"
		}
		s += lipgloss.NewStyle().Faint(true).Render("\n(Tab to Move, Enter to Save, Esc to Cancel)")

	case stateThemePicker:
		s += titleStyle.Render("SELECT THEME") + "\n"
		if m.isFirstLaunch {
			s += lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warn)).Render("Welcome! Please choose a theme to start.") + "\n"
		}
		s += "\n" + m.themeList.View()

	case stateThemeEditor:
		s += titleStyle.Render("EDIT THEME COLORS") + "\n\n"
		s += lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warn)).Render("Edit Hex codes. Changes apply instantly!") + "\n\n"

		mid := len(m.inputs) / 2
		col1 := ""
		col2 := ""
		for i := 0; i < mid; i++ {
			col1 += inputStyle.Render(m.inputs[i].View()) + "\n"
		}
		for i := mid; i < len(m.inputs); i++ {
			col2 += inputStyle.Render(m.inputs[i].View()) + "\n"
		}
		s += lipgloss.JoinHorizontal(lipgloss.Top, col1, "   ", col2)

		s += lipgloss.NewStyle().Faint(true).Render("\n(Enter to Save, Esc to Cancel)")

	case stateProcessPicker:
		s += lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Highlight)).Render("SEARCH RUNNING APPS") + "\n\n"
		s += m.searchInput.View() + "\n\n"
		s += m.procList.View()

	case stateCountdown:
		var modeStr string
		if m.mode == "kill" {
			modeStr = killStyle.Render("KILLING APPS")
		} else {
			modeStr = restoreStyle.Render("RESTORING APPS")
		}
		s += fmt.Sprintf("\n   %s IN...\n\n", modeStr)
		bigNum := lipgloss.NewStyle().Bold(true).Padding(1, 3).Foreground(lipgloss.Color(m.config.Theme.Warn)).Render(fmt.Sprintf("%d", m.countdown))
		s += fmt.Sprintf("      %s", bigNum)
		s += "\n\n   Press q to cancel."

	case stateProcessing, stateDone:
		var modeStr string
		if m.mode == "kill" {
			modeStr = killStyle.Render("KILLING...")
		} else {
			modeStr = restoreStyle.Render("RESTORING...")
		}
		s += modeStr + "\n\n"
		s += m.progress.View() + "\n\n"

		start := 0
		if len(m.logs) > 5 {
			start = len(m.logs) - 5
		}
		for _, log := range m.logs[start:] {
			s += base.Render(log) + "\n"
		}
		if m.currentState == stateDone {
			s += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Highlight)).Render("Done! Press any key to return.")
		}
	}

	return lipgloss.NewStyle().Padding(2, 4).Render(s)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
