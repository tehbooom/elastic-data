package integration

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

// Styles for the integration UI
var (
	TitleStyle        = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	ItemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	SelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	CheckboxStyle     = lipgloss.NewStyle().PaddingRight(1)
	InfoStyle         = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("240"))
	ConfigStyle       = lipgloss.NewStyle().PaddingLeft(6).Foreground(lipgloss.Color("132"))
	BreadcrumbStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("110"))
	ActivecrumbStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	HelpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// View renders the tab
func (m TabModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content strings.Builder

	// Render breadcrumb navigation
	breadcrumbs := m.renderBreadcrumbs()
	content.WriteString(breadcrumbs + "\n\n")

	// Render the current view based on state
	switch m.state {
	case StateSelectingIntegration:
		content.WriteString(m.integrationList.View())
		content.WriteString("\n\n")
		content.WriteString(HelpStyle.Render(
			"(space) Toggle selection, (enter) Configure datasets, (tab) Switch tabs, (q) Quit"))

	case StateSelectingDatasets:
		content.WriteString(m.datasetsList.View())
		content.WriteString("\n\n\n")
		content.WriteString(HelpStyle.Render(
			"(space) Toggle selection, (enter) Configure selected, (left) Back to integrations, (tab) Switch tabs, (q) Back"))

	case StateConfiguringDataset:
		content.WriteString(m.renderConfigForm())
	}

	return content.String()
}

// Render breadcrumb navigation
func (m TabModel) renderBreadcrumbs() string {
	if m.state == StateSelectingIntegration {
		return ""
	}

	breadcrumbs := fmt.Sprintf("%s > %s",
		BreadcrumbStyle.Render("Integrations"),
		ActivecrumbStyle.Render(m.currentIntegration))

	if m.state == StateConfiguringDataset {
		item := m.datasetsList.SelectedItem().(DatasetItem)
		breadcrumbs = fmt.Sprintf("%s > %s > %s",
			BreadcrumbStyle.Render("Integrations"),
			BreadcrumbStyle.Render(m.currentIntegration),
			ActivecrumbStyle.Render(item.Name))
	}

	return breadcrumbs
}

// Render configuration form
func (m TabModel) renderConfigForm() string {
	item := m.datasetsList.SelectedItem().(DatasetItem)
	form := strings.Builder{}

	form.WriteString(fmt.Sprintf("\n  Configuring: %s\n\n", item.Name))
	form.WriteString(fmt.Sprintf("  Threshold: %s\n", m.thresholdInput.View()))
	form.WriteString(fmt.Sprintf("  Unit: %s\n\n", m.unitInput.View()))
	form.WriteString(HelpStyle.Render("  (enter) Save, (esc) Cancel, (tab) Switch fields"))

	return form.String()
}
