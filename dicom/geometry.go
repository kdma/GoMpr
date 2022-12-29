package volume

import (
	"errors"

	"github.com/g3n/engine/math32"
)

func getSides(box *math32.Box3, plane *math32.Plane) []*math32.Ray {

	var edges []*math32.Ray
	edges = append(edges, math32.NewRay(&box.Min, math32.NewVector3(box.Max.X-box.Min.X, 0, 0)))
	edges = append(edges, math32.NewRay(math32.NewVector3(box.Min.X, box.Max.Y, box.Min.Z), math32.NewVector3(box.Max.X-box.Min.X, 0, 0).Normalize()))
	edges = append(edges, math32.NewRay(math32.NewVector3(box.Min.X, box.Max.Y, box.Min.Z), math32.NewVector3(box.Max.X-box.Min.X, 0, 0).Normalize()))
	edges = append(edges, math32.NewRay(math32.NewVector3(box.Min.X, box.Max.Y, box.Max.Z), math32.NewVector3(box.Max.X-box.Min.X, 0, 0).Normalize()))

	edges = append(edges, math32.NewRay(math32.NewVector3(box.Min.X, box.Min.Y, box.Min.Z), math32.NewVector3(0, box.Max.Y-box.Min.Y, 0).Normalize()))
	edges = append(edges, math32.NewRay(math32.NewVector3(box.Max.X, box.Min.Y, box.Min.Z), math32.NewVector3(0, box.Max.Y-box.Min.Y, 0).Normalize()))
	edges = append(edges, math32.NewRay(math32.NewVector3(box.Min.X, box.Min.Y, box.Max.Z), math32.NewVector3(0, box.Max.Y-box.Min.Y, 0).Normalize()))
	edges = append(edges, math32.NewRay(math32.NewVector3(box.Max.X, box.Min.Y, box.Max.Z), math32.NewVector3(0, box.Max.Y-box.Min.Y, 0).Normalize()))

	edges = append(edges, math32.NewRay(math32.NewVector3(box.Min.X, box.Min.Y, box.Min.Z), math32.NewVector3(0, 0, box.Max.Z-box.Min.Z).Normalize()))
	edges = append(edges, math32.NewRay(math32.NewVector3(box.Max.X, box.Min.Y, box.Min.Z), math32.NewVector3(0, 0, box.Max.Z-box.Min.Z).Normalize()))
	edges = append(edges, math32.NewRay(math32.NewVector3(box.Min.X, box.Max.Y, box.Min.Z), math32.NewVector3(0, 0, box.Max.Z-box.Min.Z).Normalize()))
	edges = append(edges, math32.NewRay(math32.NewVector3(box.Max.X, box.Max.Y, box.Min.Z), math32.NewVector3(0, 0, box.Max.Z-box.Min.Z).Normalize()))

	return edges
}

func rp(ray *math32.Ray, plane *math32.Plane) (math32.Vector3, error) {

	i := ray.IntersectPlane(plane, nil)
	if i != nil {
		return math32.Vector3{i.X, i.Y, i.Z}, nil
	}
	return math32.Vector3{}, errors.New("no intersection")

}

func ToPlaneUV(v math32.Vector3, pNormal math32.Vector3) *math32.Vector2 {

	v.ProjectOnPlane(&pNormal)
	return math32.NewVector2(v.X, v.Y)
}

func ToPlaneUVBatch(pts []math32.Vector3, pNormal math32.Vector3) []*math32.Vector2 {

	var res []*math32.Vector2
	for _, pt := range pts {
		v2 := ToPlaneUV(pt, pNormal)
		res = append(res, v2)
	}

	return res
}
