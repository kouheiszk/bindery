package main

import (
	"bytes"
	"context"
	"errors"
	"golang.org/x/sync/errgroup"
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

func isSupportedImageExtension(ext string) bool {
	supported := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}
	return supported[strings.ToLower(ext)]
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

func createPageFromImagePath(imagePath string, tempBaseDirectory string) (*Page, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		logError("Couldn't open image: %s", imagePath)
		return nil, err
	}
	defer file.Close()

	b, _ := ioutil.ReadAll(file)

	m, encode, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		logError("Couldn't decode image: %s", imagePath)
		return nil, err
	}

	config, _, err := image.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		logError("Couldn't get image config: %s", imagePath)
		return nil, err
	}

	switch encode {
	case "jpeg":
	case "png":
		// TODO: インターレースのPNGは変換できないので、PNGはJPEGに変換してからPDFにする
		// TODO: PNGのdocoderを取得できればinterlacingかどうか判断できるが
		// TODO: gofpdfがインターレースをサポートするようにプルリクを送るか...
		// TODO: PNGの仕様を調査する
		newImageFile, _ := os.OpenFile(imagePath, os.O_WRONLY|os.O_CREATE, 0600)
		defer newImageFile.Close()
		png.Encode(newImageFile, m)
	default:
		logError("Couldn't read images: %s", encode)
		return nil, errors.New("unsupported image type")
	}

	page := Page{}
	page.ImagePath = imagePath
	page.Width = config.Width
	page.Height = config.Height

	return &page, nil
}

func pagesFromDirectory(directoryPath string, tempBaseDirectory string) ([]Page, error) {
	children, err := filepath.Glob(filepath.Join(directoryPath, "*"))
	if err != nil {
		return nil, err
	}

	imagePaths, err := supportedImagePathsFromPaths(children, false)
	if err != nil {
		return nil, err
	}

	var pages []Page
	for _, imagePath := range imagePaths {
		page, err := createPageFromImagePath(imagePath, tempBaseDirectory)
		if err != nil {
			return nil, err
		}
		pages = append(pages, *page)
	}

	PageSort(pages)

	return pages, nil
}

func convertImagesToPdf(sourceDirectoryPath string, destPath string, tempBaseDirectory string, dispose bool) error {
	logDebug("Start create pdf from: %s", sourceDirectoryPath)
	pages, err := pagesFromDirectory(sourceDirectoryPath, tempBaseDirectory)
	if err != nil {
		return err
	}

	pdf := gofpdf.NewCustom(&gofpdf.InitType{UnitStr: "pt", Size: gofpdf.SizeType{Wd: 0, Ht: 0}})
	for _, page := range pages {
		pdf.AddPageFormat("P", gofpdf.SizeType{Wd: float64(page.Width), Ht: float64(page.Height)})
		pdf.Image(page.ImagePath, 0, 0, float64(page.Width), float64(page.Height), false, "", 0, "")
	}

	logDebug("Output pdf: %s", destPath)

	err = pdf.OutputFileAndClose(destPath)
	if err != nil {
		logError("Couldn't generate pdf: %s", destPath)
		return err
	}

	logInfo("PDF is created: %s", destPath)

	// Dispose source directories if needed
	if dispose {
		logInfo("Dispose source directory: %s", sourceDirectoryPath)
		trashDirectory(sourceDirectoryPath)
	}

	return nil
}

func processConvertImagesToPdf(directoryPaths []string, tempDirectoryPath string, concurrency int, dispose bool) error {
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
				destPath := destPathFromSourceDirectoryPath(directoryPath)
				err := convertImagesToPdf(directoryPath, destPath, tempDirectoryPath, dispose)
				if err != nil {
					cancel()
					return err
				}

				return nil
			}
		})
	}

	err := eg.Wait()
	return err
}
