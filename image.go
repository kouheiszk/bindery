package main

import "path/filepath"

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
