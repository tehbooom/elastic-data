package integration

// Integrations Item

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

// Datasets Item

type DatasetItem struct {
	Name      string
	Selected  bool
	Threshold int
	Unit      string
}

func (i DatasetItem) FilterValue() string {
	return i.Name
}

func (i DatasetItem) Title() string {
	prefix := "  "
	if i.Selected {
		prefix = "✓ "
	}
	return prefix + i.Name
}

func (i DatasetItem) Description() string {
	return ""
}

// Create a new dataset item
func NewDatasetItem(name string, selected bool, threshold int, unit string) DatasetItem {
	return DatasetItem{
		Name:      name,
		Selected:  selected,
		Threshold: threshold,
		Unit:      unit,
	}
}
