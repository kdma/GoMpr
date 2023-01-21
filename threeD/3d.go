package threeD

import (
	volume "awesomeProject/dicom"
	"fmt"
	"time"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/texture"
	"github.com/g3n/engine/util/helper"
	"github.com/g3n/engine/window"
)

type CutCallback func(float32, volume.Volume)

type GuiState struct {
	Debug         bool
	Dirty         bool
	Slice         *math32.Vector3
	AxialNode     *core.Node
	CoronalNode   *core.Node
	SagittallNode *core.Node
	CustomNode    *core.Node
	DebugNode     *core.Node
	Axial         volume.SliceFrame
	Coronal       volume.SliceFrame
	Sagittal      volume.SliceFrame
	Custom        volume.SliceFrame
}

func updateAxial(g *GuiState, v volume.Volume) {
	g.Axial = volume.Axial(v, int(g.Slice.Z))
	g.Axial.Cut(v)
}
func updateSagittal(g *GuiState, v volume.Volume) {
	g.Sagittal = volume.Sagittal(v, int(g.Slice.X))
	g.Sagittal.Cut(v)
}
func updateCoronal(g *GuiState, v volume.Volume) {
	g.Coronal = volume.Coronal(v, int(g.Slice.Y))
	g.Coronal.Cut(v)
}

func updateFree(g *GuiState, v volume.Volume) {
	g.Custom = volume.FreeRotation(v, math32.NewMatrix4())
	g.Custom.Cut(v)
}

func placeSliderButton(scene *core.Node,
	x float32,
	y float32,
	normalizedValue float32,
	v volume.Volume,
	max float32,
	cb CutCallback) *gui.Slider {
	s1 := gui.NewHSlider(400, 32)
	s1.SetPosition(x, y)
	s1.SetValue(normalizedValue)
	volumeValue := denorm(s1.Value(), max)
	s1.SetText(fmt.Sprintf("%1.0f", volumeValue))
	s1.Subscribe(gui.OnChange, func(evname string, ev interface{}) {
		denormalized := denorm(s1.Value(), max)
		s1.SetText(fmt.Sprintf("%1.0f", denormalized))
		cb(denormalized, v)
	})
	scene.Add(s1)
	return s1
}
func norm(v float32, max float32) float32 {
	return v / max
}
func denorm(v float32, max float32) float32 {
	return v * max
}
func placeButtons(scene *core.Node, labels []string, g *GuiState, v volume.Volume) {

	sagBtn := placeSliderButton(scene, 10, 0, norm(g.Slice.Y, float32(v.DcmData.Cols)), v, float32(v.DcmData.Cols), func(f float32, v volume.Volume) {
		g.Slice.X = f
		g.Dirty = true
	})

	cornBtn := placeSliderButton(scene, 10, 30, norm(g.Slice.Y, float32(v.DcmData.Rows)), v, float32(v.DcmData.Rows), func(f float32, v volume.Volume) {
		g.Slice.Y = f
		g.Dirty = true
	})

	axialBtn := placeSliderButton(scene, 10, 60, norm(g.Slice.Z, float32(v.DcmData.Depth)), v, float32(v.DcmData.Depth), func(f float32, v volume.Volume) {
		g.Slice.Z = f
		g.Dirty = true
	})

	resetBtn := gui.NewButton("Reset")
	resetBtn.SetPosition(10, 90)
	resetBtn.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		g.Slice.X = float32(v.DcmData.Cols / 2)
		sagBtn.SetValue(norm(g.Slice.X, float32(v.DcmData.Cols)))
		g.Slice.Y = float32(v.DcmData.Rows / 2)
		cornBtn.SetValue(norm(g.Slice.Y, float32(v.DcmData.Rows)))
		g.Slice.Z = float32(v.DcmData.Depth / 2)
		axialBtn.SetValue(norm(g.Slice.Z, float32(v.DcmData.Depth)))
		updateAxial(g, v)
		updateCoronal(g, v)
		updateSagittal(g, v)
	})
	scene.Add(resetBtn)

	debugBtn := gui.NewCheckBox("dbg")
	debugBtn.SetPosition(10, float32(150))
	debugBtn.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		g.Debug = !g.Debug
	})
	scene.Add(debugBtn)
}

func Init(v volume.Volume) {
	a := app.App()
	scene := core.NewNode()
	guiState := GuiState{
		Debug:       true,
		Dirty:       true,
		Slice:       math32.NewVector3(float32(v.DcmData.Cols)/2, float32(v.DcmData.Rows)/2, float32(v.DcmData.Depth)/2),
		AxialNode:   core.NewNode(),
		CoronalNode: core.NewNode(),
		DebugNode:   core.NewNode(),
		Axial:       volume.SliceFrame{},
		Coronal:     volume.SliceFrame{},
		Sagittal:    volume.SliceFrame{},
	}

	var btns []string
	btns = append(btns, "X", "Y", "Z", "Reset")
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

		if guiState.Dirty {
			updateAxial(&guiState, v)
			updateCoronal(&guiState, v)
			updateSagittal(&guiState, v)
			guiState.AxialNode = Draw(guiState.Axial, v, guiState.AxialNode, math32.NewColor("blue"))
			guiState.CoronalNode = Draw(guiState.Coronal, v, guiState.CoronalNode, math32.NewColor("green"))
			guiState.SagittallNode = Draw(guiState.Sagittal, v, guiState.SagittallNode, math32.NewColor("red"))
			guiState.DebugNode = DrawDebug(guiState.Sagittal, guiState.DebugNode, guiState.Debug)
			guiState.Dirty = false
		}

		renderer.Render(scene, cam)
		renderer.Render(guiState.AxialNode, cam)
		renderer.Render(guiState.CoronalNode, cam)
		renderer.Render(guiState.SagittallNode, cam)
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

func Draw(sliceFrame volume.SliceFrame, v volume.Volume, node *core.Node, c *math32.Color) *core.Node {

	if node != nil {
		node.RemoveAll(true)
	}

	scene := core.NewNode()
	addbox(v, scene, &math32.Color{1, 1, 1})
	addDots(sliceFrame.AABB.CalibratedCorners, scene, c, false)
	addDots(sliceFrame.Intersections, scene, c, false)

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
	plane.ApplyMatrix(math32.NewMatrix4().MakeTranslation(w/2, h/2, .5))

	tex2 := texture.NewTexture2DFromRGBA(*s.Mpr)

	mat1 := material.NewStandard(&math32.Color{1, 1, 1})
	mat1.AddTexture(tex2)
	mat1.SetSide(material.SideFront)
	mPlane := graphic.NewMesh(plane, mat1)
	mPlane.SetMatrix(math32.NewMatrix4().Multiply(s.RotatedFrame.Basis).SetPosition(s.RotatedFrame.Origin))
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
