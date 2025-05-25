package integration

import (
	"github.com/charmbracelet/bubbles/list"
)

type CompactDelegate struct {
	list.DefaultDelegate
}

func NewCompactDelegate() CompactDelegate {
	d := CompactDelegate{list.NewDefaultDelegate()}
	d.Styles.NormalTitle.
		PaddingLeft(0).
		MarginTop(0).
		PaddingTop(0).
		PaddingBottom(0)
	d.Styles.SelectedTitle.
		PaddingLeft(0).
		MarginTop(0).
		PaddingTop(0).
		PaddingBottom(0)
	d.Styles.NormalDesc.
		PaddingLeft(0).
		MarginTop(0).
		PaddingTop(0).
		PaddingBottom(0)
	d.Styles.SelectedDesc.
		PaddingLeft(0).
		MarginTop(0).
		PaddingTop(0).
		PaddingBottom(0)
	d.SetSpacing(0)
	d.ShowDescription = false
	return d
}
