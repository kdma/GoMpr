package volume

import (
	"image"

	"github.com/g3n/engine/math32"
)

type SliceFrame struct {
	RotatedFrame     RotatedFrame
	AABB             AABB
	Box2f            Box2f
	Intersections    []math32.Vector3
	Rays             []math32.Ray
	ImageSize        *math32.Vector2
	ImageSizeInMm    *math32.Vector2
	ImagePixelSize   *math32.Vector2
	FirstPixelOrigin *math32.Vector3
	Mpr              **image.RGBA
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

func FreeRotation(v Volume, basis *math32.Matrix4) SliceFrame {
	var intersections []math32.Vector3
	var rays []math32.Ray

	aabb := v.GetCorners()
	boxCenter := aabb.Box.Center(nil)

	basisOrigin := boxCenter
	z := math32.Vector3{0, 0, -1}
	z.ApplyMatrix4(basis)
	z.Normalize()

	plane := math32.NewPlane(&z, boxCenter.Length())
	for _, ray := range getSides(aabb.Box, v) {
		rays = append(rays, *ray)
		pt, err := rp(ray, plane, aabb)
		if err == nil {
			intersections = append(intersections, pt)
		}
	}

	box2f := AABB2f(ToPlaneUV(intersections, z))

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
		basisOrigin,
		&mpr,
	}
}

type RotatedFrame struct {
	Basis  *math32.Matrix4
	Origin *math32.Vector3
	Plane  *math32.Plane
}
