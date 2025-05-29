package integration

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var (
	TitleStyle = lipgloss.NewStyle().MarginLeft(2).Bold(true).Foreground(lipgloss.Color("#F04E98"))
	ItemStyle  = lipgloss.NewStyle().PaddingLeft(4)
	HelpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
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
		content.WriteString(HelpStyle.Render(
			"(hjkl/arrows) Navigate, (space) Toggle, (enter) Configure, (pgup/pgdn) Scroll, (tab) Switch tabs, (ctrl+c) Quit"))

	case StateSelectingDatasets:
		content.WriteString(m.datasetsList.View())
		content.WriteString("\n\n\n")
		content.WriteString(HelpStyle.Render(
			"(space) Toggle selection, (enter) Configure selected, (q) Back, (tab) Switch tabs, (ctrl+c) Quit"))

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
	form.WriteString(HelpStyle.Render("  (enter) Save, (q) Cancel, (tab) Switch fields"))

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
		return fmt.Sprintf("%s\n\nNo items available", TitleStyle.Render(title))
	}

	var content strings.Builder
	content.WriteString(TitleStyle.Render(title) + "\n\n")

	columns := m.calculateColumns()
	log.Debug(fmt.Sprintf("Columns is %d", columns))

	minWidth := 20
	if m.width < minWidth {
		return fmt.Sprintf("%s\n\nTerminal too narrow to display items", TitleStyle.Render(title))
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
				itemText = ItemStyle.
					Foreground(lipgloss.Color("#02BCB7")).
					Render(fmt.Sprintf("%-*s", columnWidth-2, itemText))
			} else {
				itemText = ItemStyle.Render(fmt.Sprintf("%-*s", columnWidth-2, itemText))
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

func (m *TabModel) handleGridNavigation(msg tea.KeyMsg, totalItems int) bool {
	if totalItems == 0 {
		return false
	}

	columns := m.calculateColumns()
	currentRow := m.selectedIndex / columns
	currentCol := m.selectedIndex % columns
	totalRows := (totalItems + columns - 1) / columns
	oldIndex := m.selectedIndex

	switch msg.String() {
	case "k", "up":
		if currentRow > 0 {
			newIndex := (currentRow-1)*columns + currentCol
			if newIndex >= 0 {
				m.selectedIndex = newIndex
			}
		}
	case "j", "down":
		if currentRow < totalRows-1 {
			newIndex := (currentRow+1)*columns + currentCol
			if newIndex < totalItems {
				m.selectedIndex = newIndex
			} else {
				lastRowStart := (totalRows - 1) * columns
				if lastRowStart+currentCol < totalItems {
					m.selectedIndex = lastRowStart + currentCol
				} else {
					m.selectedIndex = totalItems - 1
				}
			}
		}
	case "h", "left":
		if currentCol > 0 {
			m.selectedIndex--
		} else if currentRow > 0 {
			prevRowStart := (currentRow - 1) * columns
			prevRowEnd := min(prevRowStart+columns-1, totalItems-1)
			m.selectedIndex = prevRowEnd
		}
	case "l", "right":
		if currentCol < columns-1 && m.selectedIndex < totalItems-1 {
			nextIndex := m.selectedIndex + 1
			nextRow := nextIndex / columns
			if nextRow == currentRow {
				m.selectedIndex = nextIndex
			}
		} else if currentRow < totalRows-1 {
			m.selectedIndex = (currentRow + 1) * columns
		}
	case "pageup", "ctrl+u":
		newRow := currentRow - m.visibleRows
		if newRow < 0 {
			newRow = 0
		}
		m.selectedIndex = newRow*columns + currentCol
		if m.selectedIndex >= totalItems {
			m.selectedIndex = totalItems - 1
		}
	case "pagedown", "ctrl+d":
		newRow := currentRow + m.visibleRows
		if newRow >= totalRows {
			newRow = totalRows - 1
		}
		newIndex := newRow*columns + currentCol
		if newIndex >= totalItems {
			m.selectedIndex = totalItems - 1
		} else {
			m.selectedIndex = newIndex
		}
	case "home", "g":
		m.selectedIndex = 0
	case "end", "G":
		m.selectedIndex = totalItems - 1
	default:
		return false
	}

	if m.selectedIndex != oldIndex {
		m.ensureSelectionVisible(totalItems)
	}

	return true
}
