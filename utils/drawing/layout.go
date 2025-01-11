package drawing

import (
	"cmp"
	"gioui.org/layout"
	"github.com/fogleman/gg"
	"github.com/samber/lo"
	"slices"
)

type Flex struct {
	ExpandW       bool
	ExpandH       bool
	ExpandContent bool
	Axis          layout.Axis
	Alignment     layout.Alignment
}

type FlexChild struct {
	Flex   float64
	Rigid  bool
	Widget Widget

	// Scratch space
	space  float64
	result Result
}

func Rigid(widget Widget) FlexChild {
	return FlexChild{
		Rigid:  true,
		Widget: widget,
	}
}

func Flexed(flex float64, widget Widget) FlexChild {
	return FlexChild{
		Flex:   flex,
		Widget: widget,
	}
}

func (c Flex) Layout(children ...FlexChild) Widget {
	return func(ltx Context) Result {
		size := ltx.Min
		if c.ExpandW {
			size.X = ltx.Max.X
		}

		if c.ExpandH {
			size.Y = ltx.Max.Y
		}

		cltx := ltx
		cltx.Min = F64Point{}

		if c.ExpandContent && c.ExpandW {
			cltx.Min.X = ltx.Max.X
		}

		if c.ExpandContent && c.ExpandH {
			cltx.Min.Y = ltx.Max.Y
		}

		// Calculate rigid space and total flex
		var totalRigidSpace float64
		var totalFlex float64

		for i := 0; i < len(children); i++ {
			child := &children[i]
			if child.Rigid {
				child.result = child.Widget(cltx)

				switch c.Axis {
				case layout.Vertical:
					child.space = child.result.Size.Y
					totalRigidSpace += child.space

					if child.result.Size.X > size.X {
						size.X = child.result.Size.X
					}
				case layout.Horizontal:
					child.space = child.result.Size.X
					totalRigidSpace += child.space

					if child.result.Size.Y > size.Y {
						size.Y = child.result.Size.Y
					}
				}
			} else {
				totalFlex += child.Flex
			}
		}

		// Calculate flex children
		var flexSpace float64
		switch c.Axis {
		case layout.Vertical:
			flexSpace = size.Y - totalRigidSpace

			if !c.ExpandH {
				size.Y = totalRigidSpace
			}
		case layout.Horizontal:
			flexSpace = size.X - totalRigidSpace

			if !c.ExpandW {
				size.X = totalRigidSpace
			}
		}

		for i := 0; i < len(children); i++ {
			child := &children[i]
			if child.Rigid {
				continue
			}

			child.space = child.Flex / totalFlex * flexSpace

			cltx := ltx
			switch c.Axis {
			case layout.Vertical:
				cltx.Max.X = size.X
				cltx.Min.Y = child.space
				cltx.Max.Y = child.space
			case layout.Horizontal:
				cltx.Min.X = child.space
				cltx.Max.X = child.space
				cltx.Max.Y = size.Y
			}

			child.result = child.Widget(cltx)
		}

		return Result{
			Size: size,
			Draw: func(gg *gg.Context) {
				gg.Push()

				var currentUnit float64
				for _, child := range children {
					gg.Push()

					var xOffset float64
					var yOffset float64

					switch c.Axis {
					case layout.Vertical:
						yOffset = currentUnit
						switch c.Alignment {
						case layout.Middle:
							xOffset = size.X/2 - child.result.Size.X/2
						case layout.End:
							xOffset = size.X - child.result.Size.X
						default:
						}
					case layout.Horizontal:
						xOffset = currentUnit
						switch c.Alignment {
						case layout.Middle:
							yOffset = size.Y/2 - child.result.Size.Y/2
						case layout.End:
							yOffset = size.Y - child.result.Size.Y
						default:
						}
					}

					gg.Translate(xOffset, yOffset)

					child.result.Draw(gg)
					currentUnit += child.space

					gg.Pop()
				}

				gg.Pop()
			},
		}
	}
}

func FlexHSpacer(width float64) FlexChild {
	return Rigid(HSpacer(width))
}

func HSpacer(width float64) Widget {
	return func(ltx Context) Result {
		return Result{
			Size: F64Point{
				X: width,
				Y: 0,
			},
			Draw: func(gg *gg.Context) {},
		}
	}
}

func FlexVSpacer(height float64) FlexChild {
	return Rigid(VSpacer(height))
}

func VSpacer(height float64) Widget {
	return func(ltx Context) Result {
		return Result{
			Size: F64Point{
				X: 0,
				Y: height,
			},
			Draw: func(gg *gg.Context) {},
		}
	}
}

func Flexer(flex float64) FlexChild {
	return Flexed(flex, Empty)
}

type Inset struct {
	Top, Bottom, Left, Right float64
}

