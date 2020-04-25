package main

import (
	"fmt"

	"github.com/vinijabes/gocompositor/runnable"
)

func main() {
	fmt.Println("Hello World")

	xvfb := runnable.NewXVFB(":77")
	err := xvfb.Start()

	if err != nil {
		fmt.Println(err)
	}

	xvfb.Stop()
	// pipeline, err := gstreamer.New("videotestsrc  ! capsfilter name=filter ! autovideosink")
	// if err != nil {
	// 	fmt.Println("pipeline create error", err)
	// }

	// filter := pipeline.FindElement("filter")

	// if filter == nil {
	// 	fmt.Println("pipeline find element error ")
	// }

	// filter.SetCap("video/x-raw,width=1280,height=720")

	// pipeline.Start()

	for {

	}

}
