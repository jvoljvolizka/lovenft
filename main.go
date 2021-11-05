package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/muesli/gamut"
)

const (
	VALUE_CONST uint = 15
	MAXC             = 1<<16 - 1
)

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func c(a uint32) uint8 {
	return uint8((float64(a) / MAXC) * 255)
}

func HalfLifeRight(strikes, length int, img image.Image, seed int64) (out image.Image) {
	s1 := rand.NewSource(seed)
	rnd := rand.New(s1)
	bounds := img.Bounds()
	b := bounds
	output := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(output, bounds, img, bounds.Min, draw.Src)
	inputBounds := b

	for strikes > 0 {

		x := b.Min.X + rnd.Intn(b.Max.X-b.Min.X)
		y := b.Min.Y + rnd.Intn(b.Max.Y-b.Min.Y)
		kc := output.At(x, y)

		var streakEnd int
		if length < 0 {
			streakEnd = inputBounds.Max.X
		} else {
			streakEnd = minInt(x+length, inputBounds.Max.X)
		}

		for x < streakEnd {
			r1, g1, b1, a1 := kc.RGBA()
			r2, g2, b2, a2 := output.At(x, y).RGBA()

			kc = color.RGBA{
				c(r1/4*3 + r2/4),
				c(g1/4*3 + g2/4),
				c(b1/4*3 + b2/4),
				c(a1/4*3 + a2/4),
			}

			output.Set(x, y, kc)
			x++
		}

		strikes--
	}
	return output
}

func HalfLifeLeft(strikes, length int, img image.Image, seed int64) (out image.Image) {
	s1 := rand.NewSource(seed)
	rnd := rand.New(s1)
	bounds := img.Bounds()
	b := bounds
	output := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(output, bounds, img, bounds.Min, draw.Src)
	inputBounds := b

	for strikes > 0 {
		x := b.Min.X + rnd.Intn(b.Max.X-b.Min.X)
		y := b.Min.Y + rnd.Intn(b.Max.Y-b.Min.Y)
		kc := output.At(x, y)

		var streakEnd int
		if length < 0 {
			streakEnd = inputBounds.Min.X
		} else {
			streakEnd = minInt(x-length, inputBounds.Min.X)
		}

		for x >= streakEnd {
			r1, g1, b1, a1 := kc.RGBA()
			r2, g2, b2, a2 := output.At(x, y).RGBA()

			kc = color.RGBA{
				c(r1/4*3 + r2/4),
				c(g1/4*3 + g2/4),
				c(b1/4*3 + b2/4),
				c(a1/4*3 + a2/4),
			}

			output.Set(x, y, kc)
			x--
		}

		strikes--
	}
	return output
}

func reverseAlpha(filename string, name string) {
	pat, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	var img image.Image
	split := strings.Split(filename, ".")
	if split[len(split)-1] == "png" {
		img, err = png.Decode(pat)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		img, err = jpeg.Decode(pat)
		if err != nil {
			log.Fatal(err)
		}
	}

	size := img.Bounds().Size()
	kill := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))
	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			old := img.At(x, y)
			r, g, b, _ := old.RGBA()

			color := color.RGBA{uint8(r), uint8(g), uint8(b), 255 - uint8(r)}
			kill.Set(x, y, color)
		}
	}
	me, _ := os.Create(name + ".png")
	png.Encode(me, kill)
}

func reverseAlphaPattern(filename string, name string) {
	pat, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	var img image.Image
	split := strings.Split(filename, ".")
	if split[len(split)-1] == "png" {
		img, err = png.Decode(pat)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		img, err = jpeg.Decode(pat)
		if err != nil {
			log.Fatal(err)
		}
	}

	size := img.Bounds().Size()
	kill := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))

	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			old := img.At(x, y)
			r, g, b, _ := old.RGBA()
			Y := 0.299*float32(255-r) + 0.587*float32(255-g) + 0.114*float32(255-b)
			var ncolor color.RGBA
			ncolor = color.RGBA{uint8(Y / 256), uint8(Y / 256), uint8(Y / 256), uint8(255 - Y/256)}

			kill.Set(x, y, ncolor)
		}
	}
	me, _ := os.Create(name + ".png")
	png.Encode(me, kill)
}

