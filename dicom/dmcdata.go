package volume

import (
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
	"github.com/ungerik/go3d/mat3"
	"github.com/ungerik/go3d/mat4"
	"github.com/ungerik/go3d/vec3"
	"github.com/ungerik/go3d/vec4"
	"math"
	"strconv"
)

type DcmData struct {
	Rows        int
	Cols        int
	Depth       int
	Window      int
	Level       int
	Slope       int
	Intercept   int
	Calibration mat4.T
	Origin      vec3.T
	VoxelSize   vec3.T
}

func readPixelData(dcm dicom.Dataset, tag tag.Tag) (dicom.PixelDataInfo, error) {
	pixelDataElement, err := dcm.FindElementByTag(tag)
	if err != nil {
		return dicom.PixelDataInfo{}, err
	}
	pixelDataInfo := dicom.MustGetPixelDataInfo(pixelDataElement.Value)
	return pixelDataInfo, nil
}

func readTag(dcm dicom.Dataset, tag tag.Tag) (int, error) {
	element, err := dcm.FindElementByTag(tag)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(element.Value.GetValue().([]string)[0])
}

func readCal(dcm dicom.Dataset, tag tag.Tag) (mat3.T, []vec3.T, error) {
	element, err := dcm.FindElementByTag(tag)
	if err != nil {
		return mat3.Ident, []vec3.T{}, err
	}
	values := element.Value.GetValue().([]string)
	dirx := vec3.T{readFloat(values[0]), readFloat(values[1]), readFloat(values[2])}
	dirx.Normalize()
	diry := vec3.T{readFloat(values[3]), readFloat(values[4]), readFloat(values[5])}
	diry.Normalize()
	dirz := vec3.Cross(&dirx, &diry)
	dirz.Normalize()
	return mat3.T{dirx, diry, dirz}, []vec3.T{dirx, diry, dirz}, nil
}

func readOrigin(dcm dicom.Dataset, tag tag.Tag) (vec3.T, error) {
	element, err := dcm.FindElementByTag(tag)
	if err != nil {
		return vec3.Zero, err
	}
	values := element.Value.GetValue().([]string)
	dirx := vec3.T{readFloat(values[0]), readFloat(values[1]), readFloat(values[2])}
	return dirx, nil
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
	orientation, dirs, _ := readCal(dataset, tag.ImageOrientationPatient)
	intercept, _ := readTag(dataset, tag.RescaleIntercept)
	origin, _ := readOrigin(dataset, tag.ImagePositionPatient)
	dirz := orientation.MulVec3(&vec3.UnitZ)
	voxelSize, _ := readVoxelSize(dataset, dcm[1].dataset, tag.PixelSpacing, origin, &*dirz.Normalize())
	x := vec4.FromVec3(&voxelSize)
	cal := mat4.Ident.SetScaling(&x).AssignCoordinateSystem(&dirs[0], &dirs[1], &dirs[2]).SetTranslation(&origin)
	return DcmData{rows,
		cols,
		len(dcm),
		window,
		level,
		slope,
		intercept,
		*cal,
		origin,
		voxelSize}
}

func readVoxelSize(dcm dicom.Dataset, dcm2 dicom.Dataset, tg tag.Tag, origin vec3.T, dirZ *vec3.T) (vec3.T, error) {
	element, err := dcm.FindElementByTag(tg)
	if err != nil {
		return vec3.T{}, err
	}
	origin2, _ := readOrigin(dcm2, tag.ImagePositionPatient)
	values := element.Value.GetValue().([]string)
	dot1 := vec3.Dot(&origin, dirZ)
	dot2 := vec3.Dot(&origin2, dirZ)
	dist := math.Abs(float64(dot2 - dot1))
	return [3]float32{
		readFloat(values[0]),
		readFloat(values[1]),
		float32(dist),
	}, nil
}
