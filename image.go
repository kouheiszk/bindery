package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"image"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	_ "image/jpeg"
	_ "image/png"

	"github.com/jung-kurt/gofpdf"
)

type Image struct {
	Path   string
	Width  int
	Height int
}

type Pdf struct {
	Path                string
	SourceDirectoryPath string
}

func isSupportedImageExtension(ext string) bool {
	supported := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}
	return supported[ext]
}

func supportedImagePathsFromPaths(paths []string, shouldNotBeContainDirectory bool) ([]string, error) {
	var images []string
	for _, path := range paths {
		state, err := os.Stat(path)
		if err != nil || state.IsDir() {
			if shouldNotBeContainDirectory {
				return nil, errors.New("target directory should not be contain directory")
			}
			continue
		}

		if isSupportedImageExtension(filepath.Ext(path)) {
			images = append(images, path)
		}
	}
	return images, nil
}

func isOnlySupportedImages(paths []string) bool {
	if len(paths) == 0 {
		return false
	}

	imagePaths, err := supportedImagePathsFromPaths(paths, true)
	if err != nil {
		return false
	}

	return len(imagePaths) > 0
}

func getImageSize(imagePath string) (int, int) {
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		logError("Couldn't get image config: %s", imagePath)
		criticalError(err)
	}
	return config.Width, config.Height
}

func imagesFromDirectory(directoryPath string) ([]Image, error) {
	children, err := filepath.Glob(filepath.Join(directoryPath, "*"))
	if err != nil {
		return nil, err
	}

	imagePaths, err := supportedImagePathsFromPaths(children, false)
	if err != nil {
		return nil, err
	}

	sort.Strings(imagePaths)

	var images []Image
	for _, imagePath := range imagePaths {
		image := Image{}
		image.Path = imagePath
		image.Width, image.Height = getImageSize(imagePath)
		images = append(images, image)
	}

	return images, nil
}

func convertImagesToPdf(directoryPath string, tempBaseDirectory string) (*Pdf, error) {
	images, err := imagesFromDirectory(directoryPath)
	if err != nil {
		return nil, err
	}

	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "pt",
		Size:    gofpdf.SizeType{Wd: 0, Ht: 0},
	})
	for _, image := range images {
		pdf.AddPageFormat("P", gofpdf.SizeType{Wd: float64(image.Width), Ht: float64(image.Height)})
		pdf.Image(image.Path, 0, 0, float64(image.Width), float64(image.Height), false, "", 0, "")
	}
	pdfPath := filepath.Join(tempBaseDirectory, filepath.Base(directoryPath)+".pdf")
	err = pdf.OutputFileAndClose(pdfPath)
	if err != nil {
		return nil, err
	}

	logDebug("Create pdf: %s", pdfPath)

	return &Pdf{Path: pdfPath, SourceDirectoryPath: directoryPath}, nil
}

func processConvertImagesToPdf(directoryPaths []string, tempDirectory string, concurrency int) ([]Pdf, error) {
	var output []Pdf
	eg, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	sem := make(chan int, int(math.Min(float64(runtime.NumCPU()), float64(concurrency))))
	for _, directoryPath := range directoryPaths {
		directoryPath := directoryPath
		eg.Go(func() error {
			sem <- 1
			defer func() {
				<-sem
			}()

			select {
			case <-ctx.Done():
				return nil
			default:
				pdf, err := convertImagesToPdf(directoryPath, tempDirectory)
				if err != nil {
					cancel()
				}
				output = append(output, *pdf)
				return err
			}
		})
	}

	err := eg.Wait()
	return output, err
}
