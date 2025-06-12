package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/go-git/go-git/v5"
	"github.com/tehbooom/elastic-data/ui/errors"
)

const ElasticIntegrationsRepoURL string = "https://github.com/elastic/integrations"

type loadingCompleteMsg struct {
	result string
}

type errMsg error

type LoadingModel struct {
	spinner  spinner.Model
	width    int
	height   int
	complete bool
	result   string
	err      error
}

func NewLoadingModel() LoadingModel {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return LoadingModel{
		spinner: s,
	}
}

func (m *LoadingModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m LoadingModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		checkRepository(),
	)
}

func getConfigDir() (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		configHome = filepath.Join(home, ".config")
	}

	configDir := filepath.Join(configHome, "elastic-data")

	return configDir, nil
}

func checkRepository() tea.Cmd {
	return func() tea.Msg {
		configDir, err := getConfigDir()
		if err != nil {
			return errMsg(err)
		}

		repoDir := filepath.Join(configDir, "integrations")

		_, err = os.Stat(filepath.Join(repoDir, ".git"))
		exists := !os.IsNotExist(err)

		result, err := syncRepository(repoDir, exists)
		if err != nil {
			return errMsg(err)
		}

		return loadingCompleteMsg{result: result}
	}
}

func syncRepository(repoPath string, exists bool) (string, error) {
	if !exists {
		_, err := git.PlainClone(repoPath, false, &git.CloneOptions{
			URL: ElasticIntegrationsRepoURL,
		})
		if err != nil {
			return "", fmt.Errorf("failed to clone repository: %v", err)
		}
		return "Repository cloned successfully", nil
	} else {
		repo, err := git.PlainOpen(repoPath)
		if err != nil {
			return "", fmt.Errorf("failed to open repository: %v", err)
		}

		worktree, err := repo.Worktree()
		if err != nil {
			return "", fmt.Errorf("failed to get worktree: %v", err)
		}

		err = worktree.Pull(&git.PullOptions{
			RemoteName: "origin",
		})
		if err != nil {
			if err == git.NoErrAlreadyUpToDate {
				return "Repository already up to date", nil
			}
			return "", fmt.Errorf("failed to pull repository: %v", err)
		}
		return "Repository updated successfully", nil
	}
}

func (m LoadingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	case errMsg:
		log.Debug(msg)
		return m, func() tea.Msg {
			return errors.ShowErrorMsg{Message: fmt.Sprintf("operation failed: %v", msg)}
		}

	case loadingCompleteMsg:
		m.complete = true
		m.result = msg.result
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m LoadingModel) View() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	if m.err != nil {
		return m.err.Error()
	}

	if m.width == 0 {
		return "Loading..."
	}

	if m.complete {
		return style.Render(m.result)
	}

	loadingText := fmt.Sprintf("%s Pulling latest integrations...", m.spinner.View())
	return style.Render(loadingText)
}

func (m LoadingModel) IsComplete() bool {
	return m.complete
}
