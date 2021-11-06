package imageops

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/muesli/gamut"
)

type NftImage struct {
	Image           image.Image
	MaskSelector    uint8
	PatternSelector uint8
	reverseSelector uint8
	Color           color.Color
	strikeCount     uint16
	hlLength        uint16
}

const (
	VALUE_CONST uint = 15
	MAXC             = 1<<16 - 1
)

func (NFT *NftImage) minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func (NFT *NftImage) c(a uint32) uint8 {
	return uint8((float64(a) / MAXC) * 255)
}

func (NFT *NftImage) HalfLifeRight(seed int64) {
	length := int(NFT.hlLength)
	strikes := int(NFT.strikeCount)
	s1 := rand.NewSource(seed)
	rnd := rand.New(s1)
	bounds := NFT.Image.Bounds()
	b := bounds
	output := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(output, bounds, NFT.Image, bounds.Min, draw.Src)
	inputBounds := b

	for strikes > 0 {

		x := b.Min.X + rnd.Intn(b.Max.X-b.Min.X)
		y := b.Min.Y + rnd.Intn(b.Max.Y-b.Min.Y)
		kc := output.At(x, y)

		var streakEnd int
		if length < 0 {
			streakEnd = inputBounds.Max.X
		} else {
			streakEnd = NFT.minInt(x+length, inputBounds.Max.X)
		}

		for x < streakEnd {
			r1, g1, b1, a1 := kc.RGBA()
			r2, g2, b2, a2 := output.At(x, y).RGBA()

			kc = color.RGBA{
				NFT.c(r1/4*3 + r2/4),
				NFT.c(g1/4*3 + g2/4),
				NFT.c(b1/4*3 + b2/4),
				NFT.c(a1/4*3 + a2/4),
			}

			output.Set(x, y, kc)
			x++
		}

		strikes--
	}
	NFT.Image = output
}

func (NFT *NftImage) HalfLifeLeft(seed int64) {
	length := int(NFT.hlLength)
	strikes := int(NFT.strikeCount)
	s1 := rand.NewSource(seed)
	rnd := rand.New(s1)
	bounds := NFT.Image.Bounds()
	b := bounds
	output := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(output, bounds, NFT.Image, bounds.Min, draw.Src)
	inputBounds := b

	for strikes > 0 {
		x := b.Min.X + rnd.Intn(b.Max.X-b.Min.X)
		y := b.Min.Y + rnd.Intn(b.Max.Y-b.Min.Y)
		kc := output.At(x, y)

		var streakEnd int
		if length < 0 {
			streakEnd = inputBounds.Min.X
		} else {
			streakEnd = NFT.minInt(x-length, inputBounds.Min.X)
		}

		for x >= streakEnd {
			r1, g1, b1, a1 := kc.RGBA()
			r2, g2, b2, a2 := output.At(x, y).RGBA()

			kc = color.RGBA{
				NFT.c(r1/4*3 + r2/4),
				NFT.c(g1/4*3 + g2/4),
				NFT.c(b1/4*3 + b2/4),
				NFT.c(a1/4*3 + a2/4),
			}

			output.Set(x, y, kc)
			x--
		}

		strikes--
	}
	NFT.Image = output
}

func (NFT *NftImage) Create(heartMask image.Image, pattern image.Image) {

	length := int(NFT.hlLength)
	reverse := int(NFT.reverseSelector)
	setColor := NFT.Color

	m := image.NewRGBA(image.Rect(0, 0, 700, 700))
	analColors := gamut.Triadic(setColor)
	fmt.Println(analColors)
	draw.Draw(m, m.Bounds(), &image.Uniform{setColor}, image.Point{0, 0}, draw.Src)

	// new empty image
	src := image.NewRGBA(image.Rect(0, 0, 700, 700))
	draw.Draw(src, src.Bounds(), &image.Uniform{analColors[0]}, image.Point{0, 0}, draw.Src)

	size := pattern.Bounds().Size()
	// use half life length as seed
	s1 := rand.NewSource(int64(length))
	r1 := rand.New(s1)
	//get a background with patterns
	draw.DrawMask(m, m.Bounds(), src, image.Point{0, 0}, pattern, image.Point{r1.Intn(size.X - 700), r1.Intn(size.Y - 700)}, draw.Over)

	//create a mask with heart symbol
	lastmask := image.NewRGBA(image.Rect(0, 0, 700, 700))
	draw.Draw(lastmask, lastmask.Bounds(), &image.Uniform{analColors[1]}, image.Point{0, 0}, draw.Src)

	//create last image
	draw.DrawMask(m, lastmask.Bounds(), lastmask, image.Point{0, 0}, heartMask.(draw.Image), image.Point{0, 0}, draw.Over)
	NFT.Image = m
	//apply halflife

	//usinf length as seed for now
	if reverse == 1 {
		NFT.HalfLifeRight(int64(length))
	}
	if reverse == 2 {
		NFT.HalfLifeLeft(int64(length))
	}
}

func NewImage(tokenid string) (NftImage, error) {

	bigint := &big.Int{}
	_, s := bigint.SetString(tokenid, 10)

	if !s {
		err := fmt.Errorf("invalid tokenid:")
		return NftImage{}, err
	}
	number := math.U256(bigint)
	x := make([]byte, 10)
	number.FillBytes(x)
	fmt.Println(x)
	newNFT := NftImage{
		MaskSelector:    uint8(x[0]),
		PatternSelector: uint8(x[1]),
		reverseSelector: uint8(x[2]),
		Color:           color.RGBA{uint8(x[3]), uint8(x[4]), uint8(x[5]), 255},
	}

	mySlice := x[6:8]
	newNFT.strikeCount = binary.BigEndian.Uint16(mySlice)
	mySlice = x[8:10]
	newNFT.hlLength = binary.BigEndian.Uint16(mySlice)
	return newNFT, nil
}
