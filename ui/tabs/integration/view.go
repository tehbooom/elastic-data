package integration

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/tehbooom/elastic-data/ui/style"
)

func (m TabModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content strings.Builder

	content.WriteString("\n")

	switch m.state {
	case StateSelectingIntegration:
		m.ensureSelectionVisible(len(m.integrationList.Items()))
		content.WriteString(m.renderMultiColumnList(m.integrationList.Items(), "Available Integrations"))
		content.WriteString("\n\n")

		help := style.FormatHelp(
			"(hjkl/arrows)", "Navigate",
			"(space)", "Toggle",
			"(enter)", "Configure",
			"(pgup/pgdn)", "Scroll",
			"(tab)", "Switch tabs",
			"(ctrl+c)", "Quit",
		)

		content.WriteString(help)

	case StateSelectingDatasets:
		content.WriteString(m.datasetsList.View())
		content.WriteString("\n\n\n")
		help := style.FormatHelp(
			"(space)", "Toggle selection",
			"(enter)", "Configure selected",
			"(q)", "Back",
			"(tab)", "Switch tabs",
			"(ctrl+c)", "Quit",
		)
		content.WriteString(help)

	case StateConfiguringDataset:
		content.WriteString(m.renderConfigForm())
	}

	return content.String()
}

func (m TabModel) renderConfigForm() string {
	item := m.datasetsList.SelectedItem().(DatasetItem)
	form := strings.Builder{}

	form.WriteString(fmt.Sprintf("\n  Configuring: %s\n\n", item.Name))
	form.WriteString(fmt.Sprintf("  Threshold: %s\n", m.thresholdInput.View()))
	form.WriteString(fmt.Sprintf("  Unit: %s\n\n", m.unitInput.View()))
	help := style.FormatHelp(
		"(enter)", "Save",
		"(q)", "Cancel",
		"(tab)", "Switch fields",
	)
	form.WriteString("  " + help)

	return form.String()
}

func (m *TabModel) calculateColumns() int {
	if m.width < 40 {
		return 1
	} else if m.width < 80 {
		return 2
	} else if m.width < 120 {
		return 3
	} else {
		return 4
	}
}

func (m TabModel) renderMultiColumnList(items []list.Item, title string) string {
	if len(items) == 0 {
		return fmt.Sprintf("%s\n\nNo items available", style.TitleStyle.Render(title))
	}

	var content strings.Builder
	content.WriteString(style.TitleStyle.Render(title) + "\n\n")

	columns := m.calculateColumns()
	log.Debug(fmt.Sprintf("Columns is %d", columns))

	minWidth := 20
	if m.width < minWidth {
		return fmt.Sprintf("%s\n\nTerminal too narrow to display items", style.TitleStyle.Render(title))
	}

	columnWidth := (m.width - 4) / columns
	log.Debug(fmt.Sprintf("Columns: %d, Width: %d", columns, m.width))
	if columnWidth < 10 {
		columns = 1
		columnWidth = m.width - 4
	}

	totalItems := len(items)
	rowsNeeded := (totalItems + columns - 1) / columns

	startRow := m.scrollOffset
	endRow := startRow + m.visibleRows
	if endRow > rowsNeeded {
		endRow = rowsNeeded
	}
	log.Debug(fmt.Sprintf("Rendering rows %d-%d of %d total rows (selected: %d)",
		startRow, endRow-1, rowsNeeded, m.selectedIndex))

	for row := startRow; row < endRow; row++ {
		var rowContent strings.Builder

		for col := 0; col < columns; col++ {
			itemIndex := row*columns + col
			if itemIndex >= totalItems {
				continue
			}

			item := items[itemIndex]
			isSelected := itemIndex == m.selectedIndex

			var itemText string
			if integrationItem, ok := item.(*IntegrationItem); ok {
				itemText = integrationItem.Title()
			}

			if len(itemText) > columnWidth-2 {
				itemText = itemText[:columnWidth-5] + "..."
			}

			if isSelected {
				itemText = style.ItemStyle.
					Foreground(lipgloss.Color("#02BCB7")).
					Render(fmt.Sprintf("%-*s", columnWidth-2, itemText))
			} else {
				itemText = style.ItemStyle.Render(fmt.Sprintf("%-*s", columnWidth-2, itemText))
			}

			rowContent.WriteString(itemText)
		}

		content.WriteString(rowContent.String() + "\n")
	}

	return content.String()
}

func (m *TabModel) calculateVisibleRows() int {
	if m.height <= 0 {
		return 1
	}

	overhead := 2

	if m.state != StateSelectingIntegration {
		overhead += 2
	}

	overhead += 2

	overhead += 3

	availableHeight := m.height - overhead
	if availableHeight < 1 {
		return 1
	}
	return availableHeight
}

func (m *TabModel) ensureSelectionVisible(totalItems int) {
	if totalItems == 0 {
		return
	}

	columns := m.calculateColumns()
	if columns <= 0 {
		columns = 1
	}

	rowsNeeded := (totalItems + columns - 1) / columns
	selectedRow := m.selectedIndex / columns

	m.visibleRows = m.calculateVisibleRows()
	if m.visibleRows <= 0 {
		m.visibleRows = 5
	}

	log.Debug(fmt.Sprintf("Selection: idx=%d, row=%d, visibleRows=%d, scrollOffset=%d, totalRows=%d",
		m.selectedIndex, selectedRow, m.visibleRows, m.scrollOffset, rowsNeeded))

	if selectedRow >= m.scrollOffset+m.visibleRows {
		m.scrollOffset = selectedRow - m.visibleRows + 1
		log.Debug(fmt.Sprintf("Scrolling down to offset %d", m.scrollOffset))
	}

	if selectedRow < m.scrollOffset {
		m.scrollOffset = selectedRow
		log.Debug(fmt.Sprintf("Scrolling up to offset %d", m.scrollOffset))
	}

	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
	if m.scrollOffset > rowsNeeded-m.visibleRows {
		m.scrollOffset = rowsNeeded - m.visibleRows
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}
	}
}