func (i Inset) Layout(inner Widget) Widget {
	return func(ltx Context) Result {
		cltx := ltx
		cltx.Max.X -= i.Left + i.Right
		cltx.Max.Y -= i.Top + i.Bottom
		innerResult := inner(cltx)

		size := F64Point{
			X: innerResult.Size.X + i.Left + i.Right,
			Y: innerResult.Size.Y + i.Top + i.Bottom,
		}

		return Result{
			Size: size,
			Draw: func(gg *gg.Context) {
				gg.Push()
				gg.Translate(i.Left, i.Top)

				innerResult.Draw(gg)

				gg.Pop()
			},
		}
	}
}

func UniformInset(spacing float64) Inset {
	return Inset{
		Top:    spacing,
		Bottom: spacing,
		Left:   spacing,
		Right:  spacing,
	}
}

type Stack struct {
	Wide      bool
	Alignment layout.Direction
}

func (s Stack) Layout(children ...Widget) Widget {
	return func(ltx Context) Result {
		cltx := ltx
		cltx.Min = F64Point{}

		size := ltx.Min

		if s.Wide {
			cltx.Min.X = cltx.Max.X
		}

		childrenResults := make([]Result, len(children))
		for i, child := range children {
			cResult := child(cltx)

			if cResult.Size.X > size.X {
				size.X = cResult.Size.X
			}

			if cResult.Size.Y > size.Y {
				size.Y = cResult.Size.Y
			}

			childrenResults[i] = cResult
		}

		if s.Wide && size.X < cltx.Max.X {
			size.X = cltx.Max.X
		}

		return Result{
			Size: size,
			Draw: func(gg *gg.Context) {
				gg.Push()

				for _, result := range childrenResults {
					var xOffset, yOffset float64

					switch s.Alignment {
					case layout.N:
						xOffset = size.X/2 - result.Size.X/2
					case layout.NE:
						xOffset = size.X - result.Size.X
					case layout.W:
						yOffset = size.Y/2 - result.Size.Y/2
					case layout.Center:
						xOffset = size.X/2 - result.Size.X/2
						yOffset = size.Y/2 - result.Size.Y/2
					case layout.E:
						xOffset = size.X - result.Size.X
						yOffset = size.Y/2 - result.Size.Y/2
					case layout.SW:
						yOffset = size.Y - result.Size.Y
					case layout.S:
						xOffset = size.X/2 - result.Size.X/2
						yOffset = size.Y - result.Size.Y
					case layout.SE:
						xOffset = size.X - result.Size.X
						yOffset = size.Y - result.Size.Y
					default:
					}

					gg.Push()
					gg.Translate(xOffset, yOffset)

					result.Draw(gg)

					gg.Pop()
				}

				gg.Pop()
			},
		}
	}
}

type HorizontalWrap struct {
	Alignment   layout.Alignment
	Spacing     float64
	LineSpacing float64
}

type horizontalLine struct {
	results []Result
	size    F64Point
}

func (hw HorizontalWrap) Layout(children ...Widget) Widget {
	return func(ltx Context) Result {
		maxWidth := ltx.Max.X

		var lines []horizontalLine
		var currentLine horizontalLine

		cltx := ltx
		cltx.Min = F64Point{}

		for _, c := range children {
			cResult := c(cltx)

			currentSpacing := hw.Spacing
			if len(currentLine.results) <= 0 {
				currentSpacing = 0
			}

			if currentLine.size.X+currentSpacing+cResult.Size.X < maxWidth || len(currentLine.results) <= 0 {
				currentLine.size.X += currentSpacing + cResult.Size.X
				currentLine.size.Y = max(currentLine.size.Y, cResult.Size.Y)

				currentLine.results = append(currentLine.results, cResult)
			} else {
				lines = append(lines, currentLine)

				currentLine = horizontalLine{
					size: cResult.Size,
					results: []Result{
						cResult,
					},
				}
			}
		}

		lines = append(lines, currentLine)

		maxSize := F64Point{
			X: slices.MaxFunc(lines, func(a, b horizontalLine) int {
				return cmp.Compare(a.size.X, b.size.X)
			}).size.X,
			Y: lo.SumBy(lines, func(item horizontalLine) float64 {
				return item.size.Y
			}) + (float64(len(lines)-1) * hw.LineSpacing),
		}

		return Result{
			Size: maxSize,
			Draw: func(gg *gg.Context) {
				gg.Push()

				var currentY float64
				for _, line := range lines {
					var currentX float64

					for _, result := range line.results {
						offset := F64(currentX, currentY)

						switch hw.Alignment {
						case layout.Middle:
							offset.Y += line.size.Y/2 - result.Size.Y/2
						case layout.End:
							offset.Y += line.size.Y - result.Size.Y
						default:
						}

						gg.Push()
						gg.Translate(offset.X, offset.Y)

						result.Draw(gg)

						gg.Pop()

						currentX += hw.Spacing + result.Size.X
					}

					currentY += line.size.Y + hw.LineSpacing
				}

				gg.Pop()
			},
		}
	}
}
