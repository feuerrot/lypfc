package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"

	"golang.org/x/image/bmp"
	"golang.org/x/image/webp"
)

func bmpImage(file string) (image.Image, error) {
	data, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("can't read file %s: %v", file, err)
	}

	img, err := bmp.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("can't decode image: %v", err)
	}

	return img, nil
}

func pngImage(file string) (image.Image, error) {
	data, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("can't read file %s: %v", file, err)
	}

	img, err := png.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("can't decode image: %v", err)
	}

	return img, nil
}

func webpImage(file string) (image.Image, error) {
	data, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("can't read file %s: %v", file, err)
	}

	img, err := webp.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("can't decode image: %v", err)
	}

	return img, nil
}

func jpgImage(file string) (image.Image, error) {
	data, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("can't read file %s: %v", file, err)
	}

	img, err := jpeg.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("can't decode image: %v", err)
	}

	return img, nil
}

func MetaImage(file string, offX, offY int) func(int, int) (int, int, int, bool) {
	parts := strings.Split(file, ".")
	ext := parts[len(parts)-1]

	var img image.Image
	var err error
	if ext == "png" {
		img, err = pngImage(file)
	} else if ext == "bmp" {
		img, err = bmpImage(file)
	} else if ext == "webp" {
		img, err = webpImage(file)
	} else if ext == "jpg" || ext == "jpeg" {
		img, err = jpgImage(file)
	} else {
		err = fmt.Errorf("unknown extension %s", ext)
	}
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	f := func(x, y int) (int, int, int, bool) {
		max := img.Bounds().Max
		if x > max.X+offX ||
			y > max.Y+offY ||
			x < offX ||
			y < offY {
			return 0, 0, 0, false
		}
		c := color.RGBAModel.Convert(img.At(x-offX, y-offY))
		r, g, b, a := c.RGBA()
		return int(r / 255), int(g / 255), int(b / 255), bool(a != 0)
	}

	return f
}
