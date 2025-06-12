package cmd

import (
	"fmt"
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

func createModel(debug bool) (ui.Model, *os.File) {
	var loggerFile *os.File

	if debug {
		var fileErr error
		newConfigFile, fileErr := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if fileErr == nil {
			log.SetOutput(newConfigFile)
			log.SetTimeFormat(time.RFC3339)
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

	return ui.NewModel(), loggerFile
}

func init() {

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

		model, logger := createModel(debug)
		if logger != nil {
			defer func() {
				if err := logger.Close(); err != nil {
					log.Fatalf("Failed to close logger: %v\n", err)
				}
			}()
		}

		p := tea.NewProgram(
			model,
			tea.WithAltScreen(),
			tea.WithReportFocus(),
		)

		finalModel, err := p.Run()
		if err != nil {
			log.Printf("Error running program: %v", err)
			os.Exit(1)
		}

		if m, ok := finalModel.(ui.Model); ok {
			if hasFatal, message := m.HasFatalError(); hasFatal {
				fmt.Fprintf(os.Stderr, "Fatal error: %s\n", message)
				log.Debug("Fatal error: %s", message)
				os.Exit(1)
			}
		}
	}
}
