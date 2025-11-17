package main

import (
	"fmt"
	"log"

	"github.com/zenith110/magstripe-go"
)

func main() {
	// Connect to MSR device
	device, err := magstripe.NewMSR("/dev/ttyUSB0")
	if err != nil {
		log.Fatal("Failed to connect to device:", err)
	}
	defer device.Close()

	// Set high coercivity for writing
	fmt.Println("Setting high coercivity...")
	err = device.SetCoercivity(magstripe.HiCo)
	if err != nil {
		log.Fatal("Failed to set coercivity:", err)
	}

	// Write sample data to all tracks
	track1Data := "%B1234567890123445^DOE/JOHN^49121010000000000000?"
	track2Data := ";1234567890123445=49121010000000000?"
	track3Data := ";011234567890123445=724724100000000000000000000000000000000000000000000000000000000000000000?"

	fmt.Println("Writing tracks...")
	err = device.WriteTracks(track1Data, track2Data, track3Data)
	if err != nil {
		log.Fatal("Failed to write tracks:", err)
	}

	fmt.Println("Successfully wrote data to all tracks!")

	// Read back to verify
	fmt.Println("Reading back to verify...")
	tracks, err := device.ReadTracks()
	if err != nil {
		log.Fatal("Failed to read back tracks:", err)
	}

	fmt.Printf("Track 1: %s\n", tracks.Track1)
	fmt.Printf("Track 2: %s\n", tracks.Track2)
	fmt.Printf("Track 3: %s\n", tracks.Track3)
}
