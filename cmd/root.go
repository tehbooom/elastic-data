package cmd

import (
	slog "log"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/tehbooom/elastic-data/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

var (
	cfgFile string

	rootCmd = &cobra.Command{
		Use:     "elastic-data",
		Short:   "",
		Version: "",
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func createModel(path string, debug bool) (ui.Model, *os.File) {
	var loggerFile *os.File

	if debug {
		var fileErr error
		newConfigFile, fileErr := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if fileErr == nil {
			log.SetOutput(newConfigFile)
			log.SetTimeFormat(time.Kitchen)
			log.SetReportCaller(true)
			log.SetLevel(log.DebugLevel)
			log.Debug("Logging to debug.log")
		} else {
			loggerFile, _ = tea.LogToFile("debug.log", "debug")
			slog.Print("Failed setting up logging", fileErr)
		}
	} else {
		log.SetOutput(os.Stderr)
		log.SetLevel(log.FatalLevel)
	}

	return ui.NewModel(path), loggerFile
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&cfgFile,
		"config",
		"c",
		"",
		`use this configuration file
(default lookup:
  1. a .elastic-data.yml file
  2. $ES_DATA_CONFIG env var
  3. $XDG_CONFIG_HOME/elastic-data/config.yml
)`,
	)

	err := rootCmd.MarkPersistentFlagFilename("config", "yaml", "yml")
	if err != nil {
		log.Fatal("Cannot mark config flag as filename", err)
	}

	rootCmd.Flags().Bool(
		"debug",
		false,
		"passing this flag will allow writing debug output to debug.log",
	)

	rootCmd.Flags().BoolP(
		"help",
		"h",
		false,
		"help for elastic-data",
	)

	rootCmd.Run = func(_ *cobra.Command, args []string) {

		debug, err := rootCmd.Flags().GetBool("debug")
		if err != nil {
			log.Fatal("Cannot parse debug flag", err)
		}

		// see https://github.com/charmbracelet/lipgloss/issues/73
		lipgloss.SetHasDarkBackground(termenv.HasDarkBackground())

		model, logger := createModel(cfgFile, debug)
		if logger != nil {
			defer logger.Close()
		}

		p := tea.NewProgram(
			model,
			tea.WithAltScreen(),
			tea.WithReportFocus(),
		)
		if _, err := p.Run(); err != nil {
			log.Fatal("Failed starting the TUI", err)
		}
	}
}
