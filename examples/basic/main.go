package main

import (
	"fmt"
	"log"

	"github.com/zenith110/magstripe-go"
)

func main() {
	// Connect to MSR device
	// Use "/dev/ttyUSB0" on Linux/Mac or "COM1" on Windows
	device, err := magstripe.NewMSR("/dev/ttyUSB0")
	if err != nil {
		log.Fatal("Failed to connect to device:", err)
	}
	defer device.Close()

	// Read tracks in ISO format
	fmt.Println("Reading tracks...")
	tracks, err := device.ReadTracks()
	if err != nil {
		log.Fatal("Failed to read tracks:", err)
	}

	fmt.Printf("Track 1: %s\n", tracks.Track1)
	fmt.Printf("Track 2: %s\n", tracks.Track2)
	fmt.Printf("Track 3: %s\n", tracks.Track3)
}
