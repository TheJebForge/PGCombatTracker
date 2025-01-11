package drawing

import "github.com/fogleman/gg"

var CommonSpacing float64 = 5

func Empty(ltx Context) Result {
	return Result{
		Draw: func(gg *gg.Context) {},
	}
}
