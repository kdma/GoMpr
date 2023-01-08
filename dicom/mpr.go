package volume

import (
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"

	"github.com/g3n/engine/math32"
)

func (v Volume) Cut() (SliceFrame, error) {

	sliceFrame, err := AABBIntersections(v)
	if err != nil {
		return SliceFrame{}, error(err)
	}
	RenderSlice(sliceFrame, v)
	return sliceFrame, nil
}

func clamp(f float32, min int, max int) int {

	fInt := int(f)
	if fInt < min {
		return 0
	}
	if fInt > max {
		return max
	}
	return fInt
}

func RenderSlice(sliceFrame SliceFrame, volume Volume) {
	imgWidth := int(sliceFrame.ImageSize.X)
	imgHeight := int(sliceFrame.ImageSize.Y)
	id := math32.NewMatrix4().Copy(volume.DcmData.Calibration).SetPosition(math32.NewVec3())
	calibratedToVoXel := math32.NewMatrix4()
	calibratedToVoXel.GetInverse(id)

	//trans := math32.NewVector3(-volume.DcmData.Origin.X, -volume.DcmData.Origin.Y, -volume.DcmData.Origin.Z)
	fromSliceToVoxel := math32.NewMatrix4().Multiply(sliceFrame.Basis).Multiply(calibratedToVoXel)
	image := make([]byte, imgWidth*imgHeight)

	for x := 0; x < imgWidth; x++ {
		for y := 0; y < imgHeight; y++ {

			fx := float32(x)
			fy := float32(y)

			v := math32.NewVector3(fx, fy, 0)
			v.ApplyMatrix4(fromSliceToVoxel)
			vX := clamp(v.X, 0, volume.DcmData.Cols-1)
			vY := clamp(v.Y, 0, volume.DcmData.Rows-1)
			vZ := clamp(v.Z, 0, volume.DcmData.Depth-1)

			image[imgWidth*y+x] = volume.Data[vZ][vY][vX]

		}

	}

	Mpr(image, imgWidth, imgHeight, volume.DcmData)
}

func Mpr(slice []byte, width int, height int, data DcmData) {

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

	f, err := os.Create(filepath.Join("C:\\Users\\franc\\Desktop\\Nuova Cartella", "mpr.jpg"))
	if err != nil {
		log.Fatal(err)
	}
	_ = jpeg.Encode(f, img, &jpeg.Options{Quality: 100})
	_ = f.Close()
}
