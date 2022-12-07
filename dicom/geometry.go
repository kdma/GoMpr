package volume

import (
	"errors"
	"github.com/ungerik/go3d/mat4"
	"github.com/ungerik/go3d/vec2"
	"github.com/ungerik/go3d/vec3"
	"math"
)

type Ray struct {
	Origin    vec3.T
	Direction vec3.T
}

type Plane struct {
	Origin vec3.T
	Normal vec3.T
	Frame  mat4.T
}

func getSides(box vec3.Box, plane Plane) []Ray {

	var edges []Ray
	edges = append(edges, Ray{box.Min, vec3.T{GetX(box.Max) - GetX(box.Min), 0, 0}})
	edges = append(edges, Ray{vec3.T{GetX(box.Min), GetY(box.Max), GetZ(box.Min)}, vec3.T{GetX(box.Max) - GetX(box.Min), 0, 0}})
	edges = append(edges, Ray{vec3.T{GetX(box.Min), GetY(box.Max), GetZ(box.Min)}, vec3.T{GetX(box.Max) - GetX(box.Min), 0, 0}})
	edges = append(edges, Ray{vec3.T{GetX(box.Min), GetY(box.Max), GetZ(box.Max)}, vec3.T{GetX(box.Max) - GetX(box.Min), 0, 0}})

	edges = append(edges, Ray{vec3.T{GetX(box.Min), GetY(box.Min), GetZ(box.Min)}, vec3.T{0, GetY(box.Max) - GetY(box.Min), 0}})
	edges = append(edges, Ray{vec3.T{GetX(box.Max), GetY(box.Min), GetZ(box.Min)}, vec3.T{0, GetY(box.Max) - GetY(box.Min), 0}})
	edges = append(edges, Ray{vec3.T{GetX(box.Min), GetY(box.Min), GetZ(box.Max)}, vec3.T{0, GetY(box.Max) - GetY(box.Min), 0}})
	edges = append(edges, Ray{vec3.T{GetX(box.Max), GetY(box.Min), GetZ(box.Max)}, vec3.T{0, GetY(box.Max) - GetY(box.Min), 0}})

	edges = append(edges, Ray{vec3.T{GetX(box.Min), GetY(box.Min), GetZ(box.Min)}, vec3.T{0, 0, GetZ(box.Max) - GetZ(box.Min)}})
	edges = append(edges, Ray{vec3.T{GetX(box.Max), GetY(box.Min), GetZ(box.Min)}, vec3.T{0, 0, GetZ(box.Max) - GetZ(box.Min)}})
	edges = append(edges, Ray{vec3.T{GetX(box.Min), GetY(box.Max), GetZ(box.Min)}, vec3.T{0, 0, GetZ(box.Max) - GetZ(box.Min)}})
	edges = append(edges, Ray{vec3.T{GetX(box.Max), GetY(box.Max), GetZ(box.Min)}, vec3.T{0, 0, GetZ(box.Max) - GetZ(box.Min)}})

	return edges
}

func AABBIntersections(AAbb AABB, frame mat4.T) (SliceFrame, error) {

	var intersections []vec3.T
	var origin = frame.MulVec3(&vec3.Zero)
	var z = frame.MulVec3(&vec3.UnitZ)
	z.Normalize()
	plane := Plane{origin, z, frame}
	for _, ray := range getSides(AAbb.Box, plane) {
		pt, err := RayPlaneIntersection(ray, plane)
		if err == nil {
			intersections = append(intersections, pt)
		}
	}

	return SliceFrame{
		frame,
		plane,
		AAbb,
	}, nil
}
func RayPlaneIntersection(ray Ray, plane Plane) (vec3.T, error) {
	var d = vec3.Dot(&plane.Origin, plane.Normal.Invert())
	var t = -(d + vec3.Dot(&ray.Origin, &plane.Normal)) / vec3.Dot(&ray.Direction, &plane.Normal)
	if t < 1e06 {
		return vec3.Zero, errors.New("no intersection")
	}
	return vec3.Add(&ray.Origin, ray.Direction.Scale(t)), nil
}

func ToPlaneUV(v vec3.T, p Plane) vec2.T {

	onPlane := vec3.Sub(&v, &p.Origin)
	xDir := p.Frame.MulVec3(&vec3.UnitX)
	yDir := p.Frame.MulVec3(&vec3.UnitY)
	return vec2.T{vec3.Dot(&onPlane, &xDir), vec3.Dot(&onPlane, &yDir)}
}

type Box2f struct {
	Min vec2.T
	Max vec2.T
}

func AABBBox2(corners []vec2.T) Box2f {
	minx, miny := float32(math.MaxFloat32), float32(math.MaxFloat32)
	maxx, maxy := float32(math.SmallestNonzeroFloat32), float32(math.SmallestNonzeroFloat32)
	for _, corner := range corners {

		cx := corner.Get(0, 0)
		cy := corner.Get(0, 1)
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

	return Box2f{
		Min: vec2.T{minx, miny},
		Max: vec2.T{maxx, maxy},
	}
}
func Multiply(pts []vec3.T, frame mat4.T) []vec3.T {

	var res []vec3.T
	for _, pt := range pts {
		copy := pt
		frame.TransformVec3(&copy)
		res = append(res, copy)
	}
	return res
}

func ToPlaneUVBatch(pts []vec3.T, plane Plane) []vec2.T {

	var res []vec2.T
	for _, pt := range pts {
		v2 := ToPlaneUV(pt, plane)
		res = append(res, v2)
	}

	return res
}
