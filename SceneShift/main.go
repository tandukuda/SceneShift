package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
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

var (
	Version   = "2.1.1"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

// --- Windows API for Suspend/Resume ---
var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	ntdll                = syscall.NewLazyDLL("ntdll.dll")
	procOpenProcess      = kernel32.NewProc("OpenProcess")
	procCloseHandle      = kernel32.NewProc("CloseHandle")
	procNtSuspendProcess = ntdll.NewProc("NtSuspendProcess")
	procNtResumeProcess  = ntdll.NewProc("NtResumeProcess")
)

const (
	PROCESS_SUSPEND_RESUME    = 0x0800
	PROCESS_QUERY_INFORMATION = 0x0400
)

func openProcess(pid int32) (syscall.Handle, error) {
	handle, _, err := procOpenProcess.Call(
		uintptr(PROCESS_SUSPEND_RESUME|PROCESS_QUERY_INFORMATION),
		0,
		uintptr(pid),
	)
	if handle == 0 {
		return 0, fmt.Errorf("failed to open process PID %d: %v", pid, err)
	}
	return syscall.Handle(handle), nil
}

func suspendProcess(pid int32) error {
	handle, err := openProcess(pid)
	if err != nil {
		return err
	}
	defer procCloseHandle.Call(uintptr(handle))

	ret, _, _ := procNtSuspendProcess.Call(uintptr(handle))
	if ret != 0 {
		return fmt.Errorf("NtSuspendProcess failed with status: 0x%X", ret)
	}
	return nil
}

func resumeProcess(pid int32) error {
	handle, err := openProcess(pid)
	if err != nil {
		return err
	}
	defer procCloseHandle.Call(uintptr(handle))

	ret, _, _ := procNtResumeProcess.Call(uintptr(handle))
	if ret != 0 {
		return fmt.Errorf("NtResumeProcess failed with status: 0x%X", ret)
	}
	return nil
}

// --- Windows Essential Processes Safelist ---
var defaultSafelist = []string{
	"System", "Registry", "smss.exe", "csrss.exe", "wininit.exe",
	"services.exe", "lsass.exe", "svchost.exe", "winlogon.exe",
	"dwm.exe", "explorer.exe", "sihost.exe", "taskhostw.exe",
	"RuntimeBroker.exe", "StartMenuExperienceHost.exe",
	"MsMpEng.exe", "SecurityHealthService.exe", "SgrmBroker.exe",
	"audiodg.exe", "fontdrvhost.exe", "spoolsv.exe",
	"SearchIndexer.exe", "dllhost.exe", "conhost.exe",
}

func isInSafelist(processName string, safelist []string) bool {
	lower := strings.ToLower(processName)
	for _, safe := range safelist {
		if strings.EqualFold(safe, processName) || strings.EqualFold(safe, lower) {
			return true
		}
	}
	return false
}

// --- Default Lists for v2.1.1 ---

func getDefaultProtectionList() []string {
	return []string{
		// Critical Windows Processes
		"System", "Registry", "smss.exe", "csrss.exe", "wininit.exe",
		"services.exe", "lsass.exe", "svchost.exe", "winlogon.exe",
		"dwm.exe", "explorer.exe", "sihost.exe", "taskhostw.exe",
		"RuntimeBroker.exe", "StartMenuExperienceHost.exe",
		// Security & System
		"MsMpEng.exe", "SecurityHealthService.exe", "SgrmBroker.exe",
		"audiodg.exe", "fontdrvhost.exe", "spoolsv.exe",
		"SearchIndexer.exe", "dllhost.exe", "conhost.exe",
		"ctfmon.exe", "taskmgr.exe", "SystemSettings.exe",
	}
}

func getDefaultSafeToKill() SafeToKillConfig {
	return SafeToKillConfig{
		Bloatware: []string{
			"OneDrive.exe", "OneDriveSetup.exe", "msedge.exe",
			"MicrosoftEdgeUpdate.exe", "WidgetService.exe",
			"GameBarPresenceWriter.exe", "YourPhone.exe",
			"PhoneExperienceHost.exe", "cortana.exe",
			"SearchApp.exe", "StartMenuExperienceHost.exe",
			"TextInputHost.exe",
		},
		ChatApps: []string{
			"Discord.exe", "Slack.exe", "Teams.exe", "Zoom.exe",
			"Skype.exe", "WhatsApp.exe", "Telegram.exe", "msteams.exe",
		},
		GameLaunchers: []string{
			"Steam.exe", "EpicGamesLauncher.exe", "Battle.net.exe",
			"RiotClientServices.exe", "upc.exe", "Origin.exe",
			"GalaxyClient.exe", "Rockstar Games Launcher.exe",
		},
		Utilities: []string{
			"Spotify.exe", "SpotifyWebHelper.exe", "iTunes.exe",
			"AppleMobileDeviceService.exe", "CCXProcess.exe",
			"Creative Cloud.exe", "AdobeNotificationClient.exe",
			"NVIDIA Share.exe", "nvcontainer.exe", "GeForceExperience.exe",
			"RadeonSoftware.exe", "LightingService.exe", "iCUE.exe",
			"RzSDKService.exe",
		},
	}
}

// detectSafetyLevel determines if a process is protected, safe, or caution
func detectSafetyLevel(processName string, cfg *Config) string {
	pLower := strings.ToLower(processName)

	// Check if protected (in exclusion list)
	for _, protected := range cfg.Protection.ExclusionList {
		if strings.EqualFold(protected, processName) {
			return "protected"
		}
	}

	// Check if in safe-to-kill lists
	allSafe := append(cfg.SafeToKill.Bloatware, cfg.SafeToKill.ChatApps...)
	allSafe = append(allSafe, cfg.SafeToKill.GameLaunchers...)
	allSafe = append(allSafe, cfg.SafeToKill.Utilities...)

	for _, safe := range allSafe {
		if strings.EqualFold(safe, processName) {
			return "safe"
		}
	}

	// System-critical keywords indicate caution
	criticalKeywords := []string{"system", "service", "driver", "windows", "microsoft"}
	for _, keyword := range criticalKeywords {
		if strings.Contains(pLower, keyword) {
			return "caution"
		}
	}

	return "caution" // Default to caution for unknown processes
}

// migrateOldConfig handles backward compatibility from v2.1.0 to v2.1.1
func migrateOldConfig(cfg *Config) bool {
	migrated := false

	// Migrate old "Safelist" to new "Protection.ExclusionList"
	if len(cfg.Protection.ExclusionList) == 0 {
		cfg.Protection.ExclusionList = getDefaultProtectionList()
		migrated = true
	}

	// Add safe-to-kill lists if empty
	if len(cfg.SafeToKill.Bloatware) == 0 &&
		len(cfg.SafeToKill.ChatApps) == 0 &&
		len(cfg.SafeToKill.GameLaunchers) == 0 &&
		len(cfg.SafeToKill.Utilities) == 0 {
		cfg.SafeToKill = getDefaultSafeToKill()
		migrated = true
	}

	// Auto-detect safety levels for existing apps
	for i := range cfg.Apps {
		if cfg.Apps[i].SafetyLevel == "" {
			cfg.Apps[i].SafetyLevel = detectSafetyLevel(cfg.Apps[i].ProcessName, cfg)
			migrated = true
		}
	}

	return migrated
}

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
	Theme      ThemeConfig      `yaml:"-"`
	Hotkeys    HotkeyConfig     `yaml:"hotkeys"`
	Presets    []PresetConfig   `yaml:"presets"`
	Apps       []AppEntry       `yaml:"apps"`
	Protection ProtectionConfig `yaml:"protection"`
	SafeToKill SafeToKillConfig `yaml:"safe_to_kill"`
}

