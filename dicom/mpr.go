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
	fromDataToSlicePlane := math32.NewMatrix4().Multiply(v.DcmData.Calibration)

	centerInMm := math32.NewVector3(float32(v.DcmData.Cols)*.5, float32(v.DcmData.Rows)*.5, float32(v.DcmData.Depth)*.5)
	centerInMm.ApplyMatrix4(fromDataToSlicePlane)

	zeta := math32.NewVector3(0, 0, 1)
	zeta.ApplyMatrix4(fromDataToSlicePlane)
	zeta.Normalize()
	//angle := vec3.Angle(&vec3.T{0, 0, 1}, &zeta)

	//vedi centro di rotazione
	//q := quaternion.FromEulerAngles(math.Pi, 0, 0)
	// t := v.DcmData.Origin.Inverted()
	//delta := vec3.Sub(&centerInMm, &v.DcmData.Origin)
	frame := fromDataToSlicePlane
	sliceFrame, err := AABBIntersections(v, frame)
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

func RenderSlice(frame SliceFrame, volume Volume) {
	imgWidth := 256
	boxw := GetWidth(*frame.Box2f.Box)
	boxh := GetHeigth(*frame.Box2f.Box)
	pixelSize := boxw / float32(imgWidth)
	imgHeight := int(float32(boxh) / pixelSize)
	scale := math32.Vector3{pixelSize, pixelSize, 0}
	translate := math32.Vector3{frame.Box2f.Min.X, frame.Box2f.Min.Y, 0}
	id := math32.NewMatrix4().Copy(volume.DcmData.Calibration)
	dmcToData := math32.NewMatrix4()
	dmcToData.GetInverse(id)

	fromSliceToVoxel := math32.NewMatrix4().Scale(&scale).SetPosition(&translate).Multiply(frame.Frame).Multiply(dmcToData)

	image := make([]byte, imgWidth*imgHeight)

	for i := 0; i < imgWidth; i++ {
		for j := 0; j < imgHeight; j++ {
			fi := float32(i)
			fj := float32(j)
			fz := float32(0.0)

			v := math32.Vector3{fi, fj, fz}
			v.ApplyMatrix4(fromSliceToVoxel)
			// x = floatM[0]*fi + floatM[1]*fj + /* matrix[2] * 0 + */ floatM[3]
			// y = floatM[4]*fi + floatM[5]*fj + /* matrix[6] * 0 + */ floatM[7]
			// z = floatM[8]*fi + floatM[9]*fj + /* matrix[10] * 0 + */ floatM[11]

			vX := clamp(v.X, 0, volume.DcmData.Cols-1)
			vY := clamp(v.Y, 0, volume.DcmData.Rows-1)
			vZ := clamp(v.Z, 0, volume.DcmData.Depth-1)

			image[imgWidth*j+i] = volume.Data[vZ][vY][vX]

		}

	}

	Mpr(image, imgWidth, imgHeight, volume.DcmData)
}

func Mpr(slice []byte, width int, height int, data DcmData) {

	rescaled := make([]byte, len(slice))
	for j := 0; j < len(slice); j++ {
		pixel := float32(int(slice[j])*data.Slope + data.Intercept)

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
