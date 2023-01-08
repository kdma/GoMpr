package threeD

import (
	volume "awesomeProject/dicom"
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

func Draw(v volume.Volume, s volume.SliceFrame) {
	// Create application and scene
	a := app.App()
	scene := core.NewNode()

	// Set the scene to be managed by the gui manager
	gui.Manager().Set(scene)

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

	addbox(v, scene, &math32.Color{1, 1, 1})
	addDots(s.AABB.CalibratedCorners, scene, &math32.Color{0, 0, 1})
	addDots(s.Intersections, scene, &math32.Color{0, 1, 0})
	addRays(s.Rays, s, scene)
	addplane(s, v, scene)

	// Create and add lights to the scene
	scene.Add(light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8))
	pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 5.0)
	pointLight.SetPosition(1, 0, 2)
	scene.Add(pointLight)

	// Create and add an axis helper to the scene
	scene.Add(helper.NewAxes(100))
	grid := helper.NewGrid(1000, 10, &math32.Color{0.5, 0.5, 0.5})
	scene.Add(grid)

	// Set background color to gray
	a.Gls().ClearColor(1, 1, 1, 1.0)

	// Run the application
	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		renderer.Render(scene, cam)
	})
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
func addDots(v []math32.Vector3, scene *core.Node, c *math32.Color) {
	for _, el := range v {
		dot := geometry.NewSphere(1, 4, 4)
		mat1 := material.NewStandard(c)
		mat1.SetWireframe(true)
		mat1.SetSide(material.SideDouble)
		mDot := graphic.NewMesh(dot, mat1)
		mDot.SetPosition(el.X, el.Y, el.Z)
		scene.Add(mDot)
	}
}

func addplane(s volume.SliceFrame, v volume.Volume, scene *core.Node) {
	w := s.ImageSizeInMm.X
	h := s.ImageSizeInMm.Y
	plane := geometry.NewBox(w, h, 1)
	plane.ApplyMatrix(s.Basis)

	texfile := "C:\\Users\\franc\\Desktop\\Nuova cartella\\mpr.jpg"
	tex2, _ := texture.NewTexture2DFromImage(texfile)

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
