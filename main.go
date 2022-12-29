package main

import (
	volume "awesomeProject/dicom"
	"awesomeProject/threeD"
)

func main() {
	volume := volume.New("C:\\Users\\franc\\Desktop\\98890234_20030505_CT.tar\\98890234_20030505_CT\\98890234\\20030505\\CT\\CT2")
	volume.Render()
	s, e := volume.Cut()
	if e != nil {
		return
	}
	threeD.Draw(volume, s)
	// See also: dicom.Parse which has a generic io.Reader API.

}