type ProtectionConfig struct {
	ExclusionList []string `yaml:"exclusion_list"`
}

type SafeToKillConfig struct {
	Bloatware     []string `yaml:"bloatware"`
	ChatApps      []string `yaml:"chat_apps"`
	GameLaunchers []string `yaml:"game_launchers"`
	Utilities     []string `yaml:"utilities"`
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
	Suspend   string `yaml:"suspend"`
	Warn      string `yaml:"warn"`
}

type HotkeyConfig struct {
	Up          []string `yaml:"up"`
	Down        []string `yaml:"down"`
	Toggle      []string `yaml:"toggle"`
	SelectAll   []string `yaml:"select_all"`
	DeselectAll []string `yaml:"deselect_all"`
	KillMode    []string `yaml:"kill_mode"`
	SuspendMode []string `yaml:"suspend_mode"`
	ResumeMode  []string `yaml:"resume_mode"`
	RestoreMode []string `yaml:"restore_mode"`
	Quit        []string `yaml:"quit"`
	Help        []string `yaml:"help"`
}

type AppEntry struct {
	Name        string         `yaml:"name"`
	ProcessName string         `yaml:"process_name"`
	ExecPath    string         `yaml:"exec_path"`
	Selected    bool           `yaml:"selected"`
	SafetyLevel string         `yaml:"safety_level,omitempty"` // NEW: "protected", "safe", "caution"
	PIDs        map[int32]bool `yaml:"-"`
}

// profileItem represents a profile file in the import list
type profileItem struct {
	filename    string
	description string
	date        string
}

func (i profileItem) Title() string { return i.filename }
func (i profileItem) Description() string {
	if i.date != "" && i.description != "" {
		return fmt.Sprintf("%s - %s", i.date, i.description)
	}
	return i.description
}
func (i profileItem) FilterValue() string { return i.filename }

// ProcessStats holds resource usage information
type ProcessStats struct {
	CPUPercent float64
	RAMMB      uint64
	IsRunning  bool
}

// StatsCache caches process statistics to reduce overhead
type StatsCache struct {
	stats     map[string]ProcessStats
	timestamp map[string]time.Time
	mutex     sync.RWMutex
	ttl       time.Duration
}

// NewStatsCache creates a new stats cache with 2-second TTL
func NewStatsCache() *StatsCache {
	return &StatsCache{
		stats:     make(map[string]ProcessStats),
		timestamp: make(map[string]time.Time),
		ttl:       2 * time.Second,
	}
}

// Get retrieves stats, using cache if valid
func (sc *StatsCache) Get(processName string) ProcessStats {
	sc.mutex.RLock()
	if ts, exists := sc.timestamp[processName]; exists {
		if time.Since(ts) < sc.ttl {
			stats := sc.stats[processName]
			sc.mutex.RUnlock()
			return stats
		}
	}
	sc.mutex.RUnlock()

	// Cache miss or expired, fetch new stats
	stats := getProcessStats(processName)

	sc.mutex.Lock()
	sc.stats[processName] = stats
	sc.timestamp[processName] = time.Now()
	sc.mutex.Unlock()

	return stats
}

// OperationType represents different types of operations
type OperationType int

const (
	OpKill OperationType = iota
	OpSuspend
	OpResume
	OpRestore
)

// String returns the string representation of an operation type
func (ot OperationType) String() string {
	switch ot {
	case OpKill:
		return "KILL"
	case OpSuspend:
		return "SUSPEND"
	case OpResume:
		return "RESUME"
	case OpRestore:
		return "RESTORE"
	default:
		return "UNKNOWN"
	}
}

// AppHistoryItem stores information about an app in history
type AppHistoryItem struct {
	Name        string
	ProcessName string
	ExecPath    string
	PIDs        []int32 // For suspend operations
}

// HistoryEntry represents a single operation in history
type HistoryEntry struct {
	ID        int
	Timestamp time.Time
	Operation OperationType
	Apps      []AppHistoryItem
	Success   int // Number of successful operations
	Failed    int // Number of failed operations
}

// SessionHistory manages the history of operations
type SessionHistory struct {
	Entries []HistoryEntry
	MaxSize int
}

// ProfileMetadata stores information about an exported profile
type ProfileMetadata struct {
	Version           string    `json:"version"`
	SceneShiftVersion string    `json:"sceneshift_version"`
	ExportDate        time.Time `json:"export_date"`
	Description       string    `json:"description"`
	Author            string    `json:"author,omitempty"`
}

// ConfigProfile represents an exportable configuration
type ConfigProfile struct {
	Metadata   ProfileMetadata  `json:"metadata"`
	Apps       []AppEntry       `json:"apps"`
	Presets    []PresetConfig   `json:"presets"`
	Theme      ThemeConfig      `json:"theme"`
	Protection ProtectionConfig `json:"protection"`
	SafeToKill SafeToKillConfig `json:"safe_to_kill"`
}

// NewSessionHistory creates a new session history with default max size
func NewSessionHistory() *SessionHistory {
	return &SessionHistory{
		Entries: make([]HistoryEntry, 0),
		MaxSize: 50,
	}
}

// Add adds a new entry to history
func (sh *SessionHistory) Add(entry HistoryEntry) {
	entry.ID = len(sh.Entries)
	sh.Entries = append(sh.Entries, entry)

	// Trim if exceeds max size
	if len(sh.Entries) > sh.MaxSize {
		sh.Entries = sh.Entries[1:]
		// Renumber IDs
		for i := range sh.Entries {
			sh.Entries[i].ID = i
		}
	}
}

// GetLast returns the most recent history entry, or nil if empty
func (sh *SessionHistory) GetLast() *HistoryEntry {
	if len(sh.Entries) == 0 {
		return nil
	}
	return &sh.Entries[len(sh.Entries)-1]
}

// Clear clears all history
func (sh *SessionHistory) Clear() {
	sh.Entries = make([]HistoryEntry, 0)
}

// IsEmpty returns true if history is empty
func (sh *SessionHistory) IsEmpty() bool {
	return len(sh.Entries) == 0
}

