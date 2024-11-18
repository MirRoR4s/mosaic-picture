package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"os"
	"testing"
)

func TestResize(t *testing.T) {
	file, _ := os.Open("tiles/cat1.jpg")
	img, _, _ := image.Decode(file)

	img1 := resize(img, 10)
	
	buf1 := new(bytes.Buffer)
	jpeg.Encode(buf1, img1.SubImage(img1.Bounds()), nil)

	os.WriteFile("tiles/cat1-resize.jpg", buf1.Bytes(), 0644)
}