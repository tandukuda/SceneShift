package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	Theme   ThemeConfig    `yaml:"theme"`
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
	ProcessName string `yaml:"process_name"` // Now supports: "exe1.exe, exe2.exe"
	ExecPath    string `yaml:"exec_path"`
	Selected    bool   `yaml:"-"`
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
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Kill, k.Restore, k.Quit, k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Toggle},
		{k.SelectAll, k.DeselectAll},
		{k.Kill, k.Restore, k.Quit},
	}
}

// --- Model ---

type state int

const (
	stateMenu state = iota
	stateCountdown
	stateProcessing
	stateDone
)

type tickMsg time.Time

type model struct {
	config   Config
	keys     keyMap
	help     help.Model
	progress progress.Model
	appState []AppEntry

	// App Logic
	cursor       int
	mode         string
	currentState state

	// Countdown & Progress
	countdown   int
	progPercent float64
	logs        []string

	width  int
	height int
}

// --- Init ---

func loadConfig() (Config, error) {
	f, err := os.ReadFile("config.yaml")
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	err = yaml.Unmarshal(f, &cfg)
	return cfg, err
}

func initialModel() model {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config.yaml: %v\n", err)
		os.Exit(1)
	}

	// Fix space key string
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
	}

	prog := progress.New(
		progress.WithGradient(cfg.Theme.Kill, cfg.Theme.Highlight),
		progress.WithWidth(40),
	)

	return model{
		config:       cfg,
		keys:         keys,
		help:         help.New(),
		progress:     prog,
		appState:     cfg.Apps,
		currentState: stateMenu,
		countdown:    5,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

// --- Update ---

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

	case tea.KeyMsg:
		switch m.currentState {
		case stateMenu:
			// 1. Check Presets (Number Keys)
			for _, preset := range m.config.Presets {
				if msg.String() == preset.Key {
					m.applyPreset(preset)
					return m, nil
				}
			}

			// 2. Standard Keys
			switch {
			case key.Matches(msg, m.keys.Quit):
				return m, tea.Quit
			case key.Matches(msg, m.keys.Up):
				if m.cursor > 0 {
					m.cursor--
				}
			case key.Matches(msg, m.keys.Down):
				if m.cursor < len(m.appState)-1 {
					m.cursor++
				}
			case key.Matches(msg, m.keys.Toggle):
				m.appState[m.cursor].Selected = !m.appState[m.cursor].Selected
			case key.Matches(msg, m.keys.SelectAll):
				for i := range m.appState {
					m.appState[i].Selected = true
				}
			case key.Matches(msg, m.keys.DeselectAll):
				for i := range m.appState {
					m.appState[i].Selected = false
				}
			case key.Matches(msg, m.keys.Help):
				m.help.ShowAll = !m.help.ShowAll
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

		case stateDone:
			if key.Matches(msg, m.keys.Quit) {
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
			m.currentState = stateDone
			return m, cmd
		}
		return m, tea.Batch(cmd, waitForNextProcess(m, msg.index+1))
	}

	return m, nil
}

// Helper to apply presets
func (m *model) applyPreset(p PresetConfig) {
	for i := range m.appState {
		m.appState[i].Selected = false
	}
	for _, targetName := range p.Apps {
		for i, app := range m.appState {
			if strings.EqualFold(app.Name, targetName) {
				m.appState[i].Selected = true
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
		if index >= len(m.appState) {
			return processResultMsg{message: "All tasks completed.", percent: 1.0, done: true, index: index}
		}

		app := m.appState[index]
		percent := float64(index+1) / float64(len(m.appState))

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

// --- MODIFIED: killProcess handles multiple names now ---
func killProcess(rawNames string) error {
	procs, err := process.Processes()
	if err != nil {
		return err
	}

	// Split the string by commas to get all targets
	// e.g., "Creative Cloud.exe, Adobe Desktop.exe"
	targets := strings.Split(rawNames, ",")
	for i := range targets {
		targets[i] = strings.TrimSpace(targets[i])
	}

	var lastErr error
	killedCount := 0

	// Loop through every running process
	for _, p := range procs {
		n, err := p.Name()
		if err != nil {
			continue
		}

		// Check if this process matches ANY of our targets
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

	// If we killed at least one thing, we consider it a success
	if killedCount > 0 {
		return nil
	}
	return lastErr
}

// --- MODIFIED: startProcess handles fallback if exec_path is empty ---
func startProcess(rawNames, path string) error {
	var cmd *exec.Cmd
	if path != "" {
		cmd = exec.Command(path)
	} else {
		// If no path is provided, try to run the FIRST executable name in the list
		names := strings.Split(rawNames, ",")
		firstName := strings.TrimSpace(names[0])
		cmd = exec.Command(firstName)
	}
	return cmd.Start()
}

// --- View (UPDATED WITH LOGO) ---

func (m model) View() string {
	base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Text))

	// REMOVED 'title' variable since we use the ASCII logo now!

	// Presets Hint Style
	presetStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warn)).Italic(true).MarginTop(1)

	selected := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Select)).Bold(true)
	unselected := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Text)).Faint(true)
	killStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Kill)).Bold(true)
	restoreStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Restore)).Bold(true)

	var s string

	switch m.currentState {
	case stateMenu:
		// --- LOGO SECTION ---
		logoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#eb6f92")). // <--- PINK COLOR HERE
			Bold(true).
			MarginBottom(1)

		s += logoStyle.Render(logoASCII) + "\n"
		s += lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Text)).Faint(true).Render("  by tandukuda") + "\n\n"
		// --------------------

		// List Apps
		for i, app := range m.appState {
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

		// Show Available Presets
		var presetHints []string
		for _, p := range m.config.Presets {
			presetHints = append(presetHints, fmt.Sprintf("[%s] %s", p.Key, p.Name))
		}
		if len(presetHints) > 0 {
			s += presetStyle.Render("Presets: "+strings.Join(presetHints, "  ")) + "\n"
		}

		s += "\n" + m.help.View(m.keys)

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
