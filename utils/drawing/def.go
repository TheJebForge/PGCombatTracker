package drawing

import (
	"github.com/fogleman/gg"
	"image"
	"math"
)

func F64(x, y float64) F64Point {
	return F64Point{
		X: x,
		Y: y,
	}
}

type F64Point struct {
	X, Y float64
}

func (p F64Point) Round() image.Point {
	return image.Point{
		X: int(math.Ceil(p.X)),
		Y: int(math.Ceil(p.Y)),
	}
}

type Context struct {
	Min F64Point
	Max F64Point
}

type Result struct {
	Size F64Point
	Draw func(gg *gg.Context)
}

// Widget that accepts layouting context, returns size and function that actually draws things
type Widget func(ltx Context) Result

func NewContext(max F64Point) Context {
	return Context{
		Min: F64Point{},
		Max: max,
	}
}
