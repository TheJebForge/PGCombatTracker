package ui

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/parser"
	"PGCombatTracker/ui/components"
	"PGCombatTracker/utils"
	"bufio"
	"fmt"
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/samber/lo"
	"github.com/sqweek/dialog"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"io"
	"log"
	"os"
	"time"
)

type selectableMarker struct {
	abstract.Marker
	selected bool
}

func mapMarkerToSelectable(item abstract.Marker, _ int) selectableMarker {
	return selectableMarker{
		Marker: item,
	}
}

func NewMarkersPage(filePath string, markers []abstract.Marker) *MarkersPage {
	backIcon, err := widget.NewIcon(icons.NavigationArrowBack)
	if err != nil {
		log.Fatalln(err)
	}

	deleteIcon, err := widget.NewIcon(icons.ActionDelete)
	if err != nil {
		log.Fatalln(err)
	}

	return &MarkersPage{
		filePath: filePath,
		markers:  lo.Map(markers, mapMarkerToSelectable),
		markerList: &widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
		markerButtons:       utils.MakeClickableArray(len(markers)),
		deleteIcon:          deleteIcon,
		deleteMarkerButtons: utils.MakeClickableArray(len(markers)),
		overrideChoice:      &widget.Enum{Value: "selection"},
		timeFromEditor:      &widget.Editor{},
		timeToEditor:        &widget.Editor{},
		usernameEditor:      &widget.Editor{},
		backIcon:            backIcon,
		backButton:          &widget.Clickable{},
		openButton:          &widget.Clickable{},
		exportButton:        &widget.Clickable{},
		watchFileCheckbox: &widget.Bool{
			Value: true,
		},
	}
}

type MarkersPage struct {
	filePath            string
	markers             []selectableMarker
	markerList          *widget.List
	markerButtons       []*widget.Clickable
	deleteIcon          *widget.Icon
	deleteMarkerButtons []*widget.Clickable
	overrideChoice      *widget.Enum
	timeFrom            time.Time
	timeFromEditor      *widget.Editor
	timeFromInvalid     bool
	timeTo              time.Time
	timeToEditor        *widget.Editor
	timeToInvalid       bool
	usernameEditor      *widget.Editor
	backIcon            *widget.Icon
	backButton          *widget.Clickable
	openButton          *widget.Clickable
	exportButton        *widget.Clickable
	watchFileCheckbox   *widget.Bool
}

const dateHint = "2018-10-11 22:02:28"
const afterDateHint = "2023-08-20 23:59:59"

var maxTime = time.Unix(1<<63-62135596801, 999999999)

func textEditor(state abstract.GlobalState, editor *widget.Editor, hint string, invalid bool) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Background{}.Layout(
			gtx,
			utils.MakeRoundedBG(10, utils.LessContrastBg),
			func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(utils.CommonSpacing*2).Layout(
					gtx,
					func(gtx layout.Context) layout.Dimensions {
						style := material.Editor(state.Theme(), editor, hint)
						style.HintColor = utils.GrayText
						if invalid {
							style.Color = utils.RedText
						}
						return style.Layout(gtx)
					},
				)
			},
		)
	}
}

const markerTextSize = 14

func iconButtonStyle(state abstract.GlobalState, button *widget.Clickable, icon *widget.Icon, desc string) material.IconButtonStyle {
	style := material.IconButton(state.Theme(), button, icon, desc)
	style.Inset = layout.UniformInset(7)
	style.Size = 24
	return style
}

