package integration

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
