package drawing

import (
	"github.com/fogleman/gg"
	"image/color"
)

func Background(bg Widget, inner Widget) Widget {
	return func(ltx Context) Result {
		innerResult := inner(ltx)

		cltx := ltx
		cltx.Min = innerResult.Size
		cltx.Max = innerResult.Size
		bgResult := bg(cltx)

		return Result{
			Size: innerResult.Size,
			Draw: func(gg *gg.Context) {
				gg.Push()

				bgResult.Draw(gg)
				innerResult.Draw(gg)

				gg.Pop()
			},
		}
	}
}

func RoundedBackground(color color.NRGBA, radius float64) Widget {
	return func(ltx Context) Result {
		size := ltx.Min

		return Result{
			Size: size,
			Draw: func(gg *gg.Context) {
				gg.Push()

				gg.SetColor(color)
				gg.DrawRoundedRectangle(0, 0, size.X, size.Y, radius)
				gg.Fill()

				gg.Pop()
			},
		}
	}
}

func CustomSurface(background, inner Widget) Widget {
	return Background(
		background,
		UniformInset(CommonSpacing*2).Layout(inner),
	)
}

func RoundedSurface(color color.NRGBA, inner Widget) Widget {
	return CustomSurface(
		RoundedBackground(color, CommonSpacing*2),
		inner,
	)
}
