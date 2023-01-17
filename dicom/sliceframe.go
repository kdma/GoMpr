package volume

import (
	"image"
	"math"

	"github.com/g3n/engine/math32"
)

type SliceFrame struct {
	RotatedFrame   RotatedFrame
	AABB           AABB
	Box2f          Box2f
	Intersections  []math32.Vector3
	Rays           []math32.Ray
	ImageSize      *math32.Vector2
	ImageSizeInMm  *math32.Vector2
	ImagePixelSize *math32.Vector2
	Mpr            **image.RGBA
}

type AABB struct {
	CalibratedCorners []math32.Vector3
	Box               *math32.Box3
}

func filter(vecs []math32.Vector3) []math32.Vector3 {

	var acc []math32.Vector3
	acc = append(acc, vecs[0])
	for i := 0; i < len(vecs); i++ {

		for j := 0; j < len(acc); j++ {
			if vecs[i].Equals(&acc[j]) {
				break
			}
			if j == len(acc)-1 {
				acc = append(acc, vecs[i])
			}
		}
	}
	return acc
}

func Axial(v Volume, slice int) SliceFrame {

	origin := math32.NewVector3(0, 0, float32(slice))
	origin.ApplyMatrix4(v.DcmData.Calibration)

	basis := math32.NewMatrix4()
	zP := math32.NewVector3(0, 0, -1)
	return MakeSliceFrame(zP, origin, basis, v)
}

func Coronal(v Volume, slice int) SliceFrame {

	s := math32.NewVector3(0, float32(slice), 0)
	s.ApplyMatrix4(v.DcmData.Calibration)

	origin := math32.NewVector3(s.X, s.Y, s.Z)
	basis := math32.NewMatrix4().MakeRotationX(math.Pi / 2)

	z := math32.NewVector3(0, -1, 0)

	return MakeSliceFrame(z, origin, basis, v)
}

func Sagittal(v Volume, slice int) SliceFrame {

	basis := math32.NewMatrix4().Multiply(math32.NewMatrix4().MakeRotationY(-math.Pi / 2))

	z := math32.NewVector3(-1, 0, 0)
	s := math32.NewVector3(float32(slice), 0, 0)
	s.ApplyMatrix4(v.DcmData.Calibration)
	origin := math32.NewVector3(s.X, s.Y, s.Z)

	return MakeSliceFrame(z, origin, basis, v)
}
func MakeSliceFrame(zP *math32.Vector3, origin *math32.Vector3, basis *math32.Matrix4, v Volume) SliceFrame {
	aabb := v.GetCorners()
	var intersections []math32.Vector3
	var rays []math32.Ray

	p := math32.NewPlane(zP, origin.Length())
	for _, ray := range getSides(aabb.Box, v) {
		rays = append(rays, *ray)
		pt, err := rp(ray, p, aabb)
		if err == nil {
			intersections = append(intersections, pt)
		}
	}

	intersections = filter(intersections)
	box2f := AABB2f(ToPlaneUV(intersections, zP, origin, basis))

	imgWidth := float32(256)
	boxw := box2f.GetWidth()
	boxh := box2f.GetHeigth()
	pixelSize := boxw / float32(imgWidth)
	imgHeight := boxh / pixelSize
	imageSize := math32.NewVector2(imgWidth, imgHeight)
	imageSizeInMm := math32.NewVector2(boxw, boxh)
	imagePixelSize := math32.NewVector2(pixelSize, pixelSize)
	mpr := &image.RGBA{}
	rotatedFrame := RotatedFrame{basis, origin, p}
	return SliceFrame{
		rotatedFrame,
		aabb,
		box2f,
		intersections,
		rays,
		imageSize,
		imageSizeInMm,
		imagePixelSize,
		&mpr,
	}
}
func FreeRotation(v Volume, basis *math32.Matrix4) SliceFrame {
	var intersections []math32.Vector3
	var rays []math32.Ray

	aabb := v.GetCorners()
	boxCenter := aabb.Box.Center(nil)
	basisOrigin := boxCenter
	z := math32.Vector3{0, 0, -1}
	z.ApplyMatrix4(basis)
	z.Normalize()

	plane := math32.NewPlane(&z, basisOrigin.Length())
	for _, ray := range getSides(aabb.Box, v) {
		rays = append(rays, *ray)
		pt, err := rp(ray, plane, aabb)
		if err == nil {
			intersections = append(intersections, pt)
		}
	}

	box2f := AABB2f(ToPlaneUV(intersections, &z, basisOrigin, basis))

	imgWidth := float32(256)
	boxw := box2f.GetWidth()
	boxh := box2f.GetHeigth()
	pixelSize := boxw / float32(imgWidth)
	imgHeight := boxh / pixelSize
	imageSize := math32.NewVector2(imgWidth, imgHeight)
	imageSizeInMm := math32.NewVector2(boxw, boxh)
	imagePixelSize := math32.NewVector2(pixelSize, pixelSize)
	mpr := &image.RGBA{}
	rotatedFrame := RotatedFrame{basis, basisOrigin, plane}
	return SliceFrame{
		rotatedFrame,
		aabb,
		box2f,
		intersections,
		rays,
		imageSize,
		imageSizeInMm,
		imagePixelSize,
		&mpr,
	}
}

type RotatedFrame struct {
	Basis  *math32.Matrix4
	Origin *math32.Vector3
	Plane  *math32.Plane
}

func (sliceFrame SliceFrame) Cut(v Volume) {
	imgWidth := int(sliceFrame.ImageSize.X)
	imgHeight := int(sliceFrame.ImageSize.Y)
	id := math32.NewMatrix4().Copy(v.DcmData.Calibration)
	calibratedToVoXel := math32.NewMatrix4()
	calibratedToVoXel.GetInverse(id)

	image := make([]byte, imgWidth*imgHeight)

	yDir := math32.NewVector3(0, 1, 0)
	yDir.ApplyMatrix4(sliceFrame.RotatedFrame.Basis)
	yDir.Normalize()
	xDir := math32.NewVector3(1, 0, 0)
	xDir.ApplyMatrix4(sliceFrame.RotatedFrame.Basis)
	xDir.Normalize()

	for x := 0; x < imgWidth; x++ {
		for y := 0; y < imgHeight; y++ {

			fx := float32(x)
			fy := float32(y)
			destX := math32.NewVec3().Copy(xDir).MultiplyScalar(sliceFrame.ImagePixelSize.X * fx)
			destY := math32.NewVec3().Copy(yDir).MultiplyScalar(sliceFrame.ImagePixelSize.Y * fy)
			dcmCoords := math32.NewVec3().Add(destY).Add(destX).Add(sliceFrame.RotatedFrame.Origin)
			dcmCoords.ApplyMatrix4(calibratedToVoXel)

			vX := clamp(dcmCoords.X, 0, v.DcmData.Cols-1)
			vY := clamp(dcmCoords.Y, 0, v.DcmData.Rows-1)
			vZ := clamp(dcmCoords.Z, 0, v.DcmData.Depth-1)

			image[imgWidth*y+x] = v.Data[vZ][vY][vX]
		}
	}
	*sliceFrame.Mpr = Mpr(image, imgWidth, imgHeight, v.DcmData, false)
}
