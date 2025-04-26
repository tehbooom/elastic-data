package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Common styles used across the application
var (
	// Colors
	primaryColor   = lipgloss.Color("#7D56F4")
	secondaryColor = lipgloss.Color("#383838")
	textColor      = lipgloss.Color("#FAFAFA")
	subtextColor   = lipgloss.Color("#626262")
	successColor   = lipgloss.Color("#73F273")
	errorColor     = lipgloss.Color("#FF5555")

	// Styles
	titleStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(primaryColor).
			Padding(0, 1).
			Align(lipgloss.Center)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(secondaryColor).
			Padding(0, 1).
			Align(lipgloss.Center)

	helpStyle = lipgloss.NewStyle().
			Foreground(subtextColor).
			Align(lipgloss.Center)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Padding(1, 0)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Padding(1, 0)
)

// Demo data structures to use for simulating data sources
// In a real app, these would come from your actual data sources

// GetAvailableItems returns the list of available items
func GetAvailableItems() []string {
	return []string{
		"Elastic Common Schema",
		"Winlogbeat",
		"Nginx",
		"Apache",
		"Kubernetes",
	}
}

// GetDatasetsByItem returns the list of datasets for a given item
func GetDatasetsByItem(item string) []string {
	switch item {
	case "Elastic Common Schema":
		return []string{
			"Authentication",
			"Network",
			"Process",
		}
	case "Winlogbeat":
		return []string{
			"Security",
			"System",
			"Application",
		}
	case "Nginx":
		return []string{
			"Access",
			"Error",
		}
	case "Apache":
		return []string{
			"Access",
			"Error",
		}
	case "Kubernetes":
		return []string{
			"Pod",
			"Container",
			"Node",
		}
	default:
		return []string{
			"Default",
		}
	}
}

// SaveConfig saves the selected configuration for use in the data generation process
func SaveConfig(items []string, datasets map[string][]string, metrics map[string]string, conn ConnectionDetails) error {
	// In a real app, you'd save this to a config file or pass it to the data generator
	// For this example, we'll just print it to the console

	fmt.Println("Configuration saved:")
	fmt.Println("Selected items:", strings.Join(items, ", "))

	fmt.Println("\nSelected datasets:")
	for item, dsets := range datasets {
		fmt.Printf("  %s: %s\n", item, strings.Join(dsets, ", "))
	}

	fmt.Println("\nSelected metrics:")
	for ds, metric := range metrics {
		fmt.Printf("  %s: %s\n", ds, metric)
	}

	fmt.Println("\nConnection details:")
	fmt.Printf("  URL: %s\n", conn.URL)
	fmt.Printf("  Username: %s\n", conn.Username)
	fmt.Printf("  Password: %s\n", strings.Repeat("*", len(conn.Password)))

	return nil
}
