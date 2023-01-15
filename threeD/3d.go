package threeD

import (
	volume "awesomeProject/dicom"
	"fmt"
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

type GuiState struct {
	Debug     bool
	Angle     *math32.Vector3
	Slice     *math32.Vector3
	DcmNode   *core.Node
	DebugNode *core.Node
	Axial     volume.SliceFrame
}

func placeButtons(scene *core.Node, strip []ButtonStrip, guiState *GuiState, v volume.Volume) {
	for i, b := range strip {
		rot := Rotation(i)
		button := gui.NewButton("Slice" + mapRotation(b.Rotation))
		label := gui.NewLabel("0")
		label.SetBgColor(math32.NewColor("darkorange"))
		button.SetPosition(10, float32(i*30))
		label.SetPosition(200, float32(i*30))
		button.Subscribe(gui.OnClick, func(name string, ev interface{}) {
			if rot == X {
				guiState.Slice.X += 1
				label.SetText(fmt.Sprintf("%f", (guiState.Slice.X)))
			} else if rot == Y {
				guiState.Slice.Y += 1
				label.SetText(fmt.Sprintf("%f", (guiState.Slice.Y)))
			} else if rot == Z {
				guiState.Slice.Z += 1
				sliceFrame := volume.Axial(v, int(guiState.Slice.Z))
				v.Cut(sliceFrame)
				guiState.Axial = sliceFrame
				label.SetText(fmt.Sprintf("%f", (guiState.Slice.Z)))
			} else {

				guiState.Slice.X = 0
				guiState.Slice.Y = 0
				guiState.Slice.Z = 0

				label.SetText(fmt.Sprintf("%f", 0))
			}
		})
		scene.Add(button)
		scene.Add(label)
	}
	debugBtn := gui.NewCheckBox("dbg")
	debugBtn.SetPosition(10, float32(150))
	debugBtn.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		guiState.Debug = !guiState.Debug
	})
	scene.Add(debugBtn)
}

func incAngle(f *float32) {
	*f = *f + 0.1
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
	sliceFrame := volume.Axial(v, v.DcmData.Depth/2)
	v.Cut(sliceFrame)
	guiState := GuiState{
		Debug:     true,
		Slice:     math32.NewVector3(0, 0, float32(v.DcmData.Depth)/2),
		Angle:     math32.NewVec3(),
		DcmNode:   core.NewNode(),
		DebugNode: core.NewNode(),
		Axial:     sliceFrame,
	}
	var btns []ButtonStrip
	btns = append(btns, ButtonStrip{0}, ButtonStrip{1}, ButtonStrip{2}, ButtonStrip{3})
	// Set the scene to be managed by the gui manager
	gui.Manager().Set(scene)
	placeButtons(scene, btns, &guiState, v)

	// Create camera and orbit control
	width, height := a.GetSize()
	aspect := float32(width) / float32(height)
	cam := camera.New(aspect)
	cam.SetPosition(0, 0, 1000)
	cam.SetProjection(camera.Orthographic)
	scene.Add(cam)
	// Create and add lights to the scene
	scene.Add(light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8))
	pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 5.0)
	pointLight.SetPosition(1, 0, 2)
	scene.Add(pointLight)

	// Create and add an axis helper to the scene
	axis := helper.NewAxes(1000)
	scene.Add(axis)

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

	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {

		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)

		guiState.DcmNode = Draw(guiState.Axial, v, guiState.DcmNode)
		guiState.DebugNode = DrawDebug(guiState.Axial, guiState.DebugNode, guiState.Debug)

		renderer.Render(scene, cam)
		renderer.Render(guiState.DcmNode, cam)
		renderer.Render(guiState.DebugNode, cam)
	})
}

func DrawDebug(sliceFrame volume.SliceFrame, node *core.Node, debug bool) *core.Node {
	if node != nil {
		node.RemoveAll(true)
	}
	if !debug {
		return node
	}
	addDots([]math32.Vector3{*sliceFrame.RotatedFrame.Origin}, node, &math32.Color{1, 0, 0}, true)
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
	plane.ApplyMatrix(math32.NewMatrix4().MakeTranslation(w/2, h/2, .5).Multiply(s.RotatedFrame.Basis))

	tex2 := texture.NewTexture2DFromRGBA(*s.Mpr)

	mat1 := material.NewStandard(&math32.Color{1, 1, 1})
	mat1.AddTexture(tex2)
	mat1.SetSide(material.SideDouble)
	mPlane := graphic.NewMesh(plane, mat1)
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
