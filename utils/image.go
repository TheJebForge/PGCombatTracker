package utils

import (
	"image"
	"image/color"
	"math"
)

type CheckerImage struct {
	Size image.Point
}

func (c CheckerImage) ColorModel() color.Model {
	return color.NRGBAModel
}

func (c CheckerImage) Bounds() image.Rectangle {
	return image.Rectangle{
		Max: c.Size,
	}
}

func (c CheckerImage) At(x, y int) color.Color {
	x, y = x/2, y/2
	if (x+y)%2 == 1 {
		return LessContrastBg
	} else {
		return SecondBG
	}
}

type PieImage struct {
	Size   int
	Angles []float64
	Colors []color.NRGBA
}

func (p PieImage) ColorModel() color.Model {
	return color.NRGBAModel
}

func (p PieImage) Bounds() image.Rectangle {
	return image.Rectangle{
		Max: image.Point{
			X: p.Size,
			Y: p.Size,
		},
	}
}

func indexOfAngle(angles []float64, target float64) int {
	var currentPosition float64
	for i, angle := range angles {
		if target < currentPosition+angle {
			return i
		}
		currentPosition += angle
	}

	return len(angles) - 1
}

func samplePie(angles []float64, colors []color.NRGBA, screenX, screenY float64) color.NRGBA {
	centeredX, centeredY := screenX*2-1, (1-screenY)*2-1
	magnitude := math.Sqrt(centeredX*centeredX + centeredY*centeredY)
	normalizedX, normalizedY := centeredX/magnitude, centeredY/magnitude

	if magnitude <= 1 {
		angle := math.Atan2(-normalizedX, -normalizedY) + math.Pi
		index := indexOfAngle(angles, angle)

		if index > -1 {
			return colors[index]
		}
	}

	return color.NRGBA{}
}

func (p PieImage) At(x, y int) color.Color {
	floatingSize := float64(p.Size)
	scaledX, scaledY := float64(x)/floatingSize, float64(y)/floatingSize
	pixelUnit := 1 / floatingSize
	subPixel := pixelUnit * 0.25
	subSubPixel := subPixel / 2

	// DirectX like MSAA sampling shit so we get natural anti-aliasing
	return AverageColor(
		samplePie(p.Angles, p.Colors, scaledX+subPixel-subSubPixel, scaledY+subPixel+subSubPixel),
		samplePie(p.Angles, p.Colors, scaledX-subPixel-subSubPixel, scaledY+subPixel-subSubPixel),
		samplePie(p.Angles, p.Colors, scaledX-subPixel+subSubPixel, scaledY-subPixel-subSubPixel),
		samplePie(p.Angles, p.Colors, scaledX+subPixel+subSubPixel, scaledY-subPixel+subSubPixel),
	)
}