// --- Hardcoded Theme Presets ---
var themePresets = []ThemeConfig{
	{Name: "Rose Pine Moon", Base: "#232136", Surface: "#2a273f", Text: "#e0def4", Highlight: "#3e8fb0", Select: "#c4a7e7", Kill: "#eb6f92", Restore: "#9ccfd8", Suspend: "#f6c177", Warn: "#ea9a97"},
	{Name: "Dracula", Base: "#282a36", Surface: "#44475a", Text: "#f8f8f2", Highlight: "#bd93f9", Select: "#50fa7b", Kill: "#ff5555", Restore: "#8be9fd", Suspend: "#f1fa8c", Warn: "#ffb86c"},
	{Name: "Nord", Base: "#2e3440", Surface: "#3b4252", Text: "#eceff4", Highlight: "#88c0d0", Select: "#81a1c1", Kill: "#bf616a", Restore: "#a3be8c", Suspend: "#ebcb8b", Warn: "#d08770"},
	{Name: "Gruvbox Dark", Base: "#282828", Surface: "#3c3836", Text: "#ebdbb2", Highlight: "#458588", Select: "#d79921", Kill: "#cc241d", Restore: "#98971a", Suspend: "#fabd2f", Warn: "#d65d0e"},
	{Name: "Cyberpunk", Base: "#000b1e", Surface: "#05162a", Text: "#00ff9f", Highlight: "#00b8ff", Select: "#fcee0a", Kill: "#ff003c", Restore: "#00ff9f", Suspend: "#bd00ff", Warn: "#fcee0a"},
}

// --- KeyMap ---

type keyMap struct {
	Up           key.Binding
	Down         key.Binding
	Toggle       key.Binding
	SelectAll    key.Binding
	DeselectAll  key.Binding
	Kill         key.Binding
	Suspend      key.Binding
	Resume       key.Binding
	Restore      key.Binding
	Quit         key.Binding
	Help         key.Binding
	NewItem      key.Binding
	EditItem     key.Binding
	DeleteItem   key.Binding
	SearchProc   key.Binding
	ThemeMenu    key.Binding
	PresetMenu   key.Binding
	SafelistMenu key.Binding
	History      key.Binding
	Undo         key.Binding
	Export       key.Binding
	Import       key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Kill, k.Suspend, k.Resume, k.Restore, k.Quit, k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Toggle},
		{k.SelectAll, k.DeselectAll},
		{k.NewItem, k.EditItem, k.DeleteItem},
		{k.Kill, k.Suspend, k.Resume, k.Restore},
		{k.ThemeMenu, k.PresetMenu, k.SafelistMenu},
		{k.History, k.Undo, k.Export, k.Import},
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
	statePresetAppPicker
	stateSafelistManager
	stateHistory
	stateUndoConfirm
	stateProfileExport
	stateProfileImport
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
	statsCache    *StatsCache
	history       *SessionHistory

	// Editor Logic
	inputs     []textinput.Model
	focusIndex int
	isNewItem  bool

	// Preset Logic
	presetCursor     int
	presetPickCursor int
	tempPresetApps   map[string]bool

	// Safelist Logic
	safelistCursor int
	safelistInput  textinput.Model

	historyCursor int
	undoMessage   string

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

	// Profile Management
	profileDescription string
	profileAuthor      string
	profileMessage     string
	profileList        list.Model
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

	// Migrate old v2.1.0 config to v2.1.1
	if migrateOldConfig(&cfg) {
		// Save migrated config
		fOut, err := os.Create("config.yaml")
		if err == nil {
			defer fOut.Close()
			encoder := yaml.NewEncoder(fOut)
			encoder.SetIndent(2)
			_ = encoder.Encode(cfg)
		}
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
			SuspendMode: []string{"S"},
			ResumeMode:  []string{"U"},
			RestoreMode: []string{"R"},
			Quit:        []string{"q", "ctrl+c"},
			Help:        []string{"?"},
		},
		Presets: []PresetConfig{},
		Apps:    []AppEntry{},
		Protection: ProtectionConfig{
			ExclusionList: getDefaultProtectionList(),
		},
		SafeToKill: getDefaultSafeToKill(),
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
		Up:           key.NewBinding(key.WithKeys(cfg.Hotkeys.Up...), key.WithHelp("↑/k", "up")),
		Down:         key.NewBinding(key.WithKeys(cfg.Hotkeys.Down...), key.WithHelp("↓/j", "down")),
		Toggle:       key.NewBinding(key.WithKeys(toggleKeys...), key.WithHelp("Space", "toggle")),
		SelectAll:    key.NewBinding(key.WithKeys(cfg.Hotkeys.SelectAll...), key.WithHelp("a", "all")),
		DeselectAll:  key.NewBinding(key.WithKeys(cfg.Hotkeys.DeselectAll...), key.WithHelp("x", "none")),
		Kill:         key.NewBinding(key.WithKeys(cfg.Hotkeys.KillMode...), key.WithHelp("K", "KILL")),
		Restore:      key.NewBinding(key.WithKeys(cfg.Hotkeys.RestoreMode...), key.WithHelp("R", "RESTORE")),
		Quit:         key.NewBinding(key.WithKeys(cfg.Hotkeys.Quit...), key.WithHelp("q", "quit")),
		Help:         key.NewBinding(key.WithKeys(cfg.Hotkeys.Help...), key.WithHelp("?", "help")),
		NewItem:      key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
		EditItem:     key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		DeleteItem:   key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		SearchProc:   key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("ctrl+f", "search running")),
		ThemeMenu:    key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "theme")),
		PresetMenu:   key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "presets")),
		Suspend:      key.NewBinding(key.WithKeys(cfg.Hotkeys.SuspendMode...), key.WithHelp("S", "SUSPEND")),
		Resume:       key.NewBinding(key.WithKeys(cfg.Hotkeys.ResumeMode...), key.WithHelp("U", "RESUME")),
		SafelistMenu: key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "exclusion list")),
		History:      key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "history")),
		Undo:         key.NewBinding(key.WithKeys("u", "ctrl+z"), key.WithHelp("u/Ctrl+Z", "undo")),
		Export:       key.NewBinding(key.WithKeys("ctrl+e"), key.WithHelp("Ctrl+E", "export")),
		Import:       key.NewBinding(key.WithKeys("i"), key.WithHelp("i", "import")),
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

	safeInput := textinput.New()
	safeInput.Prompt = "➕ Add: "
	safeInput.Placeholder = "processname.exe"

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

	lProfile := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	lProfile.Title = "Select Profile"
	lProfile.SetShowHelp(false)
	lProfile.SetFilteringEnabled(false)
	lProfile.DisableQuitKeybindings()

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
		safelistInput: safeInput,
		statsCache:    NewStatsCache(),
		history:       NewSessionHistory(),
		profileList:   lProfile,
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

// getProcessStats retrieves CPU and RAM stats for a process
func getProcessStats(processName string) ProcessStats {
	stats := ProcessStats{}

	procs, err := process.Processes()
	if err != nil {
		return stats
	}

	targets := strings.Split(processName, ",")
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}

		for _, target := range targets {
			target = strings.TrimSpace(target)
			if strings.EqualFold(name, target) {
				stats.IsRunning = true

				// Get CPU usage
				if cpu, err := p.CPUPercent(); err == nil {
					stats.CPUPercent += cpu
				}

				// Get memory usage
				if mem, err := p.MemoryInfo(); err == nil {
					stats.RAMMB += mem.RSS / 1024 / 1024
				}
			}
		}
	}

	return stats
}