func create(nmane string, name string, main image.Image, setColor color.Color, length int, reverse int, strikes int) {
	m := image.NewRGBA(image.Rect(0, 0, 700, 700))
	analColors := gamut.Triadic(setColor)

	draw.Draw(m, m.Bounds(), &image.Uniform{setColor}, image.Point{0, 0}, draw.Src)
	f, err := os.Open("pattern-test/" + name)
	if err != nil {
		log.Fatal(err)
	}
	img, err := png.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	src := image.NewRGBA(image.Rect(0, 0, 700, 700))
	draw.Draw(src, src.Bounds(), &image.Uniform{analColors[0]}, image.Point{0, 0}, draw.Src)
	size := img.Bounds().Size()
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	draw.DrawMask(m, m.Bounds(), src, image.Point{0, 0}, img, image.Point{r1.Intn(size.X - 700), r1.Intn(size.Y - 700)}, draw.Over)

	fuck := image.NewRGBA(image.Rect(0, 0, 700, 700))
	draw.Draw(fuck, fuck.Bounds(), &image.Uniform{analColors[1]}, image.Point{0, 0}, draw.Src)

	draw.DrawMask(m, fuck.Bounds(), fuck, image.Point{0, 0}, main.(draw.Image), image.Point{0, 0}, draw.Over)
	var out image.Image
	if reverse > 50 {
		out = HalfLifeRight(strikes, length, m, int64(length))
	} else if reverse < 50 {
		out = HalfLifeLeft(strikes, length, m, int64(length))
	} else {
		out = m
	}

	killme, err := os.Create(nmane + "-" + name + "-anan.png")
	if err != nil {
		log.Fatal(err)
	}
	png.Encode(killme, out)

}

func generate() {
	/*
		1-8 mask
		1-19 pattern
		0-700 halflife strike
		0-3 reverse
		255 r
		255 g
		244 b
		0-700 hl length
		8/19/700/3/255/255/255/700
	*/
	start := time.Now()
	files, _ := ioutil.ReadDir("lovemasks")
	for _, file := range files {
		f, err := os.Open("lovemasks/" + file.Name())
		if err != nil {
			log.Fatal(err)
		}
		img, err := png.Decode(f)
		if err != nil {
			log.Fatal(err)
		}

		files, _ := ioutil.ReadDir("pattern-test")
		for _, f := range files {
			s1 := rand.NewSource(time.Now().UnixNano())
			r1 := rand.New(s1)
			create(file.Name(), f.Name(), img, color.RGBA{uint8(r1.Intn(256)), uint8(r1.Intn(256)), uint8(r1.Intn(256)), 255}, r1.Intn(700), r1.Intn(100), r1.Intn(700))
		}
	}

	/*files, _ = ioutil.ReadDir("lovemasks")
	for _, f := range files {
		reverseAlpha("lovemasks/"+f.Name(), f.Name()+"-boop")
	}*/
	elapsed := time.Since(start)
	fmt.Printf("took %v \n", elapsed)

}

func main() {

	/*

		uint8 mask = 8;
		uint8 pat = 19;
		uint8 rev = 3;
		uint8 r = 255;
		uint8 g = 255;
		uint8 b = 255;
		uint16 strike = 700;
		uint16 len = 700;
	*/
	bigint := &big.Int{}
	bigint.SetString("38129708230729545810620", 10)
	number := math.U256(bigint)
	x := number.Bytes()
	fmt.Println(x)
	fmt.Println(uint(x[0]))
	fmt.Println(uint(x[1]))
	fmt.Println(uint(x[2]))
	fmt.Println(uint(x[3]))
	fmt.Println(uint(x[4]))
	fmt.Println(uint(x[5]))
	mySlice := x[6:8]
	data := binary.BigEndian.Uint16(mySlice)
	fmt.Println(data)
	mySlice = x[8:10]
	data = binary.BigEndian.Uint16(mySlice)
	fmt.Println(data)
	bigint.SetString("4759477275222413344776", 10)
	number = math.U256(bigint)
	x = number.Bytes()
	fmt.Println(x)
	fmt.Println(uint(x[0]))
	fmt.Println(uint(x[1]))
	fmt.Println(uint(x[2]))
	fmt.Println(uint(x[3]))
	fmt.Println(uint(x[4]))
	fmt.Println(uint(x[5]))
	mySlice = x[6:8]
	data = binary.BigEndian.Uint16(mySlice)
	fmt.Println(data)
	mySlice = x[8:10]
	data = binary.BigEndian.Uint16(mySlice)
	fmt.Println(data)
}
