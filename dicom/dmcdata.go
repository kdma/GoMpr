package volume

import (
	"math"
	"strconv"

	"github.com/g3n/engine/math32"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

type DcmData struct {
	Rows        int
	Cols        int
	Depth       int
	Window      float32
	Level       float32
	Slope       float32
	Intercept   float32
	Calibration *math32.Matrix4
	Orientation *math32.Matrix4
	Origin      math32.Vector3
	VoxelSize   math32.Vector3
}

func readPixelData(dcm dicom.Dataset, tag tag.Tag) (dicom.PixelDataInfo, error) {
	pixelDataElement, err := dcm.FindElementByTag(tag)
	if err != nil {
		return dicom.PixelDataInfo{}, err
	}
	pixelDataInfo := dicom.MustGetPixelDataInfo(pixelDataElement.Value)
	return pixelDataInfo, nil
}

func readTag(dcm dicom.Dataset, tag tag.Tag) (float32, error) {
	element, err := dcm.FindElementByTag(tag)
	if err != nil {
		return 0, err
	}
	f, err := strconv.ParseFloat(element.Value.GetValue().([]string)[0], 32)
	if err != nil {
		return 0, err
	}
	return float32(f), nil
}

func readCal(dcm dicom.Dataset, tag tag.Tag) (*math32.Matrix4, []math32.Vector3, error) {
	element, err := dcm.FindElementByTag(tag)
	if err != nil {
		return math32.NewMatrix4(), []math32.Vector3{}, err
	}
	values := element.Value.GetValue().([]string)
	dirX := math32.Vector3{readFloat(values[0]), readFloat(values[1]), readFloat(values[2])}
	dirY := math32.Vector3{readFloat(values[3]), readFloat(values[4]), readFloat(values[5])}

	dirz := math32.NewVector3(0, 0, 0).CrossVectors(&dirX, &dirY)
	dirz.Normalize()

	m := math32.NewMatrix4().MakeBasis(&dirX, &dirY, dirz)
	return m, []math32.Vector3{dirX, dirY, *dirz}, nil
}

func readOrigin(dcm dicom.Dataset, tag tag.Tag) (math32.Vector3, error) {
	element, err := dcm.FindElementByTag(tag)
	if err != nil {
		return *math32.NewVector3(0, 0, 0), err
	}
	values := element.Value.GetValue().([]string)
	dirx := math32.NewVector3(readFloat(values[0]), readFloat(values[1]), readFloat(values[2]))
	return *dirx, nil
}
func readFloat(num string) float32 {
	f, err := strconv.ParseFloat(num, 32)
	if err != nil {
		return 0
	}
	return float32(f)
}

func readTagInt(dcm dicom.Dataset, tag tag.Tag) (int, error) {
	element, err := dcm.FindElementByTag(tag)
	if err != nil {
		return 0, err
	}
	return element.Value.GetValue().([]int)[0], nil
}

func readDcmData(dcm []DicomFile) DcmData {
	dataset := dcm[0].dataset
	window, _ := readTag(dataset, tag.WindowCenter)
	level, _ := readTag(dataset, tag.WindowWidth)
	rows, _ := readTagInt(dataset, tag.Rows)
	cols, _ := readTagInt(dataset, tag.Columns)
	slope, _ := readTag(dataset, tag.RescaleSlope)
	orientation, _, _ := readCal(dataset, tag.ImageOrientationPatient)
	intercept, _ := readTag(dataset, tag.RescaleIntercept)
	origin, _ := readOrigin(dataset, tag.ImagePositionPatient)
	z := math32.NewVector3(0, 0, 1)
	z.ApplyMatrix4(orientation)
	z.Normalize()
	voxelSize, _ := readVoxelSize(dataset, dcm[1].dataset, tag.PixelSpacing, origin, z)
	cal := math32.NewMatrix4().Multiply(orientation).Scale(voxelSize).SetPosition(&origin)
	ori := math32.NewMatrix4().Multiply(orientation)
	return DcmData{rows,
		cols,
		len(dcm),
		window,
		level,
		slope,
		intercept,
		cal,
		ori,
		origin,
		*voxelSize}
}

func toDegree(rad float32) float32 {
	return rad * (180 / math.Pi)
}
func readVoxelSize(dcm dicom.Dataset, dcm2 dicom.Dataset, tg tag.Tag, origin math32.Vector3, dirZ *math32.Vector3) (*math32.Vector3, error) {
	element, err := dcm.FindElementByTag(tg)
	if err != nil {
		return math32.NewVector3(0, 0, 0), err
	}
	origin2, _ := readOrigin(dcm2, tag.ImagePositionPatient)
	values := element.Value.GetValue().([]string)
	dot1 := origin.Dot(dirZ)
	dot2 := origin2.Dot(dirZ)
	dist := math.Abs(float64(dot2 - dot1))
	return math32.NewVector3(readFloat(values[0]), readFloat(values[1]), float32(dist)), nil
}