// pidExists checks if a PID is currently active
func pidExists(pid int32) bool {
	p, err := process.NewProcess(pid)
	if err != nil {
		return false
	}

	// Verify process is still alive
	running, err := p.IsRunning()
	return err == nil && running
}

// getProcessStatus determines the current state of a process
// Returns: "running", "suspended", "not_found"
func getProcessStatus(app AppEntry) string {
	// If we have PIDs recorded, the process is suspended
	if len(app.PIDs) > 0 {
		// Verify at least one PID still exists
		for pid := range app.PIDs {
			if pidExists(pid) {
				return "suspended"
			}
		}
		// All PIDs are gone
		return "not_found"
	}

	// Check if process is currently running
	procs, err := process.Processes()
	if err != nil {
		return "not_found"
	}

	targets := strings.Split(app.ProcessName, ",")
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		for _, target := range targets {
			target = strings.TrimSpace(target)
			if strings.EqualFold(name, target) {
				return "running"
			}
		}
	}

	return "not_found"
}

// getStatusIcon returns the appropriate icon for process status
func getStatusIcon(status string) string {
	switch status {
	case "running":
		return "▶️"
	case "suspended":
		return "⏸️"
	case "not_found":
		return "⚠️"
	default:
		return "  "
	}
}

// recordKillOperation records a kill operation in history
func (m *model) recordKillOperation(apps []AppEntry, successCount, failCount int) {
	items := make([]AppHistoryItem, 0, len(apps))
	for _, app := range apps {
		items = append(items, AppHistoryItem{
			Name:        app.Name,
			ProcessName: app.ProcessName,
			ExecPath:    app.ExecPath,
		})
	}

	m.history.Add(HistoryEntry{
		Timestamp: time.Now(),
		Operation: OpKill,
		Apps:      items,
		Success:   successCount,
		Failed:    failCount,
	})
}

// recordSuspendOperation records a suspend operation in history
func (m *model) recordSuspendOperation(apps []AppEntry, successCount, failCount int) {
	items := make([]AppHistoryItem, 0, len(apps))
	for _, app := range apps {
		// Store current PIDs for undo
		pids := make([]int32, 0, len(app.PIDs))
		for pid := range app.PIDs {
			pids = append(pids, pid)
		}

		items = append(items, AppHistoryItem{
			Name:        app.Name,
			ProcessName: app.ProcessName,
			ExecPath:    app.ExecPath,
			PIDs:        pids,
		})
	}

	m.history.Add(HistoryEntry{
		Timestamp: time.Now(),
		Operation: OpSuspend,
		Apps:      items,
		Success:   successCount,
		Failed:    failCount,
	})
}

// recordResumeOperation records a resume operation in history
func (m *model) recordResumeOperation(apps []AppEntry, successCount, failCount int) {
	items := make([]AppHistoryItem, 0, len(apps))
	for _, app := range apps {
		// Store PIDs that were resumed
		pids := make([]int32, 0, len(app.PIDs))
		for pid := range app.PIDs {
			pids = append(pids, pid)
		}

		items = append(items, AppHistoryItem{
			Name:        app.Name,
			ProcessName: app.ProcessName,
			ExecPath:    app.ExecPath,
			PIDs:        pids,
		})
	}

	m.history.Add(HistoryEntry{
		Timestamp: time.Now(),
		Operation: OpResume,
		Apps:      items,
		Success:   successCount,
		Failed:    failCount,
	})
}

// recordRestoreOperation records a restore operation in history
func (m *model) recordRestoreOperation(apps []AppEntry, successCount, failCount int) {
	items := make([]AppHistoryItem, 0, len(apps))
	for _, app := range apps {
		items = append(items, AppHistoryItem{
			Name:        app.Name,
			ProcessName: app.ProcessName,
			ExecPath:    app.ExecPath,
		})
	}

	m.history.Add(HistoryEntry{
		Timestamp: time.Now(),
		Operation: OpRestore,
		Apps:      items,
		Success:   successCount,
		Failed:    failCount,
	})
}

// performUndo undoes the last operation
func (m *model) performUndo() tea.Cmd {
	entry := m.history.GetLast()
	if entry == nil {
		m.logs = []string{"No operations to undo"}
		m.currentState = stateDone
		m.mode = "undo"
		return nil
	}

	// Build undo message
	appNames := make([]string, len(entry.Apps))
	for i, app := range entry.Apps {
		appNames[i] = app.Name
	}

	m.undoMessage = fmt.Sprintf("Undo %s operation on: %s",
		entry.Operation.String(),
		strings.Join(appNames, ", "))

	m.currentState = stateUndoConfirm
	return nil
}

// executeUndo executes the undo operation
func (m *model) executeUndo() tea.Cmd {
	entry := m.history.GetLast()
	if entry == nil {
		m.currentState = stateMenu
		return nil
	}

	// Perform undo immediately (not as a command)
	var msgs []string
	successCount := 0
	failCount := 0

	msgs = append(msgs, fmt.Sprintf("Undoing %s operation...", entry.Operation.String()))

	switch entry.Operation {
	case OpKill:
		// Undo kill = restore processes
		for _, app := range entry.Apps {
			if app.ExecPath == "" {
				msgs = append(msgs, fmt.Sprintf("[SKIP] %s: No executable path", app.Name))
				failCount++
				continue
			}

			cmd := exec.Command(app.ExecPath)
			if err := cmd.Start(); err != nil {
				msgs = append(msgs, fmt.Sprintf("[ERR]  %s: %v", app.Name, err))
				failCount++
			} else {
				msgs = append(msgs, fmt.Sprintf("[OK]   Restored %s", app.Name))
				successCount++
			}
		}

	case OpSuspend:
		// Undo suspend = resume processes
		for _, app := range entry.Apps {
			if len(app.PIDs) == 0 {
				msgs = append(msgs, fmt.Sprintf("[SKIP] %s: No PIDs recorded", app.Name))
				failCount++
				continue
			}

			resumed := 0
			for _, pid := range app.PIDs {
				if pidExists(pid) {
					if err := resumeProcess(pid); err == nil {
						resumed++
					}
				}
			}

			if resumed > 0 {
				msgs = append(msgs, fmt.Sprintf("[OK]   Resumed %s (%d processes)", app.Name, resumed))
				successCount++

				// Clear PIDs from config
				for i := range m.config.Apps {
					if m.config.Apps[i].Name == app.Name {
						for _, pid := range app.PIDs {
							delete(m.config.Apps[i].PIDs, pid)
						}
						break
					}
				}
			} else {
				msgs = append(msgs, fmt.Sprintf("[ERR]  %s: No valid PIDs found", app.Name))
				failCount++
			}
		}

	case OpResume:
		// Undo resume = re-suspend processes
		for _, app := range entry.Apps {
			appRef := m.findAppByName(app.Name)
			if appRef == nil {
				msgs = append(msgs, fmt.Sprintf("[SKIP] %s: Not found in config", app.Name))
				failCount++
				continue
			}

			if err := suspendProcessByName(appRef.ProcessName, appRef); err != nil {
				msgs = append(msgs, fmt.Sprintf("[ERR]  %s: %v", app.Name, err))
				failCount++
			} else {
				msgs = append(msgs, fmt.Sprintf("[OK]   Re-suspended %s", app.Name))
				successCount++
			}
		}

	case OpRestore:
		// Undo restore = kill processes
		for _, app := range entry.Apps {
			procs, _ := process.Processes()
			killed := false
			for _, p := range procs {
				name, _ := p.Name()
				if strings.EqualFold(name, app.ProcessName) {
					p.Kill()
					killed = true
				}
			}
			if killed {
				msgs = append(msgs, fmt.Sprintf("[OK]   Killed %s", app.Name))
				successCount++
			} else {
				msgs = append(msgs, fmt.Sprintf("[ERR]  %s: Not running", app.Name))
				failCount++
			}
		}
	}

	// Remove this entry from history after undo
	if len(m.history.Entries) > 0 {
		m.history.Entries = m.history.Entries[:len(m.history.Entries)-1]
	}

	summary := fmt.Sprintf("Undo complete: %d succeeded, %d failed", successCount, failCount)
	msgs = append(msgs, "", summary)

	// Update model directly
	m.logs = msgs
	m.progPercent = 1.0
	m.currentState = stateDone

	return nil
}

