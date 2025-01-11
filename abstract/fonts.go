package abstract

import (
	_ "embed"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

//go:embed fonts/OpenSans-Regular.ttf
var openSans []byte

func LoadFontPack() (*FontPack, error) {
	regularFont, err := truetype.Parse(openSans)
	if err != nil {
		return nil, err
	}

	return &FontPack{
		Heading:  truetype.NewFace(regularFont, &truetype.Options{Size: 48}),
		Body:     truetype.NewFace(regularFont, &truetype.Options{Size: 24}),
		Smaller:  truetype.NewFace(regularFont, &truetype.Options{Size: 18}),
		Smallest: truetype.NewFace(regularFont, &truetype.Options{Size: 14}),
	}, nil
}

type FontPack struct {
	Heading  font.Face
	Body     font.Face
	Smaller  font.Face
	Smallest font.Face
}
