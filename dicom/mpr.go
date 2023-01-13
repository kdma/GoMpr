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

func (v Volume) Cut(sliceFrame SliceFrame) {
	imgWidth := int(sliceFrame.ImageSize.X)
	imgHeight := int(sliceFrame.ImageSize.Y)
	id := math32.NewMatrix4().Copy(v.DcmData.Calibration)
	calibratedToVoXel := math32.NewMatrix4()
	calibratedToVoXel.GetInverse(id)

	fromSliceToVoxel := math32.NewMatrix4().Multiply(sliceFrame.RotatedFrame.Basis).Multiply(calibratedToVoXel)
	image := make([]byte, imgWidth*imgHeight)

	for x := 0; x < imgWidth; x++ {
		for y := 0; y < imgHeight; y++ {

			planeY := sliceFrame.Box2f.Min.Y + float32(y)*sliceFrame.ImagePixelSize.Y
			planeX := sliceFrame.Box2f.Min.X + float32(x)*sliceFrame.ImagePixelSize.X
			v1 := math32.NewVector3(planeX, planeY, 0)
			v1.Add(sliceFrame.FirstPixelOrigin)
			v1.ApplyMatrix4(fromSliceToVoxel)

			vX := clamp(v1.X, 0, v.DcmData.Cols-1)
			vY := clamp(v1.Y, 0, v.DcmData.Rows-1)
			vZ := clamp(v1.Z, 0, v.DcmData.Depth-1)

			image[imgWidth*y+x] = v.Data[vZ][vY][vX]
		}
	}
	*sliceFrame.Mpr = Mpr(image, imgWidth, imgHeight, v.DcmData, false)
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
