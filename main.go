package main

import (
	volume "awesomeProject/dicom"
	"awesomeProject/threeD"
	"flag"
	"fmt"
)

func main() {
	var dcmPath = flag.String("dcm", "", "Dicom Path")
	flag.Parse()

	if *dcmPath == "" {
		fmt.Println("Error: you must provide a valid path")
		return
	}
	volume := volume.New(*dcmPath)
	threeD.Init(volume)
}
