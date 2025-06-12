package integration

type IntegrationItem struct {
	Name     string
	Selected bool
}

func (i IntegrationItem) FilterValue() string {
	return i.Name
}

func (i IntegrationItem) Title() string {

	prefix := "☐ "
	if i.Selected {
		prefix = "☑ "
	}

	return prefix + i.Name
}

func (i IntegrationItem) Description() string {
	return ""
}

func (i *IntegrationItem) ToggleSelected() {
	i.Selected = !i.Selected
}

func NewIntegrationItem(name string, selected bool) *IntegrationItem {
	return &IntegrationItem{
		Name:     name,
		Selected: selected,
	}
}

type DatasetItem struct {
	Name                  string
	Selected              bool
	Threshold             int
	Unit                  string
	PreserveEventOriginal bool
	Events                []string
}

func (i DatasetItem) FilterValue() string {
	return i.Name
}

func (i DatasetItem) Title() string {
	prefix := "☐ "
	if i.Selected {
		prefix = "☑ "
	}
	return prefix + i.Name
}

func (i DatasetItem) Description() string {
	return ""
}

func NewDatasetItem(name string, selected bool, threshold int, unit string, preserveEventOriginal bool, events []string) DatasetItem {
	return DatasetItem{
		Name:                  name,
		Selected:              selected,
		Threshold:             threshold,
		Unit:                  unit,
		PreserveEventOriginal: preserveEventOriginal,
		Events:                events,
	}
}
