package esi

/*
>                            Extremely Simple Image file format                            <
>------------------------------------------------------------------------------------------<
> Designed for databending or glitching
> Has very little fancy features that could cause problems with decoding
> Decoder is designed to assume, without any penalties if it assumes wrong
> Width parameter is assumed to be the square root of the length of the data, rounded up
> Color format is assumed to be 8 bit color
> Can support grayscale or color in a variety of bit depths
> There is no data between or around pixels, or at the end of the file
> It is literally a sequence of bits after the minimal header
> This should prevent any issues with databending
> Header is encoded with 16 bits for width, 1 indexed, for a maximum width of 65536 pixels
> All data is big-endian encoded
> This is followed by the following 5 bits which is followed by 11 ignored bits:
>     | 0
>     |  0000 1 bit black and white
>     |  #### 4 bits encoding bit depth, 1 indexed
>     -------------------------------
>     | 1
>     |  #### 4 bits encoding bit depth per channel, 1 indexed
>
> Sample header:
>     | 01100101 01110011 01101001 00110001 ( esi1 in ascii binary )
>     | 00000000 00010000 10100000 00000000
>
>     | 00000000
>     |          00010000 width of 32
>     |                   1 color
>     |                    0100 8 bit depth
>     |                        000 00000000 ignored bits
>     |                                   ( to ease databending the entire image )
>     |                                   (         without affecting the header )
*/

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"io"
	"log"
)

var (
	errImageInvalid = errors.New("esi: image is invalid")
	errNotESI       = errors.New("esi: not an esi")
)

func init() {
	image.RegisterFormat("esi", "esi1", Decode, DecodeConfig)
}

func Decode(r io.Reader) (o image.Image, err error) {
	buf := bytes.Buffer{}
	io.Copy(&buf, r)

	cfgBuf := bytes.NewBuffer(buf.Bytes())
	cfg, err := DecodeConfig(cfgBuf)
	if err != nil {
		return o, err
	}

	log.Println(cfg)

	bounds := image.Rect(0, 0, cfg.Width, cfg.Height)

	log.Println(bounds)

	img := image.NewRGBA(bounds)

	buf.Next(8)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {

			r, e1 := buf.ReadByte()
			g, e2 := buf.ReadByte()
			b, e3 := buf.ReadByte()

			for _, err := range []error{e1, e2, e3} {
				// log.Println(err)
				if err != nil && err != io.EOF {
					return o, err
				}
			}

			c := color.RGBA{
				R: r,
				G: g,
				B: b,
				A: 255,
			}

			img.SetRGBA(x, y, c)
			// log.Println(x, y, c)
			// log.Println(img.At(x, y).RGBA())
		}
	}

	return img, nil
}

func DecodeConfig(r io.Reader) (cfg image.Config, err error) {
	// pull first 8 bytes out...
	buf := bytes.Buffer{}
	if err != nil {
		return cfg, err
	}

	buf.ReadFrom(r)

	// check first 4 == esi1
	magic := buf.Next(4)
	if !bytes.Equal(magic, []byte("esi1")) {
		return cfg, errNotESI
	}

	// read headers
	header := buf.Next(4)
	w := binary.BigEndian.Uint16(header[0:2]) // width is first two bytes (16 bit)
	cfg.Width = int(w)

	cfg.ColorModel = GetColorModel(header[2])

	// calculate length as a product of width / length of reader

	size := buf.Len()
	cfg.Height = size / 3 / cfg.Width

	// log.Println("size w/o headers", size)
	// log.Println("height:", cfg.Height)
	// log.Println("width:", cfg.Width)

	return cfg, nil
}

func Encode(w io.Writer, img image.Image) error {
	bounds := img.Bounds()

	_, err := w.Write([]byte("esi1"))
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.BigEndian, uint16(bounds.Max.X))
	if err != nil {
		return err
	}

	cm := img.ColorModel()
	cc := true
	cw := 8

	switch cm {
	case color.GrayModel:
		fallthrough
	case color.Gray16Model:
		cc = false
	}

	switch cm {
	case color.RGBA64Model:
		fallthrough
	case color.Gray16Model:
		cw = 16
	}

	configByte := byte(0)

	if cc {
		configByte |= 128
	}

	if cw == 16 {
		configByte |= 64
	} else {
		configByte |= 32
	}

	_, err = w.Write([]byte{configByte, 0})
	if err != nil {
		return err
	}

	for y := 0; y < bounds.Max.Y; y++ {
		for x := 0; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			_, err = w.Write([]byte{byte(r / 257), byte(g / 257), byte(b / 257)})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetColorModel takes the encoding data byte of an ESI header,
// and outputs the corresponding color.Model
func GetColorModel(b byte) (m color.Model) {
	cc := b&(1<<7)>>7 == 1 // color or not is first bit
	depth := b &^ 128 >> 2 // unshift first bit, read next few

	return color.RGBAModel

	if cc {
		if depth == 16 {
			m = color.RGBA64Model
		} else {
			m = color.RGBAModel
		}
	} else {
		if depth == 16 {
			m = color.Gray16Model
		} else {
			m = color.GrayModel
		}
	}

	return m
}
