package main

import (
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type Page struct {
	ImagePath string
	Width     int
	Height    int
}

type sorter []Page

func (s sorter) Len() int {
	return len(s)
}

func (s sorter) Swap(i, j int) {
	s[j], s[i] = s[i], s[j]
}

func (s sorter) Bytes(i int) []byte {
	return []byte(s[i].ImagePath)
}

func PageSort(pages []Page) {
	c := collate.New(language.English)
	c.Sort(sorter(pages))
}
