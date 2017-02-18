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
