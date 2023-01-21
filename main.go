package main

import (
	volume "awesomeProject/dicom"
	"awesomeProject/threeD"
)

func main() {
	volume := volume.New("C:\\Users\\franc\\Desktop\\OneDrive_2023-01-04\\Circle of Willis")
	threeD.Init(volume)
}
