package magstripe

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"go.bug.st/serial"
)

// MSR represents a magnetic stripe card reader/writer
type MSR struct {
	port serial.Port
}

// Protocol constants
const (
	EscapeCode = "\x1B"
	EndCode    = "\x1C"
)

// Coercivity constants
const (
	HiCo = true
	LoCo = false
)

// BPI constants
const (
	HiBPI = true
	LoBPI = false
)

// Character mappings
var (
	Track1Map  = " !\"#$%&'()*+`,./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_"
	Track23Map = "0123456789:;<=>?"
)

// TrackData holds the data from magnetic stripe tracks
type TrackData struct {
	Track1 string
	Track2 string
	Track3 string
}

// RawData holds raw binary data from tracks
type RawData struct {
	Data         string
	TotalLength  int
	ParityErrors string
	LRCError     bool
}

// NewMSR creates a new MSR instance
func NewMSR(devPath string) (*MSR, error) {
	if !strings.Contains(devPath, "/") && !strings.Contains(devPath, "\\") {
		if strings.Contains(devPath, "COM") {
			// Windows
			devPath = devPath
		} else {
			// Unix-like
			devPath = "/dev/" + devPath
		}
	}

	mode := &serial.Mode{
		BaudRate: 9600,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(devPath, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port: %w", err)
	}

	msr := &MSR{port: port}
	msr.Reset()
	return msr, nil
}

// Close closes the serial connection
func (m *MSR) Close() error {
	return m.port.Close()
}

// executeNoResult sends a command without expecting a result
func (m *MSR) executeNoResult(command string) error {
	_, err := m.port.Write([]byte(EscapeCode + command))
	if err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}

// executeWaitResult sends a command and waits for a result
func (m *MSR) executeWaitResult(command string, timeout time.Duration) (status byte, result string, data string, err error) {
	// Clear input buffer by reading any available data
	clearBuffer := make([]byte, 1024)
	for {
		n, _ := m.port.Read(clearBuffer)
		if n == 0 {
			break
		}
	}

	// Send command
	_, err = m.port.Write([]byte(EscapeCode + command))
	if err != nil {
		return 0, "", "", err
	}
	time.Sleep(100 * time.Millisecond)

	// Read response with timeout
	startTime := time.Now()
	var response []byte
	buffer := make([]byte, 1024)

	// Set read timeout
	oldTimeout := time.Millisecond * 100
	m.port.SetReadTimeout(timeout)
	defer m.port.SetReadTimeout(oldTimeout)

	for time.Since(startTime) < timeout {
		n, err := m.port.Read(buffer)
		if err != nil && n == 0 {
			// Timeout or other error with no data
			break
		}
		if n > 0 {
			response = append(response, buffer[:n]...)
			// Check if we have a complete response
			if len(response) > 0 && strings.Contains(string(response), EscapeCode) {
				break
			}
		}
		if n == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	if len(response) == 0 {
		return 0, "", "", errors.New("operation timed out")
	}

	// Parse result: status, result, data
	responseStr := string(response)
	pos := strings.LastIndex(responseStr, EscapeCode)
	if pos == -1 {
		return 0, "", "", errors.New("invalid response format")
	}

	if pos+1 >= len(responseStr) {
		return 0, "", "", errors.New("incomplete response")
	}

	status = responseStr[pos+1]
	if pos+2 < len(responseStr) {
		result = responseStr[pos+2:]
	}
	if pos > 0 {
		data = responseStr[:pos]
	}

	return status, result, data, nil
}

// Reset resets the MSR device
func (m *MSR) Reset() error {
	return m.executeNoResult("a")
}

// decodeISODataBlock decodes ISO format data block
func decodeISODataBlock(data string) (string, string, string, error) {
	// Check header
	if len(data) < 4 || data[:4] != EscapeCode+"s"+EscapeCode+"\x01" {
		return "", "", "", fmt.Errorf("bad datablock: doesn't start with <ESC>s<ESC>[01]: %v", data)
	}

	// Check end
	if len(data) < 2 || data[len(data)-2:] != "?"+EndCode {
		return "", "", "", fmt.Errorf("bad datablock: doesn't end with ?<FS>: %v", data)
	}

	// Parse strips
	var strip1, strip2, strip3 string

	// First strip
	strip1Start := 4
	strip1End := strings.Index(data[strip1Start:], EscapeCode)
	if strip1End == -1 {
		return "", "", "", fmt.Errorf("bad datablock: missing escape code after strip 1")
	}
	strip1End += strip1Start

	if strip1End == strip1Start {
		strip1End += 2
	} else {
		strip1 = data[strip1Start:strip1End]
	}

	// Second strip
	strip2Start := strip1End + 2
	if strip2Start >= len(data) || data[strip1End:strip2Start] != EscapeCode+"\x02" {
		return "", "", "", fmt.Errorf("bad datablock: missing <ESC>[02] at position %d", strip1End)
	}

	strip2End := strings.Index(data[strip2Start:], EscapeCode)
	if strip2End == -1 {
		return "", "", "", fmt.Errorf("bad datablock: missing escape code after strip 2")
	}
	strip2End += strip2Start

	if strip2End == strip2Start {
		strip2End += 2
	} else {
		strip2 = data[strip2Start:strip2End]
	}

	// Third strip
	strip3Start := strip2End + 2
	if strip3Start >= len(data) || data[strip2End:strip3Start] != EscapeCode+"\x03" {
		return "", "", "", fmt.Errorf("bad datablock: missing <ESC>[03] at position %d", strip2End)
	}

	if strip3Start < len(data) && data[strip3Start] != EscapeCode[0] {
		strip3 = data[strip3Start : len(data)-2]
	}

	return strip1, strip2, strip3, nil
}

// encodeISODataBlock encodes data into ISO format
func encodeISODataBlock(strip1, strip2, strip3 string) string {
	return "\x1bs\x1b\x01" + strip1 + "\x1b\x02" + strip2 + "\x1b\x03" + strip3 + "?\x1C"
}

// ReadTracks reads magnetic tracks in ISO format
func (m *MSR) ReadTracks() (*TrackData, error) {
	status, _, data, err := m.executeWaitResult("r", 10*time.Second)
	if err != nil {
		return nil, err
	}
	if status != '0' {
		return nil, fmt.Errorf("read error: %c", status)
	}

	strip1, strip2, strip3, err := decodeISODataBlock(data)
	if err != nil {
		return nil, err
	}

	return &TrackData{
		Track1: strip1,
		Track2: strip2,
		Track3: strip3,
	}, nil
}

// WriteTracks writes magnetic tracks in ISO format
func (m *MSR) WriteTracks(t1, t2, t3 string) error {
	data := encodeISODataBlock(t1, t2, t3)
	status, _, _, err := m.executeWaitResult("w"+data, 10*time.Second)
	if err != nil {
		return err
	}
	if status != '0' {
		return fmt.Errorf("write error: %c", status)
	}
	return nil
}

// EraseTracks erases specified magnetic tracks
func (m *MSR) EraseTracks(t1, t2, t3 bool) error {
	mask := 0
	if t1 {
		mask |= 1
	}
	if t2 {
		mask |= 2
	}
	if t3 {
		mask |= 4
	}

	status, _, _, err := m.executeWaitResult("c"+string(byte(mask)), 10*time.Second)
	if err != nil {
		return err
	}
	if status != '0' {
		return fmt.Errorf("erase error: %c", status)
	}
	return nil
}

// SetCoercivity sets coercivity mode (high or low)
func (m *MSR) SetCoercivity(hico bool) error {
	var command string
	if hico {
		command = "x"
	} else {
		command = "y"
	}

	status, _, _, err := m.executeWaitResult(command, 10*time.Second)
	if err != nil {
		return err
	}
	if status != '0' {
		return fmt.Errorf("set_coercivity error: %c", status)
	}
	return nil
}

// SetBPC sets bits per character for each track
func (m *MSR) SetBPC(bpc1, bpc2, bpc3 int) error {
	status, _, _, err := m.executeWaitResult("o"+string(byte(bpc1))+string(byte(bpc2))+string(byte(bpc3)), 10*time.Second)
	if err != nil {
		return err
	}
	if status != '0' {
		return fmt.Errorf("set_bpc error: %c", status)
	}
	return nil
}

// SetBPI sets bits per inch for tracks
func (m *MSR) SetBPI(bpi1, bpi2, bpi3 *bool) error {
	var modes []string

	if bpi1 != nil {
		if *bpi1 {
			modes = append(modes, "\xA1") // 210bpi
		} else {
			modes = append(modes, "\xA0") // 75bpi
		}
	}

	if bpi2 != nil {
		if *bpi2 {
			modes = append(modes, "\xD2")
		} else {
			modes = append(modes, "\x4B")
		}
	}

	if bpi3 != nil {
		if *bpi3 {
			modes = append(modes, "\xC1")
		} else {
			modes = append(modes, "\xC0")
		}
	}

	for _, mode := range modes {
		status, _, _, err := m.executeWaitResult("b"+mode, 10*time.Second)
		if err != nil {
			return err
		}
		if status != '0' {
			return fmt.Errorf("set_bpi error: %c for %x", status, mode)
		}
	}
	return nil
}

// ReadRawTracks reads magnetic tracks in raw format (simplified version)
func (m *MSR) ReadRawTracks() (string, string, string, error) {
	status, _, data, err := m.executeWaitResult("m", 10*time.Second)
	if err != nil {
		return "", "", "", err
	}
	if status != '0' {
		return "", "", "", fmt.Errorf("read error: %c", status)
	}

	// Simplified raw data parsing - just return the raw binary data
	// For full raw parsing with bit manipulation, additional functions would be needed
	return data, "", "", nil
}

// WriteRawTracks writes magnetic tracks in raw format (simplified version)
func (m *MSR) WriteRawTracks(t1, t2, t3 string) error {
	// Simplified raw writing - would need full bit manipulation for complete implementation
	data := "\x1bs\x1b\x01" + string(byte(len(t1))) + t1 +
		"\x1b\x02" + string(byte(len(t2))) + t2 +
		"\x1b\x03" + string(byte(len(t3))) + t3 + "?\x1C"

	status, _, _, err := m.executeWaitResult("n"+data, 10*time.Second)
	if err != nil {
		return err
	}
	if status != '0' {
		return fmt.Errorf("write error: %c", status)
	}
	return nil
}

// PackRaw packs data into raw format (simplified placeholder)
func PackRaw(data, mapping string, bcountCode, bcountOutput int) string {
	// Simplified version - for full implementation, complex bit manipulation is needed
	return data
}

// UnpackRaw unpacks raw data (simplified placeholder)
func UnpackRaw(rawData, mapping string, bcountCode, bcountOutput int) RawData {
	// Simplified version - for full implementation, complex bit manipulation is needed
	return RawData{
		Data:         rawData,
		TotalLength:  len(rawData),
		ParityErrors: strings.Repeat(" ", len(rawData)),
		LRCError:     false,
	}
}
