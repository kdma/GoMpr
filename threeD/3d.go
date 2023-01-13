package threeD

import (
	volume "awesomeProject/dicom"
	"math"
	"time"

	"github.com/g3n/engine/math32"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/texture"
	"github.com/g3n/engine/util/helper"
	"github.com/g3n/engine/window"
)

type Rotation int

const (
	X     Rotation = 0
	Y              = 1
	Z              = 2
	Reset          = 3
)

func toDegree(rad float32) float32 {
	return rad * (180 / math.Pi)
}

type ButtonStrip struct {
	Rotation Rotation
}

func placeButtons(scene *core.Node, strip []ButtonStrip, angle *math32.Vector3, debug *bool) {
	for i, b := range strip {
		rot := Rotation(i)
		button := gui.NewButton("Rotation" + mapRotation(b.Rotation))
		button.SetPosition(10, float32(i*30))
		button.Subscribe(gui.OnClick, func(name string, ev interface{}) {
			if rot == X {
				incAngle(&angle.X)
			} else if rot == Y {
				incAngle(&angle.Y)
			} else if rot == Z {
				incAngle(&angle.Z)
			} else {
				angle.X = 0
				angle.Y = 0
				angle.Z = 0
			}
		})
		scene.Add(button)
	}
	debugBtn := gui.NewCheckBox("dbg")
	debugBtn.SetPosition(10, float32(150))
	debugBtn.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		*debug = !*debug
	})
	scene.Add(debugBtn)
}

func incAngle(f *float32) {
	*f = *f + 0.2
}

func mapRotation(rot Rotation) string {
	if rot == X {
		return "X"
	} else if rot == Y {
		return "Y"
	} else if rot == Z {
		return "Z"
	} else {
		return "Reset"
	}
}

func Init(v volume.Volume) {
	a := app.App()
	scene := core.NewNode()
	angle := math32.NewVec3()
	debug := false

	var btns []ButtonStrip
	btns = append(btns, ButtonStrip{0}, ButtonStrip{1}, ButtonStrip{2}, ButtonStrip{3})
	// Set the scene to be managed by the gui manager
	gui.Manager().Set(scene)
	placeButtons(scene, btns, angle, &debug)

	// Create camera and orbit control
	width, height := a.GetSize()
	aspect := float32(width) / float32(height)
	cam := camera.New(aspect)
	cam.SetPosition(0, 0, 1000)
	cam.SetProjection(camera.Orthographic)
	scene.Add(cam)

	// Set up orbit control for the camera
	camera.NewOrbitControl(cam)

	// Set up callback to update viewport and camera aspect ratio when the window is resized
	onResize := func(evname string, ev interface{}) {
		// Get framebuffer size and update viewport accordingly
		width, height := a.GetSize()
		a.Gls().Viewport(0, 0, int32(width), int32(height))
		// Update the camera's aspect ratio
		cam.SetAspect(float32(width) / float32(height))
	}

	a.Subscribe(window.OnWindowSize, onResize)
	onResize("", nil)

	a.Gls().ClearColor(1, 1, 1, 1.0)

	dcmNode := core.NewNode()
	debugNode := core.NewNode()

	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		basis := math32.NewMatrix4().Multiply(v.DcmData.Orientation).Multiply(math32.NewMatrix4().MakeRotationFromEuler(angle))
		sliceFrame := volume.FreeRotation(v, basis)
		v.Cut(sliceFrame)

		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		renderer.Render(scene, cam)
		renderer.Render(Draw(sliceFrame, v, dcmNode), cam)
		renderer.Render(DrawDebug(sliceFrame, debugNode, debug), cam)
	})
}

func DrawDebug(sliceFrame volume.SliceFrame, node *core.Node, debug bool) *core.Node {
	if node != nil {
		node.RemoveAll(true)
	}
	if !debug {
		return node
	}
	addDots([]math32.Vector3{*sliceFrame.FirstPixelOrigin}, node, &math32.Color{1, 0, 0}, true)
	addRays(sliceFrame.Rays, sliceFrame, node)
	return node
}