// undoCmd performs the actual undo operation
func (m *model) undoCmd(entry *HistoryEntry) tea.Cmd {
	return func() tea.Msg {
		var msgs []string
		successCount := 0
		failCount := 0

		switch entry.Operation {
		case OpKill:
			// Undo kill = restore processes
			for _, app := range entry.Apps {
				if app.ExecPath == "" {
					msgs = append(msgs, fmt.Sprintf("[SKIP] %s: No executable path", app.Name))
					failCount++
					continue
				}

				cmd := exec.Command(app.ExecPath)
				if err := cmd.Start(); err != nil {
					msgs = append(msgs, fmt.Sprintf("[ERR]  %s: %v", app.Name, err))
					failCount++
				} else {
					msgs = append(msgs, fmt.Sprintf("[OK]   Restored %s", app.Name))
					successCount++
				}
			}

		case OpSuspend:
			// Undo suspend = resume processes
			for _, app := range entry.Apps {
				if len(app.PIDs) == 0 {
					msgs = append(msgs, fmt.Sprintf("[SKIP] %s: No PIDs recorded", app.Name))
					failCount++
					continue
				}

				resumed := 0
				for _, pid := range app.PIDs {
					if pidExists(pid) {
						if err := resumeProcess(pid); err == nil {
							resumed++
						}
					}
				}

				if resumed > 0 {
					msgs = append(msgs, fmt.Sprintf("[OK]   Resumed %s (%d processes)", app.Name, resumed))
					successCount++

					// Clear PIDs from config
					for i := range m.config.Apps {
						if m.config.Apps[i].Name == app.Name {
							for _, pid := range app.PIDs {
								delete(m.config.Apps[i].PIDs, pid)
							}
							break
						}
					}
				} else {
					msgs = append(msgs, fmt.Sprintf("[ERR]  %s: No valid PIDs found", app.Name))
					failCount++
				}
			}

		case OpResume:
			// Undo resume = re-suspend processes
			for _, app := range entry.Apps {
				// Find matching processes and suspend them again
				appRef := m.findAppByName(app.Name)
				if appRef == nil {
					msgs = append(msgs, fmt.Sprintf("[SKIP] %s: Not found in config", app.Name))
					failCount++
					continue
				}

				if err := suspendProcessByName(appRef.ProcessName, appRef); err != nil {
					msgs = append(msgs, fmt.Sprintf("[ERR]  %s: %v", app.Name, err))
					failCount++
				} else {
					msgs = append(msgs, fmt.Sprintf("[OK]   Re-suspended %s", app.Name))
					successCount++
				}
			}

		case OpRestore:
			// Undo restore = kill processes
			for _, app := range entry.Apps {
				if err := killProcess(app.ProcessName); err != nil {
					msgs = append(msgs, fmt.Sprintf("[ERR]  %s: %v", app.Name, err))
					failCount++
				} else {
					msgs = append(msgs, fmt.Sprintf("[OK]   Killed %s", app.Name))
					successCount++
				}
			}
		}

		// Remove this entry from history after undo
		if len(m.history.Entries) > 0 {
			m.history.Entries = m.history.Entries[:len(m.history.Entries)-1]
		}

		summary := fmt.Sprintf("Undo complete: %d succeeded, %d failed", successCount, failCount)
		msgs = append(msgs, "", summary)

		// Return a processResultMsg to update the UI
		return processResultMsg{
			message: summary,
			percent: 1.0,
			done:    true,
			index:   len(entry.Apps) - 1,
		}
	}
}

// findAppByName finds an app in the config by name
func (m *model) findAppByName(name string) *AppEntry {
	for i := range m.config.Apps {
		if m.config.Apps[i].Name == name {
			return &m.config.Apps[i]
		}
	}
	return nil
}

// scanForProfiles scans current directory for profile JSON files
func scanForProfiles() []profileItem {
	files, err := os.ReadDir(".")
	if err != nil {
		return []profileItem{}
	}

	var profiles []profileItem
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if !strings.HasPrefix(name, "sceneshift-profile-") || !strings.HasSuffix(name, ".json") {
			continue
		}

		// Try to read metadata
		data, err := os.ReadFile(name)
		if err != nil {
			continue
		}

		var profile ConfigProfile
		err = json.Unmarshal(data, &profile)
		if err != nil {
			profiles = append(profiles, profileItem{
				filename:    name,
				description: "Invalid profile",
				date:        "",
			})
			continue
		}

		profiles = append(profiles, profileItem{
			filename:    name,
			description: profile.Metadata.Description,
			date:        profile.Metadata.ExportDate.Format("2006-01-02"),
		})
	}

	return profiles
}

// exportProfile exports the current configuration to a JSON file
func (m *model) exportProfile(description, author string) error {
	profile := ConfigProfile{
		Metadata: ProfileMetadata{
			Version:           "1.0",
			SceneShiftVersion: Version,
			ExportDate:        time.Now(),
			Description:       description,
			Author:            author,
		},
		Apps:       m.config.Apps,
		Presets:    m.config.Presets,
		Theme:      m.config.Theme,
		Protection: m.config.Protection,
		SafeToKill: m.config.SafeToKill,
	}

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %v", err)
	}

	filename := fmt.Sprintf("sceneshift-profile-%s.json", time.Now().Format("2006-01-02"))
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	m.profileMessage = fmt.Sprintf("✅ Profile exported to: %s", filename)
	return nil
}

