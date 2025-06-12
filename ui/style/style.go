package style

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	TitleStyle    = lipgloss.NewStyle().MarginLeft(2).Bold(true).Foreground(lipgloss.Color("#FEC514"))
	ItemStyle     = lipgloss.NewStyle().PaddingLeft(4)
	HelpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("239"))
	HelpStyleKeys = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Bold(true)
	ErrorStyle    = lipgloss.NewStyle().
			Background(lipgloss.Color("#2D2D2D")).
			Foreground(lipgloss.Color("#FF6B6B")).
			Padding(0, 1).
			Bold(true)
)

func FormatHelp(items ...string) string {
	if len(items)%2 != 0 {
		panic("formatHelp requires an even number of arguments (key-value pairs)")
	}

	var parts []string
	helpSeparator := HelpStyle.Render("•")

	for i := 0; i < len(items); i += 2 {
		key := items[i]
		desc := items[i+1]
		part := HelpStyleKeys.Render(" "+key) + HelpStyle.Render(" "+desc+" ")
		parts = append(parts, part)
	}

	return strings.Join(parts, helpSeparator)
}
