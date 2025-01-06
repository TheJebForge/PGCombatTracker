package ui

import (
	"PGCombatTracker/abstract"
	"PGCombatTracker/utils"
	"fmt"
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"log"
	"reflect"
	"strconv"
	"strings"
)

func NewSettingsPage() *SettingsPage {
	backIcon, err := widget.NewIcon(icons.NavigationArrowBack)
	if err != nil {
		log.Fatalln(err)
	}

	addIcon, err := widget.NewIcon(icons.ContentAdd)
	if err != nil {
		log.Fatalln(err)
	}

	deleteIcon, err := widget.NewIcon(icons.ActionDelete)
	if err != nil {
		log.Fatalln(err)
	}

	return &SettingsPage{
		backIcon:   backIcon,
		backButton: &widget.Clickable{},
		addIcon:    addIcon,
		deleteIcon: deleteIcon,
		settingsList: &widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
	}
}

type SettingsPage struct {
	backIcon        *widget.Icon
	backButton      *widget.Clickable
	addIcon         *widget.Icon
	deleteIcon      *widget.Icon
	settingsList    *widget.List
	initialized     bool
	settingsWidgets []layout.Widget
}

func smallerIconButton(theme *material.Theme, button *widget.Clickable, icon *widget.Icon, desc string) material.IconButtonStyle {
	style := material.IconButton(theme, button, icon, desc)
	style.Inset = layout.UniformInset(utils.CommonSpacing)
	return style
}

func styledEditor(theme *material.Theme, editor *widget.Editor, hint string) material.EditorStyle {
	style := material.Editor(theme, editor, hint)
	style.HintColor = utils.GrayText
	return style
}

func editorLine(gtx layout.Context, theme *material.Theme, name string, editor layout.Widget) layout.Dimensions {
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return utils.LayoutMinimalX(
				gtx.Dp(35),
				material.Body2(theme, fmt.Sprintf("%v:", name)).Layout(gtx),
			)
		}),
		utils.FlexSpacerW(utils.CommonSpacing),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Background{}.Layout(
				gtx,
				utils.MakeRoundedBG(10, utils.LesserContrastBg),
				func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(utils.CommonSpacing*2).Layout(
						gtx,
						editor,
					)
				},
			)
		}),
	)

}

