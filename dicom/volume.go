package volume

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/g3n/engine/math32"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
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

func (v Volume) Render() {

	for z, slice := range v.Data {
		img := image.NewRGBA(image.Rect(0, 0, len(slice[0]), len(slice)))
		for r, row := range slice {
			for c := range row {
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
	imgb := make([][]byte, data.Rows)
	for i := 0; i < data.Rows; i++ {
		imgb[i] = make([]byte, data.Cols)
	}

	for i := 0; i < len(nativeFrame.Data); i++ {
		pixel := float32(nativeFrame.Data[i][0])*data.Slope + data.Intercept
		pixelByte := uint8(0)
		if pixel <= data.Window-0.5-(data.Level-1)/2 {
			pixelByte = 0
		} else if pixel > data.Window-0.5+(data.Level-1)/2 {
			pixelByte = 255
		} else {
			pixelByte = uint8((((pixel)-((data.Window)-0.5))/(data.Level-1) + 0.5) * (255))
		}
		c := i % data.Cols
		r := i / data.Rows
		imgb[r][c] = pixelByte
	}

	return imgb, nil
}

func (volume Volume) GetCorners() AABB {

	min := math32.Vector3{0, 0, 0}
	max := math32.Vector3{float32(volume.DcmData.Cols), float32(volume.DcmData.Rows), float32(volume.DcmData.Depth)}
	box := math32.NewBox3(&min, &max)

	var corners []math32.Vector3
	corners = append(corners, math32.Vector3{box.Min.X, box.Min.Y, box.Min.Z})
	corners = append(corners, math32.Vector3{box.Min.X, box.Max.Y, box.Min.Z})
	corners = append(corners, math32.Vector3{box.Max.X, box.Min.Y, box.Min.Z})
	corners = append(corners, math32.Vector3{box.Max.X, box.Max.Y, box.Min.Z})
	corners = append(corners, math32.Vector3{box.Min.X, box.Min.Y, box.Max.Z})
	corners = append(corners, math32.Vector3{box.Min.X, box.Max.Y, box.Max.Z})
	corners = append(corners, math32.Vector3{box.Max.X, box.Min.Y, box.Max.Z})
	corners = append(corners, math32.Vector3{box.Max.X, box.Max.Y, box.Max.Z})

	minX, minY, minZ := float32(math.MaxFloat32), float32(math.MaxFloat32), float32(math.MaxFloat32)
	maxX, maxY, maxZ := float32(math.SmallestNonzeroFloat32), float32(math.SmallestNonzeroFloat32), float32(math.SmallestNonzeroFloat32)

	calibratedCorners := []math32.Vector3{}
	for i := 0; i < len(corners); i++ {
		c := corners[i]
		c.ApplyMatrix4(volume.DcmData.Calibration)
		calibratedCorners = append(calibratedCorners, c)
		x := c.X
		y := c.Y
		z := c.Z
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
		if z < minZ {
			minZ = z
		}
		if z > maxZ {
			maxZ = z
		}
	}

	calibratedBox := math32.NewBox3(math32.NewVector3(minX, minY, minZ), math32.NewVector3(maxX, maxY, maxZ))
	return AABB{
		CalibratedCorners: calibratedCorners,
		Box:               calibratedBox,
	}
}