// importProfile imports a configuration from a JSON file
func (m *model) importProfile(filepath string, mergeMode bool) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	var profile ConfigProfile
	err = json.Unmarshal(data, &profile)
	if err != nil {
		return fmt.Errorf("invalid profile format: %v", err)
	}

	// Version check
	if profile.Metadata.SceneShiftVersion > Version {
		m.profileMessage = fmt.Sprintf("⚠️ Warning: Profile from newer version (%s)", profile.Metadata.SceneShiftVersion)
	}

	if mergeMode {
		// Merge: Add to existing config
		m.config.Apps = append(m.config.Apps, profile.Apps...)
		m.config.Presets = append(m.config.Presets, profile.Presets...)
		// Don't merge theme, protection, or safe-to-kill lists
		m.profileMessage = fmt.Sprintf("✅ Merged %d apps and %d presets", len(profile.Apps), len(profile.Presets))
	} else {
		// Replace: Overwrite existing config
		m.config.Apps = profile.Apps
		m.config.Presets = profile.Presets
		m.config.Theme = profile.Theme
		m.config.Protection = profile.Protection
		m.config.SafeToKill = profile.SafeToKill
		m.profileMessage = "✅ Profile imported successfully"
	}

	m.saveConfig()
	return nil
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
		// Ensure minimum heights to prevent negative values
		procHeight := msg.Height - 6
		if procHeight < 1 {
			procHeight = 1
		}
		themeHeight := msg.Height - 4
		if themeHeight < 1 {
			themeHeight = 1
		}
		profileHeight := msg.Height - 8
		if profileHeight < 1 {
			profileHeight = 1
		}
		m.procList.SetSize(msg.Width, procHeight)
		m.themeList.SetSize(msg.Width, themeHeight)
		m.profileList.SetSize(msg.Width, profileHeight)

	case tea.KeyMsg:
		// Global Quit (Context Aware)
		safeStates := []state{stateMenu, statePresetList, stateThemePicker, stateSafelistManager}
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

			case key.Matches(msg, m.keys.SafelistMenu):
				m.currentState = stateSafelistManager
				m.safelistCursor = 0
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

			case msg.String() == "h":
				if m.history.IsEmpty() {
					return m, nil
				}
				m.currentState = stateHistory
				m.historyCursor = len(m.history.Entries) - 1
				return m, nil

			case msg.String() == "u", msg.String() == "ctrl+z":
				return m, m.performUndo()

			case msg.String() == "ctrl+e":
				// Export profile
				m.currentState = stateProfileExport
				m.profileDescription = ""
				m.profileAuthor = ""
				m.setupProfileInputs()
				return m, nil

			case msg.String() == "i":
				// Import profile - scan for available profiles
				profiles := scanForProfiles()

				items := make([]list.Item, len(profiles))
				for i, p := range profiles {
					items[i] = p
				}

				m.profileList = list.New(items, list.NewDefaultDelegate(), 0, 0)
				m.profileList.Title = "Select Profile to Import"
				m.profileList.SetShowHelp(false)

				m.currentState = stateProfileImport
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
			case key.Matches(msg, m.keys.Suspend):
				m.mode = "suspend"
				m.currentState = stateCountdown
				m.countdown = 5
				m.progress = progress.New(
					progress.WithGradient(m.config.Theme.Suspend, m.config.Theme.Highlight),
					progress.WithWidth(40),
				)
				return m, tickCmd()
			case key.Matches(msg, m.keys.Resume):
				m.mode = "resume"
				m.currentState = stateCountdown
				m.countdown = 5
				m.progress = progress.New(
					progress.WithGradient(m.config.Theme.Restore, m.config.Theme.Highlight),
					progress.WithWidth(40),
				)
				return m, tickCmd()
			case key.Matches(msg, m.keys.SafelistMenu):
				m.currentState = stateSafelistManager
				m.safelistCursor = 0
				return m, nil

			// NEW: History keybindings
			case msg.String() == "h":
				if m.history.IsEmpty() {
					// Show a message or do nothing
					return m, nil
				}
				m.currentState = stateHistory
				m.historyCursor = len(m.history.Entries) - 1
				return m, nil

			case msg.String() == "u":
				return m, m.performUndo()

			case msg.String() == "ctrl+z":
				return m, m.performUndo()
				// END NEW

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

		case stateSafelistManager:
			var cmd tea.Cmd
			switch {
			case key.Matches(msg, m.keys.Quit), msg.String() == "esc":
				m.currentState = stateMenu
				return m, nil
			case key.Matches(msg, m.keys.Up):
				if m.safelistCursor > 0 {
					m.safelistCursor--
				}
			case key.Matches(msg, m.keys.Down):
				if m.safelistCursor < len(m.config.Protection.ExclusionList)-1 {
					m.safelistCursor++
				}
			case key.Matches(msg, m.keys.DeleteItem):
				if len(m.config.Protection.ExclusionList) > 0 {
					m.config.Protection.ExclusionList = append(
						m.config.Protection.ExclusionList[:m.safelistCursor],
						m.config.Protection.ExclusionList[m.safelistCursor+1:]...,
					)
					if m.safelistCursor >= len(m.config.Protection.ExclusionList) && m.safelistCursor > 0 {
						m.safelistCursor--
					}
					m.saveConfig()
				}
			case msg.String() == "enter":
				newProc := strings.TrimSpace(m.safelistInput.Value())
				if newProc != "" {
					m.config.Protection.ExclusionList = append(m.config.Protection.ExclusionList, newProc)
					m.safelistInput.SetValue("")
					m.saveConfig()
				}
			default:
				m.safelistInput, cmd = m.safelistInput.Update(msg)
				return m, cmd
			}
			return m, nil

		case stateHistory:
			switch msg.String() {
			case "up", "k":
				if m.historyCursor < len(m.history.Entries)-1 {
					m.historyCursor++
				}
				return m, nil

			case "down", "j":
				if m.historyCursor > 0 {
					m.historyCursor--
				}
				return m, nil

			case "u":
				// Undo last operation (most recent)
				if !m.history.IsEmpty() {
					return m, m.performUndo()
				}
				return m, nil

			case "esc":
				m.currentState = stateMenu
				return m, nil
			}

		case stateUndoConfirm:
			switch msg.String() {
			case "enter":
				return m, m.executeUndo()

			case "esc":
				m.currentState = stateMenu
				m.undoMessage = ""
				return m, nil
			}

		case stateProfileExport:
			switch msg.String() {
			case "enter":
				description := m.inputs[0].Value()
				author := m.inputs[1].Value()

				if description == "" {
					description = "SceneShift configuration"
				}

				err := m.exportProfile(description, author)
				if err != nil {
					m.profileMessage = fmt.Sprintf("❌ Export failed: %v", err)
				}

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

			// Update inputs
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

		case stateProfileImport:
			switch msg.String() {
			case "enter":
				if m.profileList.SelectedItem() != nil {
					filename := m.profileList.SelectedItem().(profileItem).filename

					err := m.importProfile(filename, true)
					if err != nil {
						m.profileMessage = fmt.Sprintf("❌ Import failed: %v", err)
					}

					m.currentState = stateMenu
				}
				return m, nil

			case "esc":
				m.currentState = stateMenu
				return m, nil
			}

			// Let the list handle the input
			m.profileList, cmd = m.profileList.Update(msg)
			return m, cmd

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
					SafetyLevel: detectSafetyLevel(m.inputs[1].Value(), &m.config), // NEW
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
			// NEW: Record operation in history before marking as done
			selectedApps := make([]AppEntry, 0)
			successCount := 0
			failCount := 0

			// Count successes and failures from logs
			for _, log := range m.logs {
				if strings.Contains(log, "[KILL]") || strings.Contains(log, "[SUSP]") ||
					strings.Contains(log, "[RESM]") || strings.Contains(log, "[RUN]") {
					successCount++
				} else if strings.Contains(log, "[ERR]") {
					failCount++
				}
			}

			// Collect selected apps
			for _, app := range m.config.Apps {
				if app.Selected {
					selectedApps = append(selectedApps, app)
				}
			}

			// Record based on mode
			switch m.mode {
			case "kill":
				m.recordKillOperation(selectedApps, successCount, failCount)
			case "suspend":
				m.recordSuspendOperation(selectedApps, successCount, failCount)
			case "resume":
				m.recordResumeOperation(selectedApps, successCount, failCount)
			case "restore":
				m.recordRestoreOperation(selectedApps, successCount, failCount)
			}
			// END NEW CODE

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

func (m *model) setupProfileInputs() {
	if m.currentState == stateProfileExport {
		m.inputs = make([]textinput.Model, 2)
		for i := range m.inputs {
			t := textinput.New()
			t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Highlight))
			m.inputs[i] = t
		}
		m.inputs[0].Prompt = "Description: "
		m.inputs[0].Placeholder = "My gaming optimization setup"
		m.inputs[1].Prompt = "Author (optional): "
		m.inputs[1].Placeholder = "Your name"
		m.focusIndex = 0
		m.inputs[0].Focus()
	} else if m.currentState == stateProfileImport {
		m.inputs = make([]textinput.Model, 1)
		t := textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Highlight))
		t.Prompt = "File path: "
		t.Placeholder = "sceneshift-profile-2026-01-28.json"
		t.Focus()
		m.inputs[0] = t
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
		var successCount, failedCount int
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

		// Check exclusion list (protected processes)
		if isInSafelist(app.ProcessName, m.config.Protection.ExclusionList) {
			return processResultMsg{
				message: fmt.Sprintf("[🛡️ PROTECTED] %s cannot be modified", app.Name),
				percent: percent,
				done:    false,
				index:   index,
			}
		}

		var msg string
		switch m.mode {
		case "kill":
			err := killProcess(app.ProcessName)
			if err != nil {
				msg = fmt.Sprintf("[ERR]  %s: %v", app.Name, err)
				failedCount++
			} else {
				msg = fmt.Sprintf("[KILL] Terminated %s", app.Name)
				successCount++
			}
		case "suspend":
			// Get reference to the actual app in config
			appRef := &m.config.Apps[index]
			err := suspendProcessByName(appRef.ProcessName, appRef)
			if err != nil {
				msg = fmt.Sprintf("[ERR]  %s: %v", app.Name, err)
				failedCount++
			} else {
				msg = fmt.Sprintf("[SUSP] Suspended %s", app.Name)
				successCount++
			}
		case "resume":
			// Get reference to the actual app in config
			appRef := &m.config.Apps[index]
			err := resumeProcessByName(appRef)
			if err != nil {
				msg = fmt.Sprintf("[ERR]  %s: %v", app.Name, err)
				failedCount++
			} else {
				msg = fmt.Sprintf("[RESM] Resumed %s", app.Name)
				successCount++
			}
		case "restore":
			if app.ExecPath == "" {
				msg = fmt.Sprintf("[SKIP] %s: no path", app.Name)
				failedCount++
			} else {
				cmd := exec.Command(app.ExecPath)
				if err := cmd.Start(); err != nil {
					msg = fmt.Sprintf("[ERR]  %s: %v", app.Name, err)
					failedCount++
				} else {
					msg = fmt.Sprintf("[REST] Launched %s", app.Name)
					successCount++
				}
			}
		}
		return processResultMsg{message: msg, percent: percent, done: false, index: index}
	}
}

