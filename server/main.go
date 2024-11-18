package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"
)

var maxMemory int64 = 10485760

func main() {
	mux := http.NewServeMux()

	files := http.FileServer(http.Dir("public"))

	mux.Handle("/static/", http.StripPrefix("/static/", files))
	mux.HandleFunc("/", upload)
	mux.HandleFunc("/mosaic", mosaic)

	server := &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: mux,
	}

	TILESDB = tilesDB()
	fmt.Printf("server running at: %v\n", server.Addr)
	server.ListenAndServe()
}

func upload(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("upload.html")
	if err != nil {
		log.Fatalf("parse template error: %v", err)
	}
	t.Execute(w, nil)
}

func mosaic(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()

	r.ParseMultipartForm(maxMemory)

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		log.Fatalf("get form data error: %v", err)
	}
	defer file.Close()
	log.Println("filename=", fileHeader.Filename)
	tileSize, _ := strconv.Atoi(r.FormValue("tile_size"))
	log.Println("tile_size=", tileSize)
	original, _, _ := image.Decode(file)
	bounds := original.Bounds()
	x0, y0 := bounds.Min.X, bounds.Min.Y
	x1, y1 := bounds.Max.X, bounds.Max.Y
	newImage := image.NewNRGBA(image.Rect(x0, y0, x1, y1))

	db := cloneTilesDB()
	sp := image.Point{0, 0}
	for y := bounds.Min.Y; y < bounds.Max.Y; y = y + tileSize {
		for x := bounds.Min.X; x < bounds.Max.X; x = x + tileSize {
			r, g, b, _ := original.At(x, y).RGBA()
			color := [3]float64{float64(r), float64(g), float64(b)}

			nearest := nearest(color, &db)
			if nearest == "" {
				nearest = "tiles/cat1.jpg"
			}
			file, err := os.Open(nearest)
			if err != nil {
				log.Fatalf("open file error: %v", err)
			}
			defer file.Close()
			img, _, err := image.Decode(file)
			if err != nil {
				log.Fatalf("decode image error: %v", err)
			}
			t := resize(img, tileSize)
			tile := t.SubImage(t.Bounds())
			tileBounds := image.Rect(x, y, x+tileSize, y+tileSize)
			draw.Draw(newImage, tileBounds, tile, sp, draw.Src)

		}
	}

	buf1 := new(bytes.Buffer)
	jpeg.Encode(buf1, original, nil)
	originalStr := base64.StdEncoding.EncodeToString(buf1.Bytes())

	buf2 := new(bytes.Buffer)
	jpeg.Encode(buf2, newImage, nil)

	mosaic := base64.StdEncoding.EncodeToString(buf2.Bytes())
	t1 := time.Now()
	images := map[string]string{
		"original": originalStr,
		"mosaic":   mosaic,
		"duration": fmt.Sprintf("%v ", t1.Sub(t0)),
	}

	t, _ := template.ParseFiles("results.html")
	t.Execute(w, images)

}
