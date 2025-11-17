package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/zenith110/magstripe-go"
)

func main() {
	// Connect to MSR device
	device, err := magstripe.NewMSR("/dev/ttyUSB0")
	if err != nil {
		log.Fatal("Failed to connect to device:", err)
	}
	defer device.Close()

	// Set bits per character for raw mode
	fmt.Println("Setting BPC to 8,8,8...")
	err = device.SetBPC(8, 8, 8)
	if err != nil {
		log.Fatal("Failed to set BPC:", err)
	}

	// Read tracks in raw format
	fmt.Println("Reading raw tracks...")
	s1, s2, s3, err := device.ReadRawTracks()
	if err != nil {
		log.Fatal("Failed to read raw tracks:", err)
	}

	// Unpack raw data for each track
	fmt.Println("Track 1 (raw):")
	result1 := magstripe.UnpackRaw(s1, magstripe.Track1Map, 6, 8)
	printRawResult(1, result1)

	fmt.Println("Track 2 (raw):")
	result2 := magstripe.UnpackRaw(s2, magstripe.Track23Map, 4, 8)
	printRawResult(2, result2)

	fmt.Println("Track 3 (raw):")
	result3 := magstripe.UnpackRaw(s3, magstripe.Track23Map, 4, 8)
	printRawResult(3, result3)
}

func printRawResult(trackNum int, res magstripe.RawData) {
	line := fmt.Sprintf("  Data: %s", res.Data)
	if len(res.Data) != res.TotalLength {
		line += fmt.Sprintf(" (+%d null bytes)", res.TotalLength-len(res.Data))
	}
	if res.LRCError {
		line += " (LRC error detected)"
	}
	fmt.Println(line)

	if strings.Contains(res.ParityErrors, "^") {
		fmt.Printf("  Parity errors: %s\n", res.ParityErrors)
	}
	fmt.Println()
}
