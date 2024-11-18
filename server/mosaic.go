package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
)

var TILESDB map[string][3]float64

// averageColor 计算一张图片颜色的平均值。
func averageColor(img image.Image) [3]float64 {
	bounds := img.Bounds()
	r, g, b := 0.0, 0.0, 0.0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r1, g1, b1, _ := img.At(x, y).RGBA()
			r, g, b = r+float64(r1), g+float64(g1), b+float64(b1)
		}
	}

	totalPixels := float64(bounds.Max.X * bounds.Max.Y)
	return [3]float64{r / totalPixels, g / totalPixels, b / totalPixels}
}

// resize 缩放图片至指定宽度
func resize(in image.Image, newWidth int) image.NRGBA {
	bounds := in.Bounds()
	ratio := bounds.Dx() / newWidth
	x0, y0 := bounds.Min.X/ratio, bounds.Min.Y/ratio
	x1, y1 := bounds.Max.X/ratio, bounds.Max.Y/ratio
	out := image.NewNRGBA(image.Rect(x0, y0, x1, y1))

	for y, j := bounds.Min.Y, bounds.Min.Y; y < bounds.Max.Y; y, j = y+ratio, j+1 {
		for x, i := bounds.Min.X, bounds.Min.X; x < bounds.Max.X; x, i = x+ratio, i+1 {
			r, g, b, a := in.At(x, y).RGBA()
			out.SetNRGBA(i, j, color.NRGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)})
		}
	}

	return *out
}

// tilesDB 扫描瓷砖图片所在目录来创建一个瓷砖图片数据库
func tilesDB() map[string][3]float64 {
	fmt.Println("Start populating tiles db ...")
	db := make(map[string][3]float64)
	files, err := os.ReadDir("tiles")
	if err != nil {
		log.Fatalf("read tiles directory error: %v", err)
	}

	for _, f := range files {
		name := "tiles/" + f.Name()
		file, err := os.Open(name)
		if err != nil {
			log.Fatal("cannot open file", name, err)
		}
		defer file.Close()

		img, _, err := image.Decode(file)
		if err != nil {
			log.Fatalln("error in populating TILEDB:", err, name)
		}
		db[name] = averageColor(img)
	}
	fmt.Println("Finished populating tiles db.")
	fmt.Println("db=", db)
	return db
}

// nearest 找出与目标图片平均颜色最接近的瓷砖图片，并返回图片的文件名。
func nearest(target [3]float64, db *map[string][3]float64) string {
	var filename string
	smallest := 1000000.0
	for k, v := range *db {
		dist := distance(target, v)
		if dist < smallest {
			filename, smallest = k, dist
		}
	}
	delete(*db, filename)
	return filename
}

func distance(p1 [3]float64, p2 [3]float64) float64 {
	return math.Sqrt(sq(p2[0]-p1[0])) + math.Sqrt(sq(p2[1]-p1[1])) +
		math.Sqrt(sq(p2[2]-p1[2]))
}

func sq(n float64) float64 {
	return n * n
}

func cloneTilesDB() map[string][3]float64 {
	db := make(map[string][3]float64)
	for k, v := range TILESDB {
		db[k] = v
	}
	return db
}