func suspendProcessByName(rawNames string, app *AppEntry) error {
	procs, err := process.Processes()
	if err != nil {
		return err
	}
	targets := strings.Split(rawNames, ",")
	for i := range targets {
		targets[i] = strings.TrimSpace(targets[i])
	}

	// Initialize PIDs map if needed
	if app.PIDs == nil {
		app.PIDs = make(map[int32]bool)
	}

	var lastErr error
	suspendedCount := 0
	for _, p := range procs {
		n, err := p.Name()
		if err != nil {
			continue
		}
		for _, t := range targets {
			if strings.EqualFold(n, t) {
				if err := suspendProcess(p.Pid); err != nil {
					lastErr = err
				} else {
					// Record the PID
					app.PIDs[p.Pid] = true
					suspendedCount++
				}
			}
		}
	}
	if suspendedCount > 0 {
		return nil
	}
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("no processes found")
}

func resumeProcessByName(app *AppEntry) error {
	// No PIDs tracked = nothing to resume
	if len(app.PIDs) == 0 {
		return fmt.Errorf("no suspended processes found for %s", app.Name)
	}

	var lastErr error
	resumedCount := 0
	var invalidPIDs []int32

	for pid := range app.PIDs {
		// Validate PID still exists
		if !pidExists(pid) {
			invalidPIDs = append(invalidPIDs, pid)
			continue
		}

		// Optional: Validate executable path matches
		if app.ExecPath != "" {
			p, err := process.NewProcess(pid)
			if err == nil {
				if exe, err := p.Exe(); err == nil {
					if !strings.EqualFold(exe, app.ExecPath) {
						invalidPIDs = append(invalidPIDs, pid)
						continue
					}
				}
			}
		}

		if err := resumeProcess(pid); err != nil {
			lastErr = err
		} else {
			// Remove PID after successful resume
			delete(app.PIDs, pid)
			resumedCount++
		}
	}

	// Clean up invalid PIDs
	for _, pid := range invalidPIDs {
		delete(app.PIDs, pid)
	}

	if len(invalidPIDs) > 0 {
		return fmt.Errorf("resumed %d, %d PIDs no longer valid", resumedCount, len(invalidPIDs))
	}

	if resumedCount > 0 {
		return nil
	}
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("no valid PIDs to resume")
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

				// Safety indicator
				safetyIcon := ""
				switch app.SafetyLevel {
				case "protected":
					safetyIcon = "🛡️ "
				case "safe":
					safetyIcon = "✓ "
				case "caution":
					safetyIcon = "⚠ "
				default:
					safetyIcon = "  "
				}

				// NEW: Status indicator and stats
				status := getProcessStatus(app)
				statusIcon := getStatusIcon(status)

				// Get stats from cache
				stats := m.statsCache.Get(app.ProcessName)
				statsStr := ""
				if stats.IsRunning {
					statsStr = fmt.Sprintf(" CPU: %.1f%% RAM: %d MB", stats.CPUPercent, stats.RAMMB)
				}

				label := fmt.Sprintf("%s %s %s%s %s%s", cursor, check, safetyIcon, app.Name, statusIcon, statsStr)

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
		s += "\n" + lipgloss.NewStyle().Faint(true).Render("Legend: 🛡️=Protected  ✓=Safe  ⚠=Caution  |  ▶️=Running  ⏸️=Suspended") + "\n"

		// Show profile message if present
		if m.profileMessage != "" {
			s += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Highlight)).Render(m.profileMessage) + "\n"
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

	case stateSafelistManager:
		titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Select)).Bold(true).Underline(true)
		s += titleStyle.Render("EXCLUSION LIST MANAGER") + "\n\n"
		s += lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warn)).Render("🛡️ Protected processes that cannot be killed or suspended:") + "\n\n"

		if len(m.config.Protection.ExclusionList) == 0 {
			s += lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warn)).Render("No processes in exclusion list.") + "\n"
		} else {
			for i, proc := range m.config.Protection.ExclusionList {
				cursor := "  "
				if m.safelistCursor == i {
					cursor = "> "
				}
				label := fmt.Sprintf("%s%s", cursor, proc)
				if m.safelistCursor == i {
					s += selected.Render(label) + "\n"
				} else {
					s += unselected.Render(label) + "\n"
				}
			}
		}
		s += "\n" + m.safelistInput.View() + "\n"
		s += lipgloss.NewStyle().Faint(true).Render("\n(Enter to Add, d: delete, esc: back)")

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
		suspendStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Suspend)).Bold(true)

		switch m.mode {
		case "kill":
			modeStr = killStyle.Render("KILLING APPS")
		case "suspend":
			modeStr = suspendStyle.Render("SUSPENDING APPS")
		case "resume":
			modeStr = restoreStyle.Render("RESUMING APPS")
		default:
			modeStr = restoreStyle.Render("LAUNCHING APPS")
		}
		s += fmt.Sprintf("\n   %s IN...\n\n", modeStr)
		bigNum := lipgloss.NewStyle().Bold(true).Padding(1, 3).Foreground(lipgloss.Color(m.config.Theme.Warn)).Render(fmt.Sprintf("%d", m.countdown))
		s += fmt.Sprintf("      %s", bigNum)
		s += "\n\n   Press q to cancel."

	case stateProcessing, stateDone:
		var modeStr string
		suspendStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Suspend)).Bold(true)

		switch m.mode {
		case "kill":
			modeStr = killStyle.Render("KILLING...")
		case "suspend":
			modeStr = suspendStyle.Render("SUSPENDING...")
		case "resume":
			modeStr = restoreStyle.Render("RESUMING...")
		default:
			modeStr = restoreStyle.Render("LAUNCHING...")
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

	case stateHistory:
		titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Select)).Bold(true).Underline(true)
		selected := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Select)).Bold(true)
		unselected := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Text)).Faint(true)

		s += titleStyle.Render("📜 SESSION HISTORY") + "\n\n"

		if m.history.IsEmpty() {
			s += lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warn)).Render("No operations recorded this session.") + "\n"
		} else {
			// Display history entries (newest first)
			for i := len(m.history.Entries) - 1; i >= 0; i-- {
				entry := m.history.Entries[i]

				cursor := "  "
				if i == m.historyCursor {
					cursor = "> "
				}

				// Format timestamp
				timeStr := entry.Timestamp.Format("15:04:05")

				// Get app names
				appNames := make([]string, len(entry.Apps))
				for j, app := range entry.Apps {
					appNames[j] = app.Name
				}
				appsStr := strings.Join(appNames, ", ")
				if len(appsStr) > 50 {
					appsStr = appsStr[:47] + "..."
				}

				// Format line
				line := fmt.Sprintf("%s[%s] %s - %d apps (%s)",
					cursor,
					timeStr,
					entry.Operation.String(),
					len(entry.Apps),
					appsStr)

				if i == m.historyCursor {
					s += selected.Render(line) + "\n"
				} else {
					s += unselected.Render(line) + "\n"
				}
			}
		}

		s += "\n" + lipgloss.NewStyle().Faint(true).Render("↑/↓: Navigate • u: Undo last • Esc: Back") + "\n"

	case stateUndoConfirm:
		titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Select)).Bold(true).Underline(true)
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warn))

		s += titleStyle.Render("⚠️  CONFIRM UNDO") + "\n\n"
		s += m.undoMessage + "\n\n"
		s += warnStyle.Render("This will reverse the operation.") + "\n"
		s += warnStyle.Render("This action cannot be undone.") + "\n\n"
		s += lipgloss.NewStyle().Faint(true).Render("Enter: Confirm • Esc: Cancel") + "\n"

	case stateProfileExport:
		titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Select)).Bold(true).Underline(true)
		inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Highlight)).MarginBottom(1)

		s += titleStyle.Render("💾 EXPORT PROFILE") + "\n\n"
		s += lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Text)).Render("Save your configuration to share or backup.") + "\n\n"

		for i := range m.inputs {
			s += inputStyle.Render(m.inputs[i].View()) + "\n"
		}

		s += "\n" + lipgloss.NewStyle().Faint(true).Render("Tab: Next field • Enter: Export • Esc: Cancel") + "\n"

	case stateProfileImport:
		titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Select)).Bold(true).Underline(true)
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warn))

		s += titleStyle.Render("📥 IMPORT PROFILE") + "\n\n"
		s += warnStyle.Render("⚠️ This will merge the profile with your current config.") + "\n\n"

		if m.profileList.Items() == nil || len(m.profileList.Items()) == 0 {
			s += lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Theme.Warn)).Render("No profiles found in current directory.") + "\n\n"
			s += lipgloss.NewStyle().Faint(true).Render("Export a profile first (Ctrl+E) or place profile files here.") + "\n"
		} else {
			s += m.profileList.View()
		}

		s += "\n" + lipgloss.NewStyle().Faint(true).Render("↑/↓: Navigate • Enter: Import • Esc: Cancel") + "\n"

	}

	return lipgloss.NewStyle().Padding(2, 4).Render(s)
}

