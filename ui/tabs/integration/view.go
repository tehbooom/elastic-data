package integration

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/log"
	"github.com/tehbooom/elastic-data/ui/style"
)

func (m *TabModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content strings.Builder

	content.WriteString("\n")

	switch m.state {
	case StateSelectingIntegration:
		m.ensureSelectionVisible(len(m.integrationList.Items()))
		tableContent := m.renderMultiColumnList(m.integrationList.Items(), "Available Integrations")
		borderedTable := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1).
			Render(tableContent)

		content.WriteString(borderedTable)
		content.WriteString("\n\n")

		help := style.FormatHelp(
			"(hjkl/arrows)", "Navigate",
			"(space)", "Toggle",
			"(/)", "Search",
			"(e)", "Show Enabled",
			"(enter)", "Configure",
			"(pgup/pgdn)", "Scroll",
			"(home/g)", "Top",
			"(end/G)", "Bottom",
			"(tab)", "Switch tabs",
			"(esc)", "Exit Search Mode",
			"(ctrl+c)", "Quit",
		)

		content.WriteString(help)

	case StateSelectingDatasets:
		if len(m.datasetsList.Items()) == 0 {
			content.WriteString("No datasets available for this integration.\n")
			return content.String()
		}

		if m.datasetsList.Index() >= len(m.datasetsList.Items()) {
			m.datasetsList.Select(0)
		}

		listView := m.datasetsList.View()

		if m.focusedDatasetComponent == FocusDatasetList {
			styledList := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#FF957D")).
				Padding(0, 1).
				Height(m.datasetsList.Height()).
				Render(listView)

			content.WriteString(styledList)
		} else {
			styledList := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#F5F7FA")).
				Height(m.datasetsList.Height()).
				Padding(0, 1).
				Render(listView)

			content.WriteString(styledList)
		}

		content.WriteString("\n")

		if !m.readmeRendered {
			m.viewport.GotoTop()

			var readMeContent string
			var err error

			readMeContent, err = m.getReadMe()
			if err != nil {
				log.Debug(err)
				readMeContent = "Unable to load README"
			}

			glamourRenderWidth := m.width - 6
			renderer, err := glamour.NewTermRenderer(
				glamour.WithStandardStyle("dark"),
				glamour.WithWordWrap(glamourRenderWidth),
			)

			m.viewport.Width = m.width - 4

			if err != nil {
				log.Debug(err)
				m.viewport.SetContent(readMeContent)
			} else {
				str, err := renderer.Render(readMeContent)
				if err != nil {
					log.Debug(err)
					log.Debug("Error rendering with glamour:", err)
					m.viewport.SetContent(readMeContent)
				} else {
					m.viewport.SetContent(str)
				}
			}
			m.readmeRendered = true
		}

		content.WriteString(m.viewport.View())
		content.WriteString("\n")
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

func (m *TabModel) renderConfigForm() string {
	item := m.datasetsList.SelectedItem().(DatasetItem)
	form := strings.Builder{}

	form.WriteString(fmt.Sprintf("\n  Configuring: %s\n\n", item.Name))
	form.WriteString(fmt.Sprintf("  Threshold: %s\n", m.thresholdInput.View()))
	form.WriteString(fmt.Sprintf("  Unit: %s\n", m.unitInput.View()))
	form.WriteString(fmt.Sprintf("  Preserve Original Event: %s\n\n", m.preserveInput.View()))
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

func (m *TabModel) renderMultiColumnList(items []list.Item, title string) string {
	displayItems := items

	if m.searchMode {
		title = fmt.Sprintf("%s (Search: %s)", title, m.searchQuery)
	}

	if m.filteredItems != nil {
		displayItems = m.filteredItems
	}

	if len(displayItems) == 0 {
		emptyMsg := "No items available"
		if m.searchMode && len(m.searchQuery) > 0 {
			emptyMsg = fmt.Sprintf("No items found for '%s'", m.searchQuery)
		}
		return fmt.Sprintf("%s\n\n%s", style.TitleStyle.Render(title), emptyMsg)
	}

	var content strings.Builder

	if m.searchMode {
		searchPrompt := fmt.Sprintf("Search: %s█", m.searchQuery)
		searchStyle := lipgloss.NewStyle().
			Width(m.width - 6).
			Align(lipgloss.Left).
			Foreground(lipgloss.Color("#FFFF00")).
			Render(searchPrompt)
		content.WriteString(searchStyle + "\n")
	}

	centeredTitle := lipgloss.NewStyle().
		Width(m.width - 6).
		Align(lipgloss.Center).
		Render(style.TitleStyle.Render(title))

	content.WriteString(centeredTitle + "\n\n")

	columns := m.calculateColumns()
	totalItems := len(displayItems)
	rowsNeeded := (totalItems + columns - 1) / columns

	startRow := m.scrollOffset
	endRow := startRow + m.visibleRows
	if endRow > rowsNeeded {
		endRow = rowsNeeded
	}

	t := table.New().
		Width(m.width - 6).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			actualRow := startRow + row
			itemIndex := actualRow*columns + col

			if itemIndex >= totalItems {
				return lipgloss.NewStyle()
			}

			isCursorSelected := itemIndex == m.selectedIndex

			var isItemSelected bool
			if itemIndex < len(displayItems) {
				if integrationItem, ok := displayItems[itemIndex].(*IntegrationItem); ok {
					isItemSelected = integrationItem.Selected
				}
			}

			if isCursorSelected {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("#48EFCF")).
					Bold(true)
			} else if isItemSelected {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("#90EE90"))
			}

			return lipgloss.NewStyle()
		})

	var tableData [][]string

	for row := startRow; row < endRow; row++ {
		var rowData []string
		for col := 0; col < columns; col++ {
			itemIndex := row*columns + col
			if itemIndex >= totalItems {
				rowData = append(rowData, "")
				continue
			}

			item := displayItems[itemIndex]
			var itemText string

			if integrationItem, ok := item.(*IntegrationItem); ok {
				itemText = integrationItem.Title()
			}

			rowData = append(rowData, itemText)
		}
		tableData = append(tableData, rowData)
	}

	t = t.Rows(tableData...)

	content.WriteString(t.Render())
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
