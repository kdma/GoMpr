package volume

import (
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/g3n/engine/math32"
)

func GetM(fromSliceToVoxel *math32.Matrix4) []float32 {
	m := make([]float32, 16)
	for i := 0; i < 4; i++ {
		v := fromSliceToVoxel.GetColumnVector3(i)
		for j := 0; j < 3; j++ {
			m[i*4+j] = v.Component(j)
		}
	}

	return m
}

func clamp(f float32, min int, max int) int {

	fInt := int(math.Round(float64(f)))
	if fInt < min {
		return 0
	}
	if fInt > max {
		return max
	}
	return fInt
}

func Mpr(slice []byte, width int, height int, data DcmData, debug bool) *image.RGBA {

	rescaled := make([]byte, len(slice))
	for j := 0; j < len(slice); j++ {
		pixel := float32(int(slice[j]))*data.Slope + data.Intercept

		if pixel <= float32(data.Window)-0.5-(float32(data.Level-1)/2) {
			pixel = 0
		} else if pixel > float32(data.Window)-0.5+float32(data.Level-1)/2 {
			pixel = 255
		} else {
			pixel = ((pixel-(float32(data.Window)-0.5))/float32(data.Level-1) + 0.5) * float32(255)
		}
		rescaled[j] = byte(pixel)
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for c := 0; c < width; c++ {
		for r := 0; r < height; r++ {
			p := slice[r*width+c]
			img.SetRGBA(c, r, color.RGBA{A: 0xFF, R: p, G: p, B: p})
		}
	}

	if debug {
		f, err := os.Create(filepath.Join("C:\\Users\\franc\\Desktop\\Nuova Cartella", "mpr.jpg"))
		if err != nil {
			log.Fatal(err)
		}
		_ = jpeg.Encode(f, img, &jpeg.Options{Quality: 100})
		_ = f.Close()
	}
	return img
}
