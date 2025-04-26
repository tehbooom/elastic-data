package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
)

type Phase string

const (
	CheckingPhase Phase = "checking"
	SyncingPhase  Phase = "syncing"
	CompletePhase Phase = "complete"
	ErrorPhase    Phase = "error"
)

type LoadingModel struct {
	spinner      spinner.Model
	progress     progress.Model
	width        int
	height       int
	startTime    time.Time
	loadingMsg   string
	downloadMsg  string
	complete     bool
	downloading  bool
	downloadSize int64
	downloaded   int64
	error        string
	phase        Phase
	progressVal  float64
}

func NewLoadingModel() LoadingModel {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
	return LoadingModel{
		spinner:     s,
		progress:    p,
		loadingMsg:  "Checking for latest integrations...",
		downloadMsg: "Pulling latest integrations...",
		phase:       CheckingPhase,
		progressVal: 0.0,
	}
}

// SetSize updates the size of the loading model
func (m *LoadingModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.progress.Width = width / 2 // Set progress bar to half of screen width
}

// Init initializes the loading model
func (m LoadingModel) Init() tea.Cmd {
	m.startTime = time.Now()
	return tea.Batch(
		m.spinner.Tick,
		checkRepository(),
	)
}

// Messages for the loading model
type repositoryCheckMsg struct {
	exists bool
	path   string
}

type syncProgressMsg struct {
	progress float64
}

type syncCompleteMsg struct {
	message string
}

type loadingErrorMsg struct {
	err string
}

type loadingCompleteMsg struct{}

// requestProgressUpdates creates a ticker for updating the progress periodically
func requestProgressUpdates() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("progress")}
	})
}

// getConfigDir returns the path to the configuration directory
func getConfigDir() (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		configHome = filepath.Join(home, ".config")
	}

	// Create ~/.config if it doesn't exist
	if _, err := os.Stat(configHome); os.IsNotExist(err) {
		if err := os.MkdirAll(configHome, 0755); err != nil {
			return "", fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	configDir := filepath.Join(configHome, "elastic-data")

	// Create app config dir if it doesn't exist
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create app config directory: %w", err)
		}
	}

	return configDir, nil
}

// Global counter for progress tracking
var (
	byteCounter int64
	counterMu   sync.Mutex
)

// Custom io.Writer that counts bytes
type countingWriter struct{}

func (cw *countingWriter) Write(p []byte) (int, error) {
	n := len(p)
	counterMu.Lock()
	byteCounter += int64(n)
	counterMu.Unlock()
	return n, nil
}

// Reset the byte counter
func resetByteCounter() {
	counterMu.Lock()
	byteCounter = 0
	counterMu.Unlock()
}

// Get the current byte count
func getByteCount() int64 {
	counterMu.Lock()
	count := byteCounter
	counterMu.Unlock()
	return count
}

// Check if repository exists
func checkRepository() tea.Cmd {
	return func() tea.Msg {
		configDir, err := getConfigDir()
		if err != nil {
			return loadingErrorMsg{err: err.Error()}
		}

		repoDir := filepath.Join(configDir, "integrations")

		// Check if repository already exists
		_, err = os.Stat(filepath.Join(repoDir, ".git"))
		exists := !os.IsNotExist(err)

		return repositoryCheckMsg{
			exists: exists,
			path:   repoDir,
		}
	}
}

// Start syncing the repository
func syncRepository(repoPath string, exists bool) tea.Cmd {
	return func() tea.Msg {
		// Reset the byte counter
		resetByteCounter()

		repoURL := "https://github.com/elastic/integrations"

		// Create a writer that counts bytes
		progressWriter := &countingWriter{}

		if !exists {
			// Clone the repository
			_, err := git.PlainClone(repoPath, false, &git.CloneOptions{
				URL:      repoURL,
				Progress: progressWriter,
			})
			if err != nil {
				return loadingErrorMsg{err: fmt.Sprintf("Failed to clone repository: %v", err)}
			}
			return syncCompleteMsg{message: "Repository cloned successfully"}
		} else {
			// Repository exists, open it and pull
			repo, err := git.PlainOpen(repoPath)
			if err != nil {
				return loadingErrorMsg{err: fmt.Sprintf("Failed to open repository: %v", err)}
			}

			// Get the worktree
			worktree, err := repo.Worktree()
			if err != nil {
				return loadingErrorMsg{err: fmt.Sprintf("Failed to get worktree: %v", err)}
			}

			// Pull the latest changes
			err = worktree.Pull(&git.PullOptions{
				RemoteName: "origin",
				Progress:   progressWriter,
			})
			if err != nil {
				if err == git.NoErrAlreadyUpToDate {
					return syncCompleteMsg{message: "Repository already up to date"}
				}
				return loadingErrorMsg{err: fmt.Sprintf("Failed to pull repository: %v", err)}
			}
			return syncCompleteMsg{message: "Repository updated successfully"}
		}
	}
}

// Update handles messages and updates the model
func (m LoadingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	case repositoryCheckMsg:
		// Change phase to syncing
		m.phase = SyncingPhase
		m.loadingMsg = "Syncing repository..."

		// Start progress update ticker
		return m, tea.Batch(
			syncRepository(msg.path, msg.exists),
			requestProgressUpdates(),
		)

	case tea.KeyMsg:
		if string(msg.Runes) == "progress" && m.phase == SyncingPhase {
			// Update progress based on bytes received
			byteCount := getByteCount()

			// Simulate progress for long operations
			// Increase progress gradually, the exact formula might need adjustment
			progressIncrement := float64(0.005) // Small increment each tick
			if byteCount > 0 {
				// Use byte count for better estimate
				// For git operations, we can't know the total, so we simulate
				progressIncrement = float64(byteCount) / 10000000 // Adjust divisor based on expected size
			}

			m.progressVal += progressIncrement
			if m.progressVal > 0.95 {
				m.progressVal = 0.95 // Cap at 95% until complete
			}

			cmd := m.progress.SetPercent(m.progressVal)

			// Continue requesting progress updates
			return m, tea.Batch(cmd, requestProgressUpdates())
		}

	case syncCompleteMsg:
		m.phase = CompletePhase
		m.complete = true
		m.loadingMsg = msg.message
		m.progressVal = 1.0 // Set to 100% when complete

		cmd := m.progress.SetPercent(1.0)
		return m, tea.Batch(
			cmd,
			tea.Tick(time.Second*2, func(time.Time) tea.Msg {
				return loadingCompleteMsg{}
			}),
		)

	case loadingErrorMsg:
		m.phase = ErrorPhase
		m.error = msg.err
		m.loadingMsg = "Error: " + msg.err
		return m, nil

	case loadingCompleteMsg:
		m.complete = true
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

// View renders the loading screen
func (m LoadingModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content string

	switch m.phase {
	case CheckingPhase:
		content = fmt.Sprintf("%s %s", m.spinner.View(), m.loadingMsg)
	case SyncingPhase:
		// Show both message and progress bar
		barView := m.progress.View()
		content = fmt.Sprintf("%s\n\n%s", m.downloadMsg, barView)
	case CompletePhase:
		content = fmt.Sprintf("✓ %s", m.loadingMsg)
	case ErrorPhase:
		content = fmt.Sprintf("❌ %s", m.loadingMsg)
	}

	// Center the content
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	return style.Render(content)
}

// IsComplete returns whether loading is complete
func (m LoadingModel) IsComplete() bool {
	return m.complete
}
