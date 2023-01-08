package volume

import (
	"math"

	"github.com/g3n/engine/math32"
)

type Box2f struct {
	Min *math32.Vector2
	Max *math32.Vector2
	Box *math32.Box2
}

func AABB2f(corners []*math32.Vector2) Box2f {
	minx, miny := float32(math.MaxFloat32), float32(math.MaxFloat32)
	maxx, maxy := float32(math.SmallestNonzeroFloat32), float32(math.SmallestNonzeroFloat32)
	for _, corner := range corners {

		cx := corner.X
		cy := corner.Y
		if cx < minx {
			minx = cx
		}
		if cx > maxx {
			maxx = cx
		}

		if cy < miny {
			miny = cy
		}

		if cy > maxy {
			maxy = cy
		}
	}

	min := math32.NewVector2(minx, miny)
	max := math32.NewVector2(maxx, maxy)
	return Box2f{
		Min: min,
		Max: max,
		Box: math32.NewBox2(min, max),
	}
}

func (b Box2f) GetWidth() float32 {
	return b.Max.X - b.Min.X
}
func (b Box2f) GetHeigth() float32 {
	return b.Max.Y - b.Min.Y
}
