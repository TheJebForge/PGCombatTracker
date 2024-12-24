package ui

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/parser"
	"PGCombatTracker/utils"
	"fmt"
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/sqweek/dialog"
	"image"
	"log"
	"os"
	"path"
	"time"
)

type FileSelectionPage struct {
	files []FileButton
	dirty bool

	// Widgets
	fileList          *widget.List
	refreshButton     *widget.Clickable
	watchFileCheckbox *widget.Bool
	browseFileButton  *widget.Clickable
	exitButton        *widget.Clickable
}

func NewFileSelectionPage() *FileSelectionPage {
	return &FileSelectionPage{
		dirty: true,

		// Widgets
		fileList: &widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
		refreshButton: &widget.Clickable{},
		watchFileCheckbox: &widget.Bool{
			Value: true,
		},
		browseFileButton: &widget.Clickable{},
	}
}

func (p *FileSelectionPage) openDialog(state abstract.GlobalState) {
	fullPath, err := dialog.File().Load()

	if err != nil {
		log.Printf("Failed to browse file: %v\n", err)
		return
	}

	p.selectFile(fullPath, state)
}

func (p *FileSelectionPage) selectFile(fullPath string, state abstract.GlobalState) {
	if state.OpenFile(fullPath, p.watchFileCheckbox.Value) {
		page, err := NewStatisticsPage(state)

		if err != nil {
			log.Printf("Failed to open statistics page: %v\n", err)
			return
		}

		state.SwitchPage(page)
	}
}

func (p *FileSelectionPage) Layout(ctx layout.Context, state abstract.GlobalState) error {
	// Refresh files if marked as dirty
	if p.dirty {
		newFiles, err := parser.GetSortedLogFiles(state.GordonFolder())

		if err != nil {
			return err
		}

		p.files = utils.Map(newFiles, NewFileButton)
		p.dirty = false
	}

	// Mark dirty if refresh button is clicked
	if p.refreshButton.Clicked(ctx) {
		p.dirty = true
	}

	if p.browseFileButton.Clicked(ctx) {
		p.openDialog(state)
	}

	layout.Flex{
		Axis: layout.Horizontal,
	}.Layout(
		ctx,
		layout.Flexed(1, p.fileListUI(state)),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Background{}.Layout(
				gtx,
				MakeColoredBG(abstract.SecondBG),
				p.sidePanelUI(state),
			)
		}),
	)

	return nil
}

func (p *FileSelectionPage) fileListUI(state abstract.GlobalState) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if p.dirty {
			return layout.UniformInset(50).Layout(
				gtx,
				material.Loader(state.Theme()).Layout,
			)
		}

		return layout.Inset{
			Top:  abstract.CommonSpacing,
			Left: abstract.CommonSpacing,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return material.List(state.Theme(), p.fileList).Layout(
				gtx,
				len(p.files),
				func(gtx layout.Context, index int) layout.Dimensions {
					button := p.files[index]

					if button.openButton.Clicked(gtx) {
						p.selectFile(path.Join(state.GordonFolder(), button.file.Name()), state)
					}

					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(
						gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return button.Layout(gtx, state)
						}),
						FlexSpacerH(abstract.CommonSpacing),
					)
				},
			)
		})
	}
}

func (p *FileSelectionPage) sidePanelUI(state abstract.GlobalState) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(abstract.CommonSpacing).Layout(
			gtx,
			func(gtx layout.Context) layout.Dimensions {
				return Canvas{
					ExpandVertical: true,
					MinSize:        image.Point{X: gtx.Dp(320)},
				}.Layout(
					gtx,
					CanvasItem{
						Widget: func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{
								Axis: layout.Vertical,
							}.Layout(
								gtx,
								layout.Rigid(material.CheckBox(
									state.Theme(),
									p.watchFileCheckbox,
									"Watch for changes in file",
								).Layout),
								FlexSpacerH(abstract.CommonSpacing),
								layout.Rigid(material.Button(
									state.Theme(),
									p.browseFileButton,
									"Browse File",
								).Layout),
							)
						},
					},
					CanvasItem{
						Anchor: layout.NE,
						Widget: material.Button(state.Theme(), p.refreshButton, "Refresh").Layout,
					},
				)
			},
		)
	}
}

func (p *FileSelectionPage) SetupWindow(state abstract.GlobalState) {
	state.Window().Option(
		app.MinSize(800, 600),
		app.Decorated(true),
	)
}

type FileButton struct {
	file       os.FileInfo
	openButton *widget.Clickable
}

func NewFileButton(fileInfo os.FileInfo) FileButton {
	return FileButton{
		file:       fileInfo,
		openButton: &widget.Clickable{},
	}
}

func (b FileButton) Layout(gtx layout.Context, state abstract.GlobalState) layout.Dimensions {
	return layout.Background{}.Layout(
		gtx,
		MakeRoundedBG(10, abstract.LessContrastBg),
		func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(10).Layout(
				gtx,
				func(gtx layout.Context) layout.Dimensions {
					return Canvas{
						ExpandHorizontal: true,
					}.Layout(
						gtx,
						CanvasItem{
							Offset: image.Point{X: gtx.Dp(0), Y: gtx.Dp(0)},
							Widget: material.Body1(state.Theme(), b.file.Name()).Layout,
						},
						CanvasItem{
							Offset: image.Point{X: gtx.Dp(0), Y: gtx.Dp(20)},
							Widget: WithColor(material.Subtitle2(
								state.Theme(),
								fmt.Sprintf("Modified at %v", b.file.ModTime().Format(time.DateTime)),
							), abstract.GrayText).Layout,
						},
						CanvasItem{
							Anchor: layout.E,
							Widget: material.Button(state.Theme(), b.openButton, "Open").Layout,
						},
					)
				},
			)
		},
	)
}
