package main

import (
	"os"
	"path/filepath"
)

var backgroundLocation string = filepath.Join(os.Getenv("windir"), "system32/oobe/info/backgrounds", "backgroundDefault.jpg")
var exePath, _ = filepath.Abs(os.Args[0])
var imgCategories = []string{"buildings", "food", "nature", "people", "technology", "objects"}

func validCategory(cat string) bool {
    for _, v := range imgCategories {
        if cat == v {
            return true
        }
    }
    return false
}
