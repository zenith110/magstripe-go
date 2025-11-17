# Magstripe Go

A Go library and command-line tool for working with magnetic stripe card readers/writers, specifically the MSR605 and compatible devices.

This is a Go port of the original Python MSR driver by Damien Bobillot.

## Features

- Read magnetic stripe cards in ISO format
- Write data to magnetic stripe cards
- Erase magnetic stripe tracks
- Set coercivity (HiCo/LoCo)
- Configure bits per inch (BPI) and bits per character (BPC)
- Support for tracks 1, 2, and 3
- Cross-platform support (Windows, Linux, macOS)

## Installation

### As a Go Module

```bash
go get github.com/abrahan/magstripe-go
```

### Building from Source

```bash
git clone https://github.com/abrahan/magstripe-go.git
cd magstripe-go
go mod tidy
go test
```

### Building the Command-Line Tool

```bash
cd cmd/msr
go build -o msr .
```

## Usage as Library

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/abrahan/magstripe-go"
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
    
    // Write tracks
    err = device.WriteTracks("track1data", "track2data", "track3data")
    if err != nil {
        log.Fatal("Failed to write tracks:", err)
    }
    
    // Set coercivity
    err = device.SetCoercivity(magstripe.HiCo)
    if err != nil {
        log.Fatal("Failed to set coercivity:", err)
    }
}
```

## API Reference

### Types

#### MSR
The main struct representing the magnetic stripe reader/writer connection.

#### TrackData
```go
type TrackData struct {
    Track1 string
    Track2 string
    Track3 string
}
```

#### RawData
```go
type RawData struct {
    Data         string
    TotalLength  int
    ParityErrors string
    LRCError     bool
}
```

### Functions

#### NewMSR(devPath string) (*MSR, error)
Creates a new MSR connection to the specified device path.

#### (*MSR) ReadTracks() (*TrackData, error)
Reads all magnetic tracks in ISO format.

#### (*MSR) WriteTracks(t1, t2, t3 string) error
Writes data to magnetic tracks in ISO format.

#### (*MSR) EraseTracks(t1, t2, t3 bool) error
Erases the specified magnetic tracks.

#### (*MSR) SetCoercivity(hico bool) error
Sets coercivity mode (true for high coercivity, false for low coercivity).

#### (*MSR) SetBPC(bpc1, bpc2, bpc3 int) error
Sets bits per character for each track (5-8 bits).

#### (*MSR) SetBPI(bpi1, bpi2, bpi3 *bool) error
Sets bits per inch for tracks (nil to skip, true for high BPI, false for low BPI).

#### (*MSR) ReadRawTracks() (string, string, string, error)
Reads magnetic tracks in raw format (simplified implementation).

#### (*MSR) WriteRawTracks(t1, t2, t3 string) error
Writes magnetic tracks in raw format (simplified implementation).

### Constants

```go
const (
    HiCo  = true   // High coercivity
    LoCo  = false  // Low coercivity
    HiBPI = true   // High bits per inch
    LoBPI = false  // Low bits per inch
)
```

## Command-Line Tool

The package includes a command-line tool `msr` that provides access to all MSR functions.

### Usage

```bash
msr [options] [data...]
```

### Options

- `-r`: Read magnetic tracks
- `-w`: Write magnetic tracks  
- `-e`: Erase magnetic tracks
- `-C`: Select high coercivity mode
- `-c`: Select low coercivity mode
- `-b`: Set bit per inch for each track (h=high, l=low)
- `-d`: Path to serial communication device (required)
- `-0`: Use raw encoding/decoding (don't use ISO)
- `-t`: Select tracks (1, 2, 3, 12, 23, 13, 123) [default: 123]
- `-B`: Set bits per character for each track (5-8)

### Examples

Read all tracks:
```bash
msr -d /dev/ttyUSB0 -r
```

Read specific tracks:
```bash
msr -d COM1 -r -t 12
```

Write data to tracks:
```bash
msr -d /dev/ttyUSB0 -w -t 123 "track1data" "track2data" "track3data"
```

Erase tracks:
```bash
msr -d /dev/ttyUSB0 -e -t 123
```

Set high coercivity:
```bash
msr -d /dev/ttyUSB0 -C
```

Set bits per inch:
```bash
msr -d /dev/ttyUSB0 -b hhl
```

## Device Compatibility

This library is designed for the MSR605 magnetic stripe reader/writer and compatible devices. It communicates over a serial connection at 9600 baud.

### Supported Operating Systems
- Windows (COM ports)
- Linux (/dev/ttyUSB*, /dev/ttyACM*)
- macOS (/dev/cu.*, /dev/tty.*)

## Protocol Details

The library implements the MSR605 communication protocol:

- **Escape Code**: `\x1B`
- **End Code**: `\x1C`
- **Commands**:
  - `a`: Reset device
  - `r`: Read tracks (ISO format)
  - `m`: Read tracks (raw format)  
  - `w`: Write tracks (ISO format)
  - `n`: Write tracks (raw format)
  - `c`: Erase tracks
  - `x`: Set high coercivity
  - `y`: Set low coercivity
  - `b`: Set bits per inch
  - `o`: Set bits per character

## Track Formats

- **Track 1**: 79 characters max, alphanumeric
- **Track 2**: 40 characters max, numeric
- **Track 3**: 107 characters max, numeric

### Character Mappings

- **Track 1**: ` !"#$%&'()*+`,./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_`
- **Track 2/3**: `0123456789:;<=>?`

## Development

### Running Tests

```bash
go test -v
```

### Examples

The `examples/` directory contains sample programs demonstrating various use cases:

- `examples/basic/` - Basic reading example
- `examples/write/` - Writing and verification example  
- `examples/raw/` - Raw format reading example

### Dependencies

- `go.bug.st/serial v1.6.2` - Cross-platform serial port library

## Error Handling

The library provides comprehensive error handling for:
- Serial communication failures
- Device command errors
- Data format validation
- Timeout conditions

All functions return Go-standard errors that can be checked and handled appropriately.

## Limitations

- Raw format bit manipulation is simplified in this version
- Some advanced MSR605 features may not be fully implemented
- Hardware access requires appropriate system permissions

## Contributing

Contributions are welcome! Please ensure:
1. All tests pass (`go test`)
2. Code follows Go conventions
3. New features include tests
4. Documentation is updated

## License

GNU GPL version 3

## Credits

Original Python implementation by Damien Bobillot (damien.bobillot.2002+msr@m4x.org)

Go port maintains compatibility with the original design and protocol while providing a more robust and type-safe interface.