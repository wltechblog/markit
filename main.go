package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"io"
	"log"
	"math"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imageorient"
	"github.com/disintegration/imaging"
)

var dapp fyne.App

func main() {
// We need a few things to be available before we start creating objects
	original := canvas.NewImageFromResource(resourceNoPng)
	watermark := canvas.NewImageFromResource(resourceNoPng)
	var originalImage, watermarkImage image.Image
	var err error

	dapp = app.NewWithID("net.wanderlounge.markit")
	w := dapp.NewWindow("markit")
	c := container.NewVBox()

	imageChosen := widget.NewLabel("Image: Not Chosen")
	imageButton := widget.NewButton("Choose Image", func() {
		d := dialog.NewFileOpen(func(f fyne.URIReadCloser, err error) {
			if err != nil || f == nil {
				return
			}
			originalImage, _, err = imageorient.Decode(f)
			if err != nil {
				dialog.NewError(fmt.Errorf("can't decode image: %s", err), w).Show()
				return
			}
			for k := range c.Objects {
				if c.Objects[k] == original {
					n := canvas.NewImageFromImage(originalImage)
					n.FillMode = canvas.ImageFillContain
					n.SetMinSize(fyne.Size{Width: 100, Height: 100})
					n.Show()
					c.Objects[k] = n
					original = n
					c.Refresh()
					imageChosen.Hide()
				}
			}
		}, w)
		var filters []string
		filters = append(filters, ".jpg", ".jpeg")
		d.SetFilter(storage.NewExtensionFileFilter(filters))
		d.Show()
	})

	wmChosen := widget.NewLabel("Watermark: Not Chosen")
	wmButton := widget.NewButton("Choose Watermark", func() {
		d := dialog.NewFileOpen(func(f fyne.URIReadCloser, e error) {
			if e != nil || f == nil {
				return
			}
			for k := range c.Objects {
				if c.Objects[k] == watermark {
					var err error
					wmData, err := io.ReadAll(f)
					if err != nil {
						dialog.NewError(fmt.Errorf("can't read watermark: %s", err), w).Show()
						return

					}
					watermarkImage, _, err = image.Decode(strings.NewReader(string(wmData)))
					if err != nil {
						dialog.NewError(fmt.Errorf("can't decode watermark: %s", err), w).Show()
						return

					}

					n := canvas.NewImageFromImage(watermarkImage)
					n.FillMode = canvas.ImageFillContain
					n.SetMinSize(fyne.Size{Width: 100, Height: 100})
					n.Show()
					c.Objects[k] = n
					watermark = n
					c.Refresh()
					dapp.Preferences().SetString("watermark", base64.StdEncoding.EncodeToString(wmData))

				}
			}

			watermark = canvas.NewImageFromReader(f, "Watermark")
			watermark.Show()
			wmChosen.Hide()

		}, w)
		var filters []string
		filters = append(filters, ".png")
		d.SetFilter(storage.NewExtensionFileFilter(filters))
		d.Show()
	})

	goButton := widget.NewButton("Go", func() {
		if original == nil || watermark == nil {
			dialog.NewError(fmt.Errorf("images not selected"), w).Show()
			return
		}
		err := mark(originalImage, watermarkImage, w)
		if err != nil {
			dialog.NewError(err, w).Show()
		} 
	})
	watermark.Hide()
	original.Hide()
	c.Add(imageButton)
	c.Add(imageChosen)
	c.Add(original)
	c.Add(wmButton)
	c.Add(wmChosen)
	c.Add(watermark)
	c.Add(goButton)

	w.SetContent(c)
	w.Resize(fyne.Size{Width: 800, Height: 600})
	original.SetMinSize(fyne.Size{Width: 100, Height: 100})

	// Storing the watermark image in base64 encoding in the preferences API seems wasteful but
	// seems to be the best way to handle Android's premissions complexity... Storing the URI works
	// on other platforms...

	if dapp.Preferences().String("watermark") != "" {
		tmp := dapp.Preferences().String("watermark")
		wm, _ := base64.StdEncoding.DecodeString(tmp)
		for k := range c.Objects {
			if c.Objects[k] == watermark {
				watermarkImage, _, err = image.Decode(bytes.NewReader(wm))
				if err != nil {
					log.Println(err)
				} else {

				n := canvas.NewImageFromImage(watermarkImage)
				n.FillMode = canvas.ImageFillContain
				n.SetMinSize(fyne.Size{Width: 100, Height: 100})
				n.Show()
				c.Objects[k] = n
				watermark = n
				c.Refresh()
				wmChosen.Hide()
				}
			}
		}
	} else {
		log.Println("No preference value")
	}

	w.ShowAndRun()
}

func mark(original, watermark image.Image, w fyne.Window) error {
	// Get aspect ratio of watermark
	ratio := float64(watermark.Bounds().Dy()) / float64(watermark.Bounds().Dx())

	// Calculate watermark size to be about 33% of the width of the original image
	wmWidth := int(math.Round(float64(original.Bounds().Dx()) * 0.33))
	wmHeight := int(math.Round(float64(wmWidth) * ratio))
	mark := imaging.Resize(watermark, wmWidth, wmHeight, imaging.Lanczos)

	// OK now where do we place it? 60% across and 80% down
	b := original.Bounds()
	offset := image.Pt(int(math.Round(float64((b.Dx()))*0.95))-wmWidth, int(math.Round(float64(b.Dy()))*0.95)-wmHeight)

	output := image.NewRGBA(b)
	draw.Draw(output, b, original, image.Point{X: 0, Y: 0}, draw.Src)
	draw.Draw(output, mark.Bounds().Add(offset), mark, image.Point{X: 0, Y: 0}, draw.Over)

	save := dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.NewError(err, w).Show()
			return
		}
		defer uc.Close()
		err = jpeg.Encode(uc, output, nil)
		if err != nil {
			dialog.NewError(fmt.Errorf("writing output:\n %s:\n %s", uc.URI().String(), err), w).Show()
			return
		}
		dapp.Preferences().SetString("savelocation", uc.URI().Path())
	}, w)
	if dapp.Preferences().String("savelocation") != "" {
		tmp, err := storage.ParseURI(dapp.Preferences().String("savelocation"))
		if err == nil {
			lister, err := storage.ListerForURI(tmp)
			if err == nil {
				save.SetLocation(lister)
			}

		}
	}
	save.SetFileName(fmt.Sprintf("WM_%s.jpg", time.Now().Format(time.DateTime)))
	save.Show()

	return nil
}
