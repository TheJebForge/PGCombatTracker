package abstract

import "gioui.org/layout"

type Page interface {
	Layout(ctx layout.Context, state GlobalState) error
	SetupWindow(state GlobalState)
}