func (m *MarkersPage) markerListUI(state abstract.GlobalState) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if m.backButton.Clicked(gtx) {
			state.SwitchPage(NewFileSelectionPage())
		}

		return layout.UniformInset(utils.CommonSpacing).Layout(
			gtx,
			func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(
					gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{
							Axis:      layout.Horizontal,
							Alignment: layout.Middle,
						}.Layout(
							gtx,
							layout.Rigid(iconButtonStyle(state, m.backButton, m.backIcon, "Back").Layout),
						)
					}),
					utils.FlexSpacerH(utils.CommonSpacing),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						if len(m.markers) == 0 {
							return components.Canvas{
								ExpandHorizontal: true,
								ExpandVertical:   true,
							}.Layout(
								gtx,
								components.CanvasItem{
									Anchor: layout.Center,
									Widget: func(gtx layout.Context) layout.Dimensions {
										style := material.Label(state.Theme(), 32, "No markers found")
										style.Color = utils.GrayText
										return style.Layout(gtx)
									},
								},
							)
						}

						return material.List(state.Theme(), m.markerList).Layout(
							gtx,
							len(m.markers),
							func(gtx layout.Context, index int) layout.Dimensions {
								if index >= len(m.markers) {
									return layout.Dimensions{}
								}

								last := index+1 == len(m.markers)
								button := m.markerButtons[index]
								deleteButton := m.deleteMarkerButtons[index]
								item := m.markers[index]

								if button.Clicked(gtx) {
									item.selected = !item.selected
									m.markers[index] = item
								}

								if deleteButton.Clicked(gtx) {
									state.DeleteMarker(m.filePath, item.Marker)
									newMarkers, err := state.FindMarkers(m.filePath)
									if err != nil {
										log.Println("Failed to find new markers", err)
									} else {
										m.markers = lo.Map(newMarkers, mapMarkerToSelectable)
										m.markerButtons = utils.MakeClickableArray(len(newMarkers))
										m.deleteMarkerButtons = utils.MakeClickableArray(len(newMarkers))
									}
								}

								style := material.ButtonLayout(state.Theme(), button)
								if !item.selected {
									style.Background = utils.SecondBG
								}

								buttonWidget := func(gtx layout.Context) layout.Dimensions {
									return style.Layout(
										gtx,
										func(gtx layout.Context) layout.Dimensions {
											return layout.UniformInset(utils.CommonSpacing*2).Layout(
												gtx,
												func(gtx layout.Context) layout.Dimensions {
													return layout.Flex{
														Axis:      layout.Horizontal,
														Alignment: layout.Middle,
													}.Layout(
														gtx,
														layout.Rigid(material.Label(state.Theme(), markerTextSize, item.Time.Format(time.DateTime)).Layout),
														utils.FlexSpacerW(utils.CommonSpacing*2),
														layout.Flexed(1, material.Label(state.Theme(), markerTextSize, item.Name).Layout),
													)
												},
											)
										},
									)
								}

								optionalDeleteWidget := func(gtx layout.Context) layout.Dimensions {
									if item.UserDefined {
										return layout.Flex{
											Axis:      layout.Horizontal,
											Alignment: layout.Middle,
										}.Layout(
											gtx,
											layout.Flexed(1, buttonWidget),
											utils.FlexSpacerW(utils.CommonSpacing),
											layout.Rigid(iconButtonStyle(state, deleteButton, m.deleteIcon, "Delete").Layout),
										)
									} else {
										return buttonWidget(gtx)
									}
								}

								if last {
									return optionalDeleteWidget(gtx)
								} else {
									return layout.Flex{
										Axis: layout.Vertical,
									}.Layout(
										gtx,
										layout.Rigid(optionalDeleteWidget),
										utils.FlexSpacerH(utils.CommonSpacing),
									)
								}
							},
						)
					}),
				)
			},
		)
	}
}

func figureOutTimeFrames(markers []selectableMarker) []abstract.MarkerTimeFrame {
	var result []abstract.MarkerTimeFrame

	var start selectableMarker
	wasSelected := false

	for _, marker := range markers {
		if !wasSelected && marker.selected {
			start = marker
			wasSelected = true
		} else if wasSelected && !marker.selected {
			result = append(result, abstract.MarkerTimeFrame{
				User: start.User,
				From: start.Time,
				To:   marker.Time,
			})
			wasSelected = false
		}
	}

	if len(result) == 0 || wasSelected {
		result = append(result, abstract.MarkerTimeFrame{
			User: start.User,
			From: start.Time,
			To:   maxTime,
		})
	}

	return result
}

