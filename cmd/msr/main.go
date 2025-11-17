package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/abrahan/magstripe-go"
)

func main() {
	var (
		read   = flag.Bool("r", false, "read magnetic tracks")
		write  = flag.Bool("w", false, "write magnetic tracks")
		erase  = flag.Bool("e", false, "erase magnetic tracks")
		hico   = flag.Bool("C", false, "select high coercivity mode")
		loco   = flag.Bool("c", false, "select low coercivity mode")
		bpi    = flag.String("b", "", "bit per inch for each track (h or l)")
		device = flag.String("d", "", "path to serial communication device")
		raw    = flag.Bool("0", false, "do not use ISO encoding/decoding")
		tracks = flag.String("t", "123", "select tracks (1, 2, 3, 12, 23, 13, 123)")
		bpc    = flag.String("B", "", "bit per character for each track (5 to 8)")
		help   = flag.Bool("help", false, "show help")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [data...]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Driver for the magnetic strip card reader/writer MSR605\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -d /dev/ttyUSB0 -r                    # read all tracks\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -d COM1 -r -t 12                     # read tracks 1&2 (Windows)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -d /dev/ttyUSB0 -w -t 123 \"t1\" \"t2\" \"t3\"  # write tracks\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -d /dev/ttyUSB0 -e -t 123             # erase all tracks\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -d /dev/ttyUSB0 -C                    # set high coercivity\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -d /dev/ttyUSB0 -c                    # set low coercivity\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -d /dev/ttyUSB0 -b hhl                # set BPI: high, high, low\n", os.Args[0])
	}

	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	// Count mutually exclusive operations
	opCount := 0
	if *read {
		opCount++
	}
	if *write {
		opCount++
	}
	if *erase {
		opCount++
	}
	if *hico {
		opCount++
	}
	if *loco {
		opCount++
	}
	if *bpi != "" {
		opCount++
	}

	if opCount != 1 {
		fmt.Fprintf(os.Stderr, "Error: Must specify exactly one operation (-r, -w, -e, -C, -c, or -b)\n\n")
		flag.Usage()
		os.Exit(1)
	}

	data := flag.Args()

	// Validate arguments
	if (*read || *erase) && len(data) != 0 {
		fmt.Fprintf(os.Stderr, "Error: too many arguments for read/erase operation\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if *write && len(data) != len(*tracks) {
		fmt.Fprintf(os.Stderr, "Error: number of data arguments must match number of tracks\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Parse tracks
	trackFlags := [3]bool{false, false, false}
	trackData := [3]string{"", "", ""}

	for i, trackChar := range *tracks {
		n, err := strconv.Atoi(string(trackChar))
		if err != nil || n < 1 || n > 3 || trackFlags[n-1] {
			fmt.Fprintf(os.Stderr, "Error: invalid tracks specification '%s'\n\n", *tracks)
			flag.Usage()
			os.Exit(1)
		}
		trackFlags[n-1] = true
		if *write && i < len(data) {
			trackData[n-1] = data[i]
		}
	}

	// Parse BPC
	bpc1, bpc2, bpc3 := 8, 8, 8
	if *bpc != "" {
		if len(*bpc) != 3 {
			fmt.Fprintf(os.Stderr, "Error: BPC must be 3 characters (e.g., '888')\n")
			os.Exit(1)
		}
		var err error
		if bpc1, err = strconv.Atoi(string((*bpc)[0])); err != nil || bpc1 < 5 || bpc1 > 8 {
			fmt.Fprintf(os.Stderr, "Error: invalid BPC format, must be 5-8\n")
			os.Exit(1)
		}
		if bpc2, err = strconv.Atoi(string((*bpc)[1])); err != nil || bpc2 < 5 || bpc2 > 8 {
			fmt.Fprintf(os.Stderr, "Error: invalid BPC format, must be 5-8\n")
			os.Exit(1)
		}
		if bpc3, err = strconv.Atoi(string((*bpc)[2])); err != nil || bpc3 < 5 || bpc3 > 8 {
			fmt.Fprintf(os.Stderr, "Error: invalid BPC format, must be 5-8\n")
			os.Exit(1)
		}
	} else if *raw {
		*bpc = "888" // force setup for raw mode
	}

	// Parse BPI
	var bpi1, bpi2, bpi3 *bool
	if *bpi != "" {
		if len(*bpi) != 3 {
			fmt.Fprintf(os.Stderr, "Error: BPI must be 3 characters (e.g., 'hhl')\n")
			os.Exit(1)
		}
		for _, char := range *bpi {
			if char != 'h' && char != 'l' {
				fmt.Fprintf(os.Stderr, "Error: BPI characters must be 'h' or 'l'\n")
				os.Exit(1)
			}
		}
		val1 := (*bpi)[0] != 'l'
		val2 := (*bpi)[1] != 'l'
		val3 := (*bpi)[2] != 'l'
		bpi1 = &val1
		bpi2 = &val2
		bpi3 = &val3
	}

	// Connect to device
	if *device == "" {
		fmt.Fprintf(os.Stderr, "Error: device path required (-d)\n\n")
		flag.Usage()
		os.Exit(1)
	}

	dev, err := magstripe.NewMSR(*device)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to connect to device: %v\n", err)
		os.Exit(1)
	}
	defer dev.Close()

	// Execute operations
	if err := executeOperation(dev, *read, *write, *erase, *hico, *loco, *raw, *bpi != "",
		trackFlags, trackData, bpc1, bpc2, bpc3, bpi1, bpi2, bpi3, *bpc != ""); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func executeOperation(dev *magstripe.MSR, read, write, erase, hicoOp, locoOp, raw, bpiOp bool,
	trackFlags [3]bool, trackData [3]string, bpc1, bpc2, bpc3 int,
	bpi1, bpi2, bpi3 *bool, setBPC bool) error {

	// Set BPC if needed
	if setBPC {
		if err := dev.SetBPC(bpc1, bpc2, bpc3); err != nil {
			return fmt.Errorf("failed to set BPC: %w", err)
		}
	}

	switch {
	case read && raw:
		s1, s2, s3, err := dev.ReadRawTracks()
		if err != nil {
			return fmt.Errorf("failed to read raw tracks: %w", err)
		}

		printResult := func(num int, res magstripe.RawData) {
			line := fmt.Sprintf("%d=%s", num, res.Data)
			if len(res.Data) != res.TotalLength {
				line += fmt.Sprintf(" (+%d null)", res.TotalLength-len(res.Data))
			}
			if res.LRCError {
				line += " (LRC error)"
			}
			fmt.Println(line)
			if strings.Contains(res.ParityErrors, "^") {
				fmt.Printf("  %s <- parity errors\n", res.ParityErrors)
			}
		}

		if trackFlags[0] {
			printResult(1, magstripe.UnpackRaw(s1, magstripe.Track1Map, 6, bpc1))
		}
		if trackFlags[1] {
			printResult(2, magstripe.UnpackRaw(s2, magstripe.Track23Map, 4, bpc2))
		}
		if trackFlags[2] {
			printResult(3, magstripe.UnpackRaw(s3, magstripe.Track23Map, 4, bpc3))
		}

	case read: // ISO mode
		tracks, err := dev.ReadTracks()
		if err != nil {
			return fmt.Errorf("failed to read tracks: %w", err)
		}

		if trackFlags[0] {
			fmt.Printf("1=%s\n", tracks.Track1)
		}
		if trackFlags[1] {
			fmt.Printf("2=%s\n", tracks.Track2)
		}
		if trackFlags[2] {
			fmt.Printf("3=%s\n", tracks.Track3)
		}

	case write && raw:
		d1, d2, d3 := "", "", ""
		if trackFlags[0] {
			d1 = magstripe.PackRaw(trackData[0], magstripe.Track1Map, 6, bpc1)
		}
		if trackFlags[1] {
			d2 = magstripe.PackRaw(trackData[1], magstripe.Track23Map, 4, bpc2)
		}
		if trackFlags[2] {
			d3 = magstripe.PackRaw(trackData[2], magstripe.Track23Map, 4, bpc3)
		}
		return dev.WriteRawTracks(d1, d2, d3)

	case write: // ISO mode
		return dev.WriteTracks(trackData[0], trackData[1], trackData[2])

	case erase:
		return dev.EraseTracks(trackFlags[0], trackFlags[1], trackFlags[2])

	case locoOp:
		return dev.SetCoercivity(magstripe.LoCo)

	case hicoOp:
		return dev.SetCoercivity(magstripe.HiCo)

	case bpiOp:
		return dev.SetBPI(bpi1, bpi2, bpi3)
	}

	return nil
}
