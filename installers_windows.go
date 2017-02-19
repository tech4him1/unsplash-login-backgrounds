package main

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func enableBackgrounds() {
	log.Print("Enabling Backgrounds....")
	logonUI, _, err := registry.CreateKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Authentication\LogonUI\Background`, registry.QUERY_VALUE+registry.SET_VALUE)
	if err != nil {
		log.Fatalln(err, "You may need to use the command line parameter '--elevate'.")
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
	// Make sure we are not overwriting a backup that is already there.
	if _, err = os.Stat(fmt.Sprintf("%s.bkp", backgroundLocation)); os.IsNotExist(err) {
		// Backup the curent background by appending `.bkp` to its current filename.
		fileErr := os.Rename(backgroundLocation, fmt.Sprintf("%s.bkp", backgroundLocation))
		// If an error is just saying that there was no image there, ignore it.
		if (fileErr != nil) && (!os.IsNotExist(fileErr)) {
			log.Fatalln(fileErr, "You may need to use the command line parameter '--elevate'.")
		}
	} else {
		log.Fatalln("Background backup already exists!  Are you sure that you have not already enabled this?")
	}
	log.Println("Done.")
}

func runEveryBoot(updateCycle time.Duration, imgCategory string) {
	log.Print("Setting to run every boot....")
	// We don't need to use the --elevate parameter when we are running the task because we are scheduling the task with the highest privileges (it will already be elevated).
	args := []string{"/create", "/sc", "onlogon", "/tn", "unsplash-login-backgrounds", "/rl", "highest", "/tr", fmt.Sprintf("'%s' --time '%v' --type '%v'", exePath, updateCycle, imgCategory)}
	if _, err := exec.Command("schtasks", args...).Output(); err != nil {
		log.Fatal(string(err.(*exec.ExitError).Stderr), "You may need to use the command line parameter '--elevate'.")
	}
	log.Println("Done.")
}

func disableBackgroundUpdater() {
	// If there is a login image backed up, ask the user if they want to restore it (they may just want to keep the current downloaded image without it updating).
	fmt.Print("Would you like to restore the original/OEM image [y/N]? ")
	var restoreImg string
	fmt.Scanln(&restoreImg)
	if restoreImg == "y" || restoreImg == "Y" {
		log.Print("Restoring old background....")
		fileErr := os.Rename(fmt.Sprintf("%s.bkp", backgroundLocation), backgroundLocation)
		if fileErr != nil {
			if os.IsNotExist(fileErr) {
				// If there is no backup image, the original image would be the Windows one, so just delete the current image to restore that.
				fileErr = os.Remove(backgroundLocation)
				if fileErr != nil {
					log.Fatalln(fileErr, "You may need to use the command line parameter '--elevate'.")
				}
			} else {
				log.Fatalln(fileErr, "You may need to use the command line parameter '--elevate'.")
			}
		}
		log.Println("Done.")
	}
}

func disableRunEveryBoot() {
		log.Print("Unsetting to run every boot....")
		args := []string{"/delete", "/tn", "unsplash-login-backgrounds"}
		if _, err := exec.Command("schtasks", args...).Output(); err != nil {
			log.Fatal(string(err.(*exec.ExitError).Stderr), "You may need to use the command line parameter '--elevate'.")
		}
		log.Println("Done.")
}
