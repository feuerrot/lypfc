package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"strings"

	"golang.org/x/image/bmp"
	"golang.org/x/image/webp"
)

func MetaImage(file string, offX, offY int) (func(int, int) (int, int, int, bool), error) {
	data, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("can't read file %s: %v", file, err)
	}

	var img image.Image
	parts := strings.Split(file, ".")
	ext := strings.ToLower(parts[len(parts)-1])
	if ext == "png" {
		img, err = png.Decode(data)
	} else if ext == "bmp" {
		img, err = bmp.Decode(data)
	} else if ext == "webp" {
		img, err = webp.Decode(data)
	} else if ext == "jpg" || ext == "jpeg" {
		img, err = jpeg.Decode(data)
	} else {
		err = fmt.Errorf("unknown extension %s", ext)
	}
	if err != nil {
		return nil, fmt.Errorf("can't decode image: %v", err)
	}

	f := func(x, y int) (int, int, int, bool) {
		max := img.Bounds().Max
		if x > max.X+offX ||
			y > max.Y+offY ||
			x < offX ||
			y < offY {
			return 0, 0, 0, false
		}
		r, g, b, a := img.At(x-offX, y-offY).RGBA()
		return int(r / 256), int(g / 256), int(b / 256), bool(a != 0)
	}

	return f, nil
}
