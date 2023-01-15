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

			fx := float32(x)
			fy := float32(y)
			dcmCoords := math32.NewVector3(fx, fy, 0)
			dcmCoords.ApplyMatrix4(fromSliceToVoxel)

			vX := clamp(dcmCoords.X, 0, v.DcmData.Cols-1)
			vY := clamp(dcmCoords.Y, 0, v.DcmData.Rows-1)
			vZ := clamp(dcmCoords.Z, 0, v.DcmData.Depth-1)

			image[imgWidth*y+x] = v.Data[vZ][vY][vX]
		}
	}
	*sliceFrame.Mpr = Mpr(image, imgWidth, imgHeight, v.DcmData, false)
}

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
