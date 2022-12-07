package volume

import (
	"fmt"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
	"github.com/ungerik/go3d/mat4"
	"github.com/ungerik/go3d/quaternion"
	"github.com/ungerik/go3d/vec3"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
)

type DicomFile struct {
	filePath string
	slice    [][]uint16
	dataset  dicom.Dataset
}

type Volume struct {
	Dicoms  []DicomFile
	Data    [][][]byte
	DcmData DcmData
}

func (v Volume) Cut() error {
	Aabb := getCorners(v)
	//center := vec3.T{float32(v.DcmData.Cols) * 0.5, float32(v.DcmData.Rows) * 0.5, float32(v.DcmData.Depth) * 0.5}

	fromDataToSlicePlane := mat4.Ident.TranslateZ(2)
	fromDataToSlicePlane.MultMatrix(&v.DcmData.Calibration)

	zero := vec3.Zero
	zeta := vec3.UnitZ
	origin := fromDataToSlicePlane.MulVec3(&zero)
	z := fromDataToSlicePlane.MulVec3(&zeta)
	z.Normalize()
	angle := vec3.Angle(&z, &zeta)
	//vedi centro di rotazione
	q := quaternion.FromZAxisAngle(angle)
	frame := mat4.Ident.AssignQuaternion(&q)
	frame.Translate(&origin)
	sliceFrame, err := AABBIntersections(Aabb, *frame)
	if err != nil {
		return error(err)
	}
	box2f := AABBBox2(ToPlaneUVBatch(sliceFrame.AABB.CalibratedCorners, sliceFrame.Plane))
	return nil
}

func RenderSlice(frame SliceFrame) {
	imgWidth := 256
	fromSliceToVoxel := mat4.Ident.ScaleVec3()
}
func (v Volume) Render() {

	for z, slice := range v.Data {
		img := image.NewRGBA(image.Rect(0, 0, len(slice[0]), len(slice)))
		for r, row := range slice {
			for c, _ := range row {
				pixel := v.Data[z][r][c]
				img.SetRGBA(c, r, color.RGBA{A: 0xFF, R: pixel, G: pixel, B: pixel})
			}
		}
		f, err := os.Create(filepath.Join("C:\\Users\\franc\\Desktop\\Nuova Cartella", fmt.Sprintf("image_%d.jpg", z)))
		if err != nil {
			log.Fatal(err)
		}
		_ = jpeg.Encode(f, img, &jpeg.Options{Quality: 100})
		_ = f.Close()
	}
}

func New(folderPath string) Volume {
	dicoms := importDicoms(folderPath)
	data := make([][][]byte, len(dicoms))

	header := readDcmData(dicoms)
	for i, dcm := range dicoms {
		dcmInfo, _ := readPixelData(dcm.dataset, tag.PixelData)
		img, _ := loadFrame(header, dcmInfo)
		data[i] = img
	}
	return Volume{Dicoms: dicoms, Data: data, DcmData: header}
}

func importDicoms(folderPath string) []DicomFile {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		log.Fatal(err)
	}
	var paths []DicomFile
	for _, file := range files {
		filepath := filepath.Join(folderPath, file.Name())
		dataset, err := dicom.ParseFile(filepath, nil)
		if err != nil {
			log.Fatal(err)
			continue
		}
		paths = append(paths, DicomFile{filePath: filepath, dataset: dataset})
	}
	return paths
}

func loadFrame(data DcmData, pixeldata dicom.PixelDataInfo) ([][]byte, error) {
	frame := pixeldata.Frames[0]
	nativeFrame, _ := frame.GetNativeFrame()
	img := make([][]byte, data.Rows)
	for i := 0; i < data.Rows; i++ {
		img[i] = make([]byte, data.Cols)
	}

	for j := 0; j < len(nativeFrame.Data); j++ {
		pixel := float32(nativeFrame.Data[j][0]*data.Slope + data.Intercept)

		if pixel <= float32(data.Window)-0.5-(float32(data.Level-1)/2) {
			pixel = 0
		} else if pixel > float32(data.Window)-0.5+float32(data.Level-1)/2 {
			pixel = 255
		} else {
			pixel = ((pixel-(float32(data.Window)-0.5))/float32(data.Level-1) + 0.5) * float32(255)
		}
		img[j/data.Cols][j%data.Cols] = byte(pixel)
	}
	return img, nil
}
