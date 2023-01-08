package volume

import (
	"errors"
	"math"

	"github.com/g3n/engine/math32"
)

func getSides(box *math32.Box3, v Volume) []*math32.Ray {

	var edges []*math32.Ray
	noT := math32.NewMatrix4().Identity().Copy(v.DcmData.Calibration).SetPosition(math32.NewVec3())
	dirX := math32.NewVector3(1, 0, 0).ApplyMatrix4(noT).Normalize()
	dirY := math32.NewVector3(0, 1, 0).ApplyMatrix4(noT).Normalize()
	dirZ := math32.NewVector3(0, 0, 1).ApplyMatrix4(noT).Normalize()

	edges = append(edges, math32.NewRay(math32.NewVector3(box.Min.X, box.Min.Y, box.Min.Z), dirX))
	edges = append(edges, math32.NewRay(math32.NewVector3(box.Min.X, box.Max.Y, box.Min.Z), dirX))
	edges = append(edges, math32.NewRay(math32.NewVector3(box.Min.X, box.Min.Y, box.Max.Z), dirX))
	edges = append(edges, math32.NewRay(math32.NewVector3(box.Min.X, box.Max.Y, box.Max.Z), dirX))

	edges = append(edges, math32.NewRay(math32.NewVector3(box.Min.X, box.Min.Y, box.Min.Z), dirY))
	edges = append(edges, math32.NewRay(math32.NewVector3(box.Max.X, box.Min.Y, box.Min.Z), dirY))
	edges = append(edges, math32.NewRay(math32.NewVector3(box.Min.X, box.Min.Y, box.Max.Z), dirY))
	edges = append(edges, math32.NewRay(math32.NewVector3(box.Max.X, box.Min.Y, box.Max.Z), dirY))

	edges = append(edges, math32.NewRay(math32.NewVector3(box.Min.X, box.Min.Y, box.Min.Z), dirZ))
	edges = append(edges, math32.NewRay(math32.NewVector3(box.Max.X, box.Min.Y, box.Min.Z), dirZ))
	edges = append(edges, math32.NewRay(math32.NewVector3(box.Min.X, box.Max.Y, box.Min.Z), dirZ))
	edges = append(edges, math32.NewRay(math32.NewVector3(box.Max.X, box.Max.Y, box.Min.Z), dirZ))

	return edges
}
func rp2(ray *math32.Ray, plane *math32.Plane, aabb AABB) (math32.Vector3, error) {

	intersection := ray.IntersectBox(aabb.Box, nil)
	if intersection != nil {
		return *intersection, nil
	}
	return *math32.NewVec3(), errors.New("no intersection")
}

func rp(ray *math32.Ray, plane *math32.Plane, aabb AABB) (math32.Vector3, error) {

	ct := math32.Vector3{}
	i := ray.IntersectPlane(plane, &ct)
	isXNan := math.IsNaN(float64(i.X))
	isYNan := math.IsNaN(float64(i.Y))
	isZNan := math.IsNaN(float64(i.Z))
	if i != nil && !isXNan && !isYNan && !isZNan {
		return math32.Vector3{X: i.X, Y: i.Y, Z: i.Z}, nil
	} else {
		return math32.Vector3{}, errors.New("no intersection")
	}

}

func ToPlaneUV(pts []math32.Vector3, pNormal math32.Vector3) []*math32.Vector2 {

	var res []*math32.Vector2
	for _, pt := range pts {
		ptCopy := math32.NewVector3(pt.X, pt.Y, pt.Z)
		ptCopy.ProjectOnPlane(&pNormal)
		v2 := math32.NewVector2(ptCopy.X, ptCopy.Y)
		res = append(res, v2)
	}

	return res
}
