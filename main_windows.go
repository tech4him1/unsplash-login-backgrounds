package main

import (
	"flag"
	"fmt"
	"github.com/tech4him1/elevate"
	"golang.org/x/sys/windows/registry"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"strings"
)

var backgroundLocation string = filepath.Join(os.Getenv("windir"), "system32/oobe/info/backgrounds", "backgroundDefault.jpg")
var exePath, _ = filepath.Abs(os.Args[0])
var imgCategories = []string{"buildings", "food", "nature", "people", "technology", "objects"}

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

func enableBackgrounds() {
	log.Print("Enabling Backgrounds....")
	logonUI, _, err := registry.CreateKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Authentication\LogonUI\Background`, registry.QUERY_VALUE+registry.SET_VALUE)
	if err != nil {
		log.Fatalln(err, "You may need to use the command line parameter '--elevate'.")
	}
	if val, _, _ := logonUI.GetIntegerValue("OEMBackground"); val == 1 {
		log.Fatalln("Backgrounds already enabled!  Either the OEM has enabled login backgrounds, or you are using another login background tool.")
	}
	err = logonUI.SetDWordValue("OEMBackground", 1)
	if err != nil {
		log.Fatalln(err, "You may need to use the command line parameter '--elevate'.")
	}
	logonUI.Close()
	log.Println("Done.")

	log.Print("Setting up background folder....")
	dirErr := os.MkdirAll(filepath.Dir(backgroundLocation), os.ModePerm)
	if dirErr != nil {
		log.Fatalln(dirErr, "You may need to use the command line parameter '--elevate'.")
	}
	log.Println("Done.")

	log.Print("Backing up old background....")
	fileErr := os.Rename(backgroundLocation, fmt.Sprintf("%s.bkp", backgroundLocation))
	if (fileErr != nil) && (!os.IsNotExist(fileErr)) {
		log.Fatalln(fileErr, "You may need to use the command line parameter '--elevate'.")
	}
	log.Println("Done.")
}

func runEveryBoot(updateCycle time.Duration, imgCategory string) {
	log.Print("Setting to run every boot....")
	// We don't need to use the --elevate parameter here because we are scheduling the task with the highest privileges.
	args := []string{"/create", "/sc", "onlogon", "/tn", "unsplash-login-backgrounds", "/rl", "highest", "/tr", fmt.Sprintf("'%s' --time '%v' --type '%v'", exePath, updateCycle, imgCategory)}
	if _, err := exec.Command("schtasks", args...).Output(); err != nil {
		log.Fatal(string(err.(*exec.ExitError).Stderr))
	}
	log.Println("Done.")
}

func validCategory(cat string) bool {
    for _, v := range imgCategories {
        if cat == v {
            return true
        }
    }
    return false
}
