package volume

import (
	"github.com/ungerik/go3d/mat4"
	"github.com/ungerik/go3d/vec3"
)

type SliceFrame struct {
	Frame mat4.T
	Plane Plane
	AABB  AABB
}

type AABB struct {
	CalibratedCorners []vec3.T
	Box               vec3.Box
}

func getCorners(volume Volume) AABB {

	box := vec3.Box{
		Min: vec3.Zero,
		Max: vec3.T{float32(volume.DcmData.Cols), float32(volume.DcmData.Rows), float32(volume.DcmData.Depth)},
	}

	var corners []vec3.T
	center := box.Center()
	corners = append(corners, vec3.Add(&center, &vec3.T{GetX(box.Min), GetY(box.Min), GetZ(box.Min)}))
	corners = append(corners, vec3.Add(&center, &vec3.T{GetX(box.Min), GetY(box.Max), GetZ(box.Min)}))
	corners = append(corners, vec3.Add(&center, &vec3.T{GetX(box.Max), GetY(box.Min), GetZ(box.Min)}))
	corners = append(corners, vec3.Add(&center, &vec3.T{GetX(box.Max), GetY(box.Max), GetZ(box.Min)}))
	corners = append(corners, vec3.Add(&center, &vec3.T{GetX(box.Min), GetY(box.Min), GetZ(box.Max)}))
	corners = append(corners, vec3.Add(&center, &vec3.T{GetX(box.Min), GetY(box.Max), GetZ(box.Max)}))
	corners = append(corners, vec3.Add(&center, &vec3.T{GetX(box.Max), GetY(box.Min), GetZ(box.Max)}))
	corners = append(corners, vec3.Add(&center, &vec3.T{GetX(box.Max), GetY(box.Max), GetZ(box.Max)}))

	for i := 0; i < len(corners); i++ {
		corners[i] = volume.DcmData.Calibration.MulVec3(&corners[i])
	}
	return AABB{
		CalibratedCorners: corners,
		Box:               box,
	}
}

func GetX(v vec3.T) float32 {
	return v.Get(0, 0)
}
func GetY(v vec3.T) float32 {
	return v.Get(0, 1)
}
func GetZ(v vec3.T) float32 {
	return v.Get(0, 2)
}
