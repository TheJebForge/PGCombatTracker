package components

import (
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"hash/fnv"
	"image"
	"image/color"
	"math"
	"math/rand"
)

func seedHash(txt string) int64 {
	h := fnv.New64()
	_, _ = h.Write([]byte(txt))
	sum := h.Sum64()
	return int64(sum)
}

func StringToColor(txt string) color.NRGBA {
	const low = 60
	const ran = 50

	r := rand.New(rand.NewSource(seedHash(txt)))

	return color.NRGBA{
		R: uint8(low + r.Intn(ran)),
		G: uint8(low + r.Intn(ran)),
		B: uint8(low + r.Intn(ran)),
		A: 255,
	}
}

func BarWidget(color color.NRGBA, height unit.Dp, progress float64) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		maxWidth := gtx.Constraints.Max.X
		width := int(math.Floor(float64(maxWidth) * progress))
		height := gtx.Dp(height)

		size := image.Point{
			X: width,
			Y: height,
		}
		maxSize := image.Point{
			X: maxWidth,
			Y: height,
		}

		defer clip.UniformRRect(image.Rectangle{
			Max: size,
		}, 5).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, color)

		return layout.Dimensions{
			Size: maxSize,
		}
	}
}
