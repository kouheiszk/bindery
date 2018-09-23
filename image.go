package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"math"
	"math/rand"
	"path/filepath"
	"runtime"
	"time"
)

func isSupportedImageExtension(ext string) bool {
	supported := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}
	return supported[ext]
}

func isOnlySupportedImages(paths []string) bool {
	for _, path := range paths {
		if !isSupportedImageExtension(filepath.Ext(path)) {
			return false
		}
	}
	return true
}

func convertImagesToPdf(directoryPath string) error {
	fmt.Println(directoryPath)
	time.Sleep(1600 * time.Millisecond)
	if rand.Intn(3) == 1 {
		return errors.New("error")
	}

	return nil
}

func processConvertImagesToPdf(directoryPaths []string, concurrency int) error {
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
				err := convertImagesToPdf(directoryPath)
				if err != nil {
					cancel()
				}
				return err
			}
		})
	}

	err := eg.Wait()
	return err
}