func (m *MarkersPage) getTimeFrames() []abstract.MarkerTimeFrame {
	switch m.overrideChoice.Value {
	case "selection":
		return figureOutTimeFrames(m.markers)
	case "custom":
		var start time.Time
		end := maxTime

		if !m.timeFrom.IsZero() {
			start = m.timeFrom
		}

		if !m.timeTo.IsZero() {
			end = m.timeTo
		}

		return []abstract.MarkerTimeFrame{
			{
				User: m.usernameEditor.Text(),
				From: start,
				To:   end,
			},
		}
	default:
		return []abstract.MarkerTimeFrame{
			{
				From: time.Time{},
				To:   maxTime,
			},
		}
	}
}

var timeLocation = time.Now().Location()

func processDateEditor(gtx layout.Context, editor *widget.Editor, invalid *bool, timeValue *time.Time) {
	if _, ok := editor.Update(gtx); ok {
		newTime, err := time.ParseInLocation(time.DateTime, editor.Text(), timeLocation)
		if err != nil {
			*invalid = true
		} else {
			*timeValue = newTime
			*invalid = false
		}
	}
}

func (m *MarkersPage) exportWithMarkers() error {
	destinationPath, err := dialog.File().SetStartFile(
		"ExportedLog.txt",
	).Filter(
		"Project Gorgon Log File", "txt",
	).Save()
	if err != nil {
		return err
	}

	sourceFile, err := os.Open(m.filePath)
	if err != nil {
		return err
	}
	defer func(sourceFile *os.File) {
		_ = sourceFile.Close()
	}(sourceFile)

	sourceBuffer := bufio.NewReader(sourceFile)

	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer func(destinationFile *os.File) {
		_ = destinationFile.Close()
	}(destinationFile)

	destinationBuffer := bufio.NewWriter(destinationFile)

	// Files ready, start export
	var relevantMarker *selectableMarker
	markers := m.markers

	writeMarker := func() error {
		_, err = destinationBuffer.WriteString(
			fmt.Sprintf("%v\t%v\n", relevantMarker.Time.Format(parser.TimeFormat), &abstract.MarkerLine{
				User: relevantMarker.User,
				Name: relevantMarker.Name,
			}),
		)
		return err
	}

	for {
		sourceLine, err := sourceBuffer.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if relevantMarker != nil {
					err = writeMarker()
					if err != nil {
						return err
					}
				}

				break
			} else {
				return err
			}
		}

		event := parser.ParseLine(sourceLine)

		if event == nil {
			continue
		}

		if relevantMarker == nil {
			for i, marker := range markers {
				if marker.UserDefined {
					relevantMarker = &marker
					if i+1 < len(markers) {
						markers = markers[(i + 1):]
					} else {
						markers = nil
					}
					break
				}
			}
		}

		if relevantMarker != nil && relevantMarker.Time.Before(event.Time) {
			err = writeMarker()
			if err != nil {
				return err
			}
			relevantMarker = nil
		}

		_, err = destinationBuffer.WriteString(sourceLine)
		if err != nil {
			return err
		}
	}

	err = destinationBuffer.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (m *MarkersPage) sidePanelUI(state abstract.GlobalState) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		panelWidth := gtx.Dp(300)

		processDateEditor(gtx, m.timeFromEditor, &m.timeFromInvalid, &m.timeFrom)
		processDateEditor(gtx, m.timeToEditor, &m.timeToInvalid, &m.timeTo)

		if m.exportButton.Clicked(gtx) {
			err := m.exportWithMarkers()
			if err != nil {
				log.Println(err)
			}
		}

		if m.openButton.Clicked(gtx) {
			if state.OpenFile(m.filePath, m.watchFileCheckbox.Value, m.getTimeFrames()) {
				page, err := NewStatisticsPage(state, m.filePath)

				if err != nil {
					log.Printf("Failed to open statistics page: %v\n", err)
				} else {
					state.SwitchPage(page)
				}
			}
		}

		return layout.Background{}.Layout(
			gtx,
			utils.MakeColoredBG(utils.SecondBG),
			func(gtx layout.Context) layout.Dimensions {
				return utils.LayoutMinimalX(panelWidth, layout.UniformInset(utils.CommonSpacing*2).Layout(
					gtx,
					func(gtx layout.Context) layout.Dimensions {
						cgtx := gtx
						cgtx.Constraints.Min.X = panelWidth
						return layout.Flex{
							Axis: layout.Vertical,
						}.Layout(
							cgtx,
							layout.Rigid(material.CheckBox(
								state.Theme(),
								m.watchFileCheckbox,
								"Watch for changes in file",
							).Layout),
							utils.FlexSpacerH(utils.CommonSpacing*2),
							layout.Rigid(material.RadioButton(state.Theme(), m.overrideChoice, "selection", "Use selected markers").Layout),
							utils.FlexSpacerH(utils.CommonSpacing*2),
							layout.Rigid(material.RadioButton(state.Theme(), m.overrideChoice, "everything", "Just load everything").Layout),
							utils.FlexSpacerH(utils.CommonSpacing*2),
							layout.Rigid(material.RadioButton(state.Theme(), m.overrideChoice, "custom", "Use custom time frame").Layout),
							utils.FlexSpacerH(utils.CommonSpacing),
							layout.Rigid(material.Body2(state.Theme(), "Start with user:").Layout),
							utils.FlexSpacerH(utils.CommonSpacing),
							layout.Rigid(textEditor(state, m.usernameEditor, "JohnDoe", false)),
							utils.FlexSpacerH(utils.CommonSpacing),
							layout.Rigid(material.Body2(state.Theme(), "Read from:").Layout),
							utils.FlexSpacerH(utils.CommonSpacing),
							layout.Rigid(textEditor(state, m.timeFromEditor, dateHint, m.timeFromInvalid)),
							utils.FlexSpacerH(utils.CommonSpacing),
							layout.Rigid(material.Body2(state.Theme(), "Read until:").Layout),
							utils.FlexSpacerH(utils.CommonSpacing),
							layout.Rigid(textEditor(state, m.timeToEditor, afterDateHint, m.timeToInvalid)),
							layout.Flexed(1, layout.Spacer{}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								style := material.Button(state.Theme(), m.exportButton, "Export File with Markers")
								const horizontalInset = 78
								const verticalInset = 5
								style.Inset = layout.Inset{
									Left: horizontalInset, Right: horizontalInset,
									Top: verticalInset, Bottom: verticalInset,
								}
								style.TextSize = 14

								return style.Layout(gtx)
							}),
							utils.FlexSpacerH(utils.CommonSpacing),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								style := material.Button(state.Theme(), m.openButton, "Open")
								const horizontalInset = 135
								const verticalInset = 20
								style.Inset = layout.Inset{
									Left: horizontalInset, Right: horizontalInset,
									Top: verticalInset, Bottom: verticalInset,
								}
								style.TextSize = 20

								return style.Layout(gtx)
							}),
						)
					},
				))
			},
		)
	}
}

func (m *MarkersPage) Layout(ctx layout.Context, state abstract.GlobalState) error {
	layout.Flex{
		Axis: layout.Horizontal,
	}.Layout(
		ctx,
		layout.Flexed(1, m.markerListUI(state)),
		layout.Rigid(m.sidePanelUI(state)),
	)

	return nil
}

func (m *MarkersPage) SetupWindow(state abstract.GlobalState) {
	state.Window().Option(
		app.MinSize(800, 600),
		app.Decorated(true),
	)
}
