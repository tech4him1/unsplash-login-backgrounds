package main

import (
	"flag"
	"fmt"
	"github.com/tech4him1/elevate"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"strings"
)

func main() {
	imgCategory := flag.String("type", "nature", fmt.Sprintf("Image category (Options: %v).", strings.Join(imgCategories, ", ")))
	updateCycle := flag.Duration("time", time.Hour, "Image update cycle.")
	enable := flag.Bool("enable", false, "Enable custom login backgrounds on this computer.")
	elevateFlag := flag.Bool("elevate", false, "Run with admin privileges if necessary (can create UAC prompt).")
	flag.Parse()

	if !validCategory(*imgCategory) {
		log.Fatalf("Invalid category %v.  Valid categories: %v.", *imgCategory, strings.Join(imgCategories, ", "))
	}

	if *enable == true {
		// If we are already elevated (to admin privileges), run the enable functions.  If we are not elevated, elevate.
		if *elevateFlag == true {
			elevate.Elevate(exePath, fmt.Sprintf("--enable --time '%v' --type '%v'", updateCycle, imgCategory), "")
			return
		} else {
			enableBackgrounds()
			runEveryBoot(*updateCycle, *imgCategory)
			return
		}
	}

	for {
		updateBackground(*imgCategory)
		time.Sleep(*updateCycle)
	}
}

func updateBackground(imgCategory string) {
	imgFile, fileErr := os.Create(backgroundLocation)
	if fileErr != nil {
		log.Fatalln(fileErr, "You may need to use the command line parameter '--elevate'.")
	}
	defer imgFile.Close()

	imgDownload, downlErr := http.Get(fmt.Sprintf("https://source.unsplash.com/category/%s", imgCategory))
	if downlErr != nil {
		log.Println(downlErr)
		updateBackground(imgCategory) // Try again.
		return
	} else if imgDownload.Header["Content-Type"][0] != "image/jpeg" {
		log.Println("Image must be a jpeg.", "Trying again....")
		updateBackground(imgCategory) // Try again.
		return
	}
	_, saveErr := io.Copy(imgFile, imgDownload.Body)
	if saveErr != nil {
		log.Fatalln(saveErr, "You may need to use the command line parameter '--elevate'.")
	}
	imgDownload.Body.Close()

	log.Print("Image updated.")
}