func (s *SettingsPage) reflectField(state abstract.GlobalState, field reflect.Value, name string) layout.Widget {
	normalizedName := strings.Join(utils.PascalSplit(name), " ")

	switch field.Type().Kind() {
	case reflect.Bool:
		checkbox := &widget.Bool{
			Value: field.Bool(),
		}

		return func(gtx layout.Context) layout.Dimensions {
			if checkbox.Update(gtx) {
				field.SetBool(checkbox.Value)
				state.SaveSettings()
			}

			return material.CheckBox(state.Theme(), checkbox, normalizedName).Layout(gtx)
		}
	case reflect.String:
		editor := &widget.Editor{
			SingleLine: true,
		}
		editor.SetText(field.String())

		return func(gtx layout.Context) layout.Dimensions {
			if _, ok := editor.Update(gtx); ok {
				field.SetString(editor.Text())
				state.SaveSettings()
			}

			return editorLine(
				gtx,
				state.Theme(),
				normalizedName,
				styledEditor(state.Theme(), editor, "Empty").Layout,
			)
		}
	case reflect.Int:
		editor := &widget.Editor{
			SingleLine: true,
		}
		editor.SetText(strconv.Itoa(int(field.Int())))
		invalid := false

		return func(gtx layout.Context) layout.Dimensions {
			if _, ok := editor.Update(gtx); ok {
				num, err := strconv.Atoi(editor.Text())

				if err != nil {
					invalid = true
				} else {
					invalid = false

					field.SetInt(int64(num))
					state.SaveSettings()
				}
			}

			return editorLine(
				gtx,
				state.Theme(),
				normalizedName,
				func(gtx layout.Context) layout.Dimensions {
					style := styledEditor(state.Theme(), editor, "Empty")

					if invalid {
						style.Color = utils.RedText
					} else {
						style.Color = state.Theme().Fg
					}

					return style.Layout(gtx)
				},
			)
		}
	case reflect.Uint:
		editor := &widget.Editor{
			SingleLine: true,
		}
		editor.SetText(strconv.FormatUint(field.Uint(), 10))
		invalid := false

		return func(gtx layout.Context) layout.Dimensions {
			if _, ok := editor.Update(gtx); ok {
				num, err := strconv.ParseUint(editor.Text(), 10, 64)

				if err != nil {
					invalid = true
				} else {
					invalid = false

					field.SetUint(num)
					state.SaveSettings()
				}
			}

			return editorLine(
				gtx,
				state.Theme(),
				normalizedName,
				func(gtx layout.Context) layout.Dimensions {
					style := styledEditor(state.Theme(), editor, "Empty")

					if invalid {
						style.Color = utils.RedText
					} else {
						style.Color = state.Theme().Fg
					}

					return style.Layout(gtx)
				},
			)
		}
	case reflect.Float64:
		editor := &widget.Editor{
			SingleLine: true,
		}
		editor.SetText(strconv.FormatFloat(field.Float(), 'f', -1, 64))
		invalid := false

		return func(gtx layout.Context) layout.Dimensions {
			if _, ok := editor.Update(gtx); ok {
				num, err := strconv.ParseFloat(editor.Text(), 64)

				if err != nil {
					invalid = true
				} else {
					invalid = false

					field.SetFloat(num)
					state.SaveSettings()
				}
			}

			return editorLine(
				gtx,
				state.Theme(),
				normalizedName,
				func(gtx layout.Context) layout.Dimensions {
					style := styledEditor(state.Theme(), editor, "Empty")

					if invalid {
						style.Color = utils.RedText
					} else {
						style.Color = state.Theme().Fg
					}

					return style.Layout(gtx)
				},
			)
		}
	case reflect.Slice:
		list := &widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		}
		addButton := &widget.Clickable{}

		var widgets []layout.Widget

		var refreshChildren func()

		refreshChildren = func() {
			widgets = nil
			childrenAmount := field.Len()
			for i := 0; i < childrenAmount; i++ {
				childField := field.Index(i)
				innerWidget := s.reflectField(state, childField, strconv.Itoa(i))
				deleteButton := &widget.Clickable{}

				widgets = append(widgets, func(gtx layout.Context) layout.Dimensions {
					line := func(gtx layout.Context) layout.Dimensions {
						if deleteButton.Clicked(gtx) {
							fieldLen := field.Len()
							newSlice := reflect.MakeSlice(field.Type(), fieldLen-1, fieldLen-1)

							newIndex := 0
							for j := 0; j < field.Len(); j++ {
								if j != i {
									newSlice.Index(newIndex).Set(field.Index(j))
									newIndex++
								}
							}

							field.Set(newSlice)
							state.SaveSettings()
							refreshChildren()
						}

						return layout.Flex{
							Axis:      layout.Horizontal,
							Alignment: layout.Middle,
						}.Layout(
							gtx,
							layout.Flexed(1, innerWidget),
							utils.FlexSpacerW(utils.CommonSpacing*2),
							layout.Rigid(smallerIconButton(state.Theme(), deleteButton, s.deleteIcon, "Delete").Layout),
						)
					}

					if i+1 == childrenAmount {
						return line(gtx)
					} else {
						return layout.Flex{
							Axis: layout.Vertical,
						}.Layout(
							gtx,
							layout.Rigid(line),
							utils.FlexSpacerH(utils.CommonSpacing),
						)
					}
				})
			}
		}

		refreshChildren()

		return func(gtx layout.Context) layout.Dimensions {
			if addButton.Clicked(gtx) {
				newField := reflect.Append(field, reflect.Zero(field.Type().Elem()))
				field.Set(newField)
				state.SaveSettings()
				refreshChildren()
			}

			return layout.Inset{
				Left: utils.CommonSpacing, Right: utils.CommonSpacing,
			}.Layout(
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
								utils.FlexSpacerW(utils.CommonSpacing),
								layout.Flexed(1, material.Body2(state.Theme(), normalizedName).Layout),
								layout.Rigid(smallerIconButton(state.Theme(), addButton, s.addIcon, "Add").Layout),
							)
						}),
						utils.FlexSpacerH(utils.CommonSpacing),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Background{}.Layout(
								gtx,
								utils.MakeRoundedBG(10, utils.SecondBG),
								func(gtx layout.Context) layout.Dimensions {
									return layout.UniformInset(utils.CommonSpacing*2).Layout(
										gtx,
										func(gtx layout.Context) layout.Dimensions {
											return material.List(state.Theme(), list).Layout(
												gtx,
												len(widgets),
												func(gtx layout.Context, index int) layout.Dimensions {
													if index < len(widgets) {
														last := index+1 == len(widgets)

														encasedWidget := func(gtx layout.Context) layout.Dimensions {
															return layout.Background{}.Layout(
																gtx,
																utils.MakeRoundedBG(10, utils.BG),
																func(gtx layout.Context) layout.Dimensions {
																	bottom := utils.CommonSpacing
																	if last {
																		bottom = utils.CommonSpacing * 2
																	}

																	return layout.Inset{
																		Top:    utils.CommonSpacing * 2,
																		Left:   utils.CommonSpacing * 2,
																		Right:  utils.CommonSpacing * 2,
																		Bottom: bottom,
																	}.Layout(
																		gtx,
																		widgets[index],
																	)
																},
															)
														}

														if last {
															return encasedWidget(gtx)
														} else {
															return layout.Flex{
																Axis: layout.Vertical,
															}.Layout(
																gtx,
																layout.Rigid(encasedWidget),
																utils.FlexSpacerH(utils.CommonSpacing),
															)
														}
													} else {
														return layout.Dimensions{}
													}
												},
											)
										},
									)
								},
							)
						}),
					)
				},
			)
		}
	default:
		return nil
	}
}

