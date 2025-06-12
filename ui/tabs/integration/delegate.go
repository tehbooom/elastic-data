package integration

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type CompactDelegate struct {
	list.DefaultDelegate
}

func NewCompactDelegate() CompactDelegate {
	d := CompactDelegate{list.NewDefaultDelegate()}
	d.SetSpacing(0)
	d.ShowDescription = false
	return d
}

func (d CompactDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	var str string

	if item, ok := listItem.(DatasetItem); ok {
		str = item.Title()
	} else if item, ok := listItem.(*IntegrationItem); ok {
		str = item.Title()
	} else {
		str = listItem.FilterValue()
	}

	isSelected := index == m.Index()

	if isSelected && len(m.Items()) > 1 {
		str = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#48EFCF")).
			Render(str)
	}

	_, _ = fmt.Fprint(w, str)
}
