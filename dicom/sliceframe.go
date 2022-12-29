package volume

import (
	"github.com/g3n/engine/math32"
)

type SliceFrame struct {
	Frame   *math32.Matrix4
	Plane   *math32.Plane
	AABB    AABB
	Box2f   Box2f
	Corners []math32.Vector3
}

type AABB struct {
	CalibratedCorners []math32.Vector3
	Box               *math32.Box3
}

func AABBIntersections(v Volume, frame *math32.Matrix4) (SliceFrame, error) {
	var intersections []math32.Vector3
	origin := math32.Vector3{0, 0, 0}
	z := math32.Vector3{0, 0, 1}
	origin.ApplyMatrix4(frame)
	z.ApplyMatrix4(frame)

	z.Normalize()
	plane := math32.NewPlane(&z, origin.Length())
	corners := v.GetCorners()
	for _, ray := range getSides(corners.Box, plane) {
		pt, err := rp(ray, plane)
		if err == nil {
			intersections = append(intersections, pt)
		}
	}

	box2f := AABB2f(ToPlaneUVBatch(intersections, z))

	return SliceFrame{
		frame,
		plane,
		corners,
		box2f,
		intersections,
	}, nil
}