func Draw(sliceFrame volume.SliceFrame, v volume.Volume, node *core.Node) *core.Node {

	if node != nil {
		node.RemoveAll(true)
	}

	scene := core.NewNode()
	addbox(v, scene, &math32.Color{1, 1, 1})
	addDots(sliceFrame.AABB.CalibratedCorners, scene, &math32.Color{0, 0, 1}, false)
	addDots(sliceFrame.Intersections, scene, &math32.Color{0, 1, 0}, false)

	addPlane(sliceFrame, v, scene)
	addBasis(sliceFrame, v, scene)

	// Create and add lights to the scene
	scene.Add(light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8))
	pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 5.0)
	pointLight.SetPosition(1, 0, 2)
	scene.Add(pointLight)

	// Create and add an axis helper to the scene
	scene.Add(helper.NewAxes(100))
	grid := helper.NewGrid(1000, 10, &math32.Color{0.5, 0.5, 0.5})
	scene.Add(grid)
	return scene
}

func addBasis(s volume.SliceFrame, v volume.Volume, scene *core.Node) {

	axis := helper.NewAxes(100)
	axis.SetMatrix(s.RotatedFrame.Basis)
	axis.SetPositionVec(s.AABB.Box.Center(nil))
	scene.Add(axis)
}

func addRays(r []math32.Ray, s volume.SliceFrame, scene *core.Node) {
	for i, el := range r {
		c := math32.Color{0, 0, 0}
		if i < 4 {
			c.Set(1, 0, 0)
		}
		if i >= 4 && i < 8 {
			c.Set(0, 1, 0)
		}
		if i >= 8 {
			c.Set(0, 0, 1)
		}
		dot := geometry.NewSphere(1, 4, 4)
		mat1 := material.NewStandard(&c)
		mat1.SetWireframe(true)
		mat1.SetSide(material.SideDouble)
		mDot := graphic.NewMesh(dot, mat1)
		mDot.SetPosition(el.Origin().X, el.Origin().Y, el.Origin().Z)
		scene.Add(mDot)

		// Line segments
		geom10 := geometry.NewGeometry()
		positions := math32.NewArrayF32(0, 0)
		rayO := el.Origin()
		rayD := el.Direction()
		end := rayD.MultiplyScalar(1000)
		dest := rayO.Add(end)
		positions.Append(
			el.Origin().X, el.Origin().Y, el.Origin().Z, dest.X, dest.Y, dest.Z,
		)
		geom10.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
		mat10 := material.NewStandard(&c)
		mesh10 := graphic.NewLines(geom10, mat10)
		scene.Add(mesh10)

	}
}

func addDots(v []math32.Vector3, scene *core.Node, c *math32.Color, magnify bool) {
	for _, el := range v {
		size := 1.0
		if magnify {
			size = 2.5
		}
		dot := geometry.NewSphere(size, 4, 4)
		mat1 := material.NewStandard(c)
		mat1.SetWireframe(true)
		mat1.SetSide(material.SideDouble)
		mDot := graphic.NewMesh(dot, mat1)
		mDot.SetPosition(el.X, el.Y, el.Z)
		scene.Add(mDot)
	}
}

func addPlane(s volume.SliceFrame, v volume.Volume, scene *core.Node) {
	w := s.ImageSizeInMm.X
	h := s.ImageSizeInMm.Y
	plane := geometry.NewBox(w, h, 1)
	plane.ApplyMatrix(s.RotatedFrame.Basis)

	tex2 := texture.NewTexture2DFromRGBA(*s.Mpr)

	mat1 := material.NewStandard(&math32.Color{1, 1, 1})
	mat1.AddTexture(tex2)
	mat1.SetSide(material.SideDouble)
	mPlane := graphic.NewMesh(plane, mat1)
	mPlane.SetPositionVec(s.AABB.Box.Center(nil))
	scene.Add(mPlane)
}

func addbox(v volume.Volume, scene *core.Node, color *math32.Color) {
	geom := geometry.NewBox(float32(v.DcmData.Cols), float32(v.DcmData.Rows), float32(v.DcmData.Depth))
	geom.ApplyMatrix(math32.NewMatrix4().Identity().SetPosition(math32.NewVector3(float32(v.DcmData.Cols)/2, float32(v.DcmData.Rows)/2, float32(v.DcmData.Depth)/2)))
	mat := material.NewStandard(color)
	mat.SetWireframe(true)
	mesh := graphic.NewMesh(geom, mat)
	mm := math32.NewMatrix4().Identity().Copy(v.DcmData.Calibration)
	mesh.SetMatrix(mm)
	scene.Add(mesh)
}
