package integration

type IntegrationItem struct {
	Name     string
	Selected bool
}

func (i IntegrationItem) FilterValue() string {
	return i.Name
}

func (i IntegrationItem) Title() string {
	prefix := "  "
	if i.Selected {
		prefix = "✓ "
	}
	return prefix + i.Name
}

func (i IntegrationItem) Description() string {
	return ""
}

// ToggleSelected toggles the selected state
func (i *IntegrationItem) ToggleSelected() {
	i.Selected = !i.Selected
}

// Create a new integration item
func NewIntegrationItem(name string, selected bool) *IntegrationItem {
	return &IntegrationItem{
		Name:     name,
		Selected: selected,
	}
}