func main() {
	// Handle version flag
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v", "version":
			fmt.Printf("╭─────────────────────────────────╮\n")
			fmt.Printf("│   SceneShift v%-18s│\n", Version)
			fmt.Printf("│   Built: %-23s│\n", BuildDate)
			if GitCommit != "unknown" {
				fmt.Printf("│   Commit: %-22s│\n", GitCommit[:7])
			}
			fmt.Printf("│   Platform: Windows             │\n")
			fmt.Printf("│   License: MIT                  │\n")
			fmt.Printf("│   Author: tandukuda             │\n")
			fmt.Printf("╰─────────────────────────────────╯\n")
			os.Exit(0)
		case "--help", "-h", "help":
			printHelp()
			os.Exit(0)
		}
	}

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf(`SceneShift v%s - Terminal Process Manager

USAGE:
    SceneShift.exe [OPTIONS]

OPTIONS:
    --version, -v       Show version information
    --help, -h          Show this help message

RUNNING:
    Simply run 'SceneShift.exe' to start the TUI interface

KEYBINDINGS (in TUI):
    K                   Kill selected processes
    S                   Suspend selected processes
    U                   Resume suspended processes
    R                   Launch/Restore processes

    Space               Toggle selection
    a                   Select all
    x                   Deselect all

    n                   New app entry
    e                   Edit selected app
    d                   Delete selected app

    p                   Manage presets
    t                   Change theme
    w                   Manage safelist

    ?                   Toggle help
    q                   Quit

DOCUMENTATION:
    https://github.com/tandukuda/SceneShift

`, Version)
}
