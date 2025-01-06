package ui

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/parser"
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"fmt"
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/samber/lo"
	"github.com/sqweek/dialog"
	"golang.org/x/exp/shiny/materialdesign/icons"
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
	fileList         *widget.List
	refreshIcon      *widget.Icon
	refreshButton    *widget.Clickable
	browseFileIcon   *widget.Icon
	browseFileButton *widget.Clickable
	exitButton       *widget.Clickable
	settingsIcon     *widget.Icon
	settingsButton   *widget.Clickable

	modalLayer *components.ModalLayer
	dropdown   *components.Dropdown
}

func NewFileSelectionPage() *FileSelectionPage {
	refreshIcon, err := widget.NewIcon(icons.NavigationRefresh)
	if err != nil {
		log.Fatalln(err)
	}

	browseIcon, err := widget.NewIcon(icons.FileFolderOpen)
	if err != nil {
		log.Fatalln(err)
	}

	settingsIcon, err := widget.NewIcon(icons.ActionSettings)
	if err != nil {
		log.Fatalln(err)
	}

	return &FileSelectionPage{
		dirty: true,

		// Widgets
		fileList: &widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
		refreshIcon:   refreshIcon,
		refreshButton: &widget.Clickable{},

		browseFileIcon:   browseIcon,
		browseFileButton: &widget.Clickable{},

		settingsIcon:   settingsIcon,
		settingsButton: &widget.Clickable{},
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
	markers, err := state.FindMarkers(fullPath)

	if err != nil {
		log.Printf("Failed to open markers page: %v\n", err)
		return
	}

	state.SwitchPage(NewMarkersPage(fullPath, markers))

	//if state.OpenFile(fullPath, p.watchFileCheckbox.Value) {
	//	page, err := NewStatisticsPage(state)
	//
	//	if err != nil {
	//		log.Printf("Failed to open statistics page: %v\n", err)
	//		return
	//	}
	//
	//	state.SwitchPage(page)
	//}
}

func (p *FileSelectionPage) Layout(ctx layout.Context, state abstract.GlobalState) error {
	// Refresh files if marked as dirty
	if p.dirty {
		newFiles, err := parser.GetSortedLogFiles(state.GorgonFolder())

		if err != nil {
			return err
		}

		p.files = lo.Map(newFiles, NewFileButton)
		p.dirty = false
	}

	// Mark dirty if refresh button is clicked
	if p.refreshButton.Clicked(ctx) {
		p.dirty = true
	}

	if p.browseFileButton.Clicked(ctx) {
		p.openDialog(state)
	}

	if p.settingsButton.Clicked(ctx) {
		state.SwitchPage(NewSettingsPage())
	}

	layout.Flex{
		Axis: layout.Vertical,
	}.Layout(
		ctx,
		layout.Flexed(1, p.fileListUI(state)),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Background{}.Layout(
				gtx,
				utils.MakeColoredBG(utils.SecondBG),
				p.sidePanelUI(state),
			)
		}),
	)

	return nil
}

func (p *FileSelectionPage) fileListUI(state abstract.GlobalState) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if p.dirty {
			return components.Canvas{
				ExpandHorizontal: true,
				ExpandVertical:   true,
			}.Layout(
				gtx,
				components.CanvasItem{
					Anchor: layout.Center,
					Widget: material.Loader(state.Theme()).Layout,
				},
			)
		}

		return layout.Inset{
			Top:  utils.CommonSpacing,
			Left: utils.CommonSpacing,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return material.List(state.Theme(), p.fileList).Layout(
				gtx,
				len(p.files),
				func(gtx layout.Context, index int) layout.Dimensions {
					button := p.files[index]

					if button.openButton.Clicked(gtx) {
						p.selectFile(path.Join(state.GorgonFolder(), button.file.Name()), state)
					}

					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(
						gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return button.Layout(gtx, state)
						}),
						utils.FlexSpacerH(utils.CommonSpacing),
					)
				},
			)
		})
	}
}

func (p *FileSelectionPage) sidePanelUI(state abstract.GlobalState) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(utils.CommonSpacing).Layout(
			gtx,
			func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(
					gtx,
					layout.Rigid(material.IconButton(state.Theme(), p.refreshButton, p.refreshIcon, "Refresh").Layout),
					utils.FlexSpacerW(utils.CommonSpacing),
					layout.Rigid(material.IconButton(state.Theme(), p.browseFileButton, p.browseFileIcon, "Browse File").Layout),
					utils.FlexSpacerW(utils.CommonSpacing),
					layout.Flexed(1, layout.Spacer{}.Layout),
					layout.Rigid(material.IconButton(state.Theme(), p.settingsButton, p.settingsIcon, "Settings").Layout),
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

func NewFileButton(fileInfo os.FileInfo, _ int) FileButton {
	return FileButton{
		file:       fileInfo,
		openButton: &widget.Clickable{},
	}
}

func (b FileButton) Layout(gtx layout.Context, state abstract.GlobalState) layout.Dimensions {
	return layout.Background{}.Layout(
		gtx,
		utils.MakeRoundedBG(10, utils.LessContrastBg),
		func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(10).Layout(
				gtx,
				func(gtx layout.Context) layout.Dimensions {
					return components.Canvas{
						ExpandHorizontal: true,
					}.Layout(
						gtx,
						components.CanvasItem{
							Offset: image.Point{X: gtx.Dp(0), Y: gtx.Dp(0)},
							Widget: material.Body1(state.Theme(), b.file.Name()).Layout,
						},
						components.CanvasItem{
							Offset: image.Point{X: gtx.Dp(0), Y: gtx.Dp(20)},
							Widget: utils.WithColor(material.Subtitle2(
								state.Theme(),
								fmt.Sprintf("Modified at %v", b.file.ModTime().Format(time.DateTime)),
							), utils.GrayText).Layout,
						},
						components.CanvasItem{
							Anchor: layout.E,
							Widget: material.Button(state.Theme(), b.openButton, "Open").Layout,
						},
					)
				},
			)
		},
	)
}
