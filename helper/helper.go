package helper

import (
	"image"
	"image/color"
)

func reverseAlpha(img image.Image) image.Image {

	size := img.Bounds().Size()
	out := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))
	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			old := img.At(x, y)
			r, g, b, _ := old.RGBA()

			color := color.RGBA{uint8(r), uint8(g), uint8(b), 255 - uint8(r)}
			out.Set(x, y, color)
		}
	}
	return out
}

func reverseAlphaPattern(img image.Image) image.Image {
	size := img.Bounds().Size()
	out := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))

	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			old := img.At(x, y)
			r, g, b, _ := old.RGBA()
			Y := 0.299*float32(255-r) + 0.587*float32(255-g) + 0.114*float32(255-b)
			var ncolor color.RGBA
			ncolor = color.RGBA{uint8(Y / 256), uint8(Y / 256), uint8(Y / 256), uint8(255 - Y/256)}

			out.Set(x, y, ncolor)
		}
	}
	return out
}
