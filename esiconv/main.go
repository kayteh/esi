package main

import (
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path"

	"github.com/kayteh/esi"
)

func main() {
	convPath := os.Args[1]
	outPath := os.Args[2]

	fin, err := os.Open(convPath)
	if err != nil {
		log.Fatalln(err)
	}

	fout, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalln(err)
	}

	img, _, err := image.Decode(fin)
	if err != nil {
		log.Fatalln(err)
	}

	fin.Close()

	ext := path.Ext(outPath)
	switch ext[1:] {
	case "esi":
		err = esi.Encode(fout, img)
	case "png":
		err = png.Encode(fout, img)
	case "jpg":
		fallthrough
	case "jpeg":
		err = jpeg.Encode(fout, img, nil)
	case "gif":
		err = gif.Encode(fout, img, nil)
	default:
		log.Fatalf("extension %s unknown", ext)
	}

	if err != nil {
		log.Fatalln(err)
	}

	fout.Close()
}