func (s *SettingsPage) reflect(state abstract.GlobalState) []layout.Widget {
	settings := state.Settings()

	reflected := reflect.ValueOf(settings).Elem()
	reflectedType := reflected.Type()

	var widgets []layout.Widget

	for i := 0; i < reflected.NumField(); i++ {
		field := reflected.Field(i)
		name := reflectedType.Field(i).Name

		if result := s.reflectField(state, field, name); result != nil {
			widgets = append(widgets, result)
		}
	}

	return widgets
}

func (s *SettingsPage) navBar(state abstract.GlobalState) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if s.backButton.Clicked(gtx) {
			state.SwitchPage(NewFileSelectionPage())
		}

		return layout.Background{}.Layout(
			gtx,
			utils.MakeColoredBG(utils.SecondBG),
			func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(utils.CommonSpacing).Layout(
					gtx,
					func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{
							Axis:      layout.Horizontal,
							Alignment: layout.Middle,
						}.Layout(
							gtx,
							layout.Rigid(material.IconButton(state.Theme(), s.backButton, s.backIcon, "Back").Layout),
							utils.FlexSpacerW(utils.CommonSpacing*2),
							layout.Rigid(material.H4(state.Theme(), "Settings").Layout),
						)
					},
				)
			},
		)
	}
}

func (s *SettingsPage) body(state abstract.GlobalState) layout.Widget {
	if !s.initialized {
		s.initialized = true
		s.settingsWidgets = s.reflect(state)
	}

	return func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(utils.CommonSpacing).Layout(
			gtx,
			func(gtx layout.Context) layout.Dimensions {
				return material.List(state.Theme(), s.settingsList).Layout(
					gtx,
					len(s.settingsWidgets),
					func(gtx layout.Context, index int) layout.Dimensions {
						if index+1 == len(s.settingsWidgets) {
							return s.settingsWidgets[index](gtx)
						} else {
							return layout.Flex{
								Axis: layout.Vertical,
							}.Layout(
								gtx,
								layout.Rigid(s.settingsWidgets[index]),
								utils.FlexSpacerH(utils.CommonSpacing*2),
							)
						}
					},
				)
			},
		)
	}
}

func (s *SettingsPage) Layout(ctx layout.Context, state abstract.GlobalState) error {
	layout.Flex{
		Axis: layout.Vertical,
	}.Layout(
		ctx,
		layout.Rigid(s.navBar(state)),
		layout.Flexed(1, s.body(state)),
	)

	return nil
}

func (s *SettingsPage) SetupWindow(state abstract.GlobalState) {
	state.Window().Option(
		app.MinSize(800, 600),
		app.Decorated(true),
	)
}
