package magstripe

import (
	"testing"
)

func TestEncodeDecodeISODataBlock(t *testing.T) {
	strip1 := "TRACK1DATA"
	strip2 := "TRACK2DATA"
	strip3 := "TRACK3DATA"

	// Test encoding
	encoded := encodeISODataBlock(strip1, strip2, strip3)
	if len(encoded) == 0 {
		t.Fatal("Encoded data should not be empty")
	}

	// Test decoding
	decoded1, decoded2, decoded3, err := decodeISODataBlock(encoded)
	if err != nil {
		t.Fatalf("Failed to decode ISO data block: %v", err)
	}

	if decoded1 != strip1 {
		t.Errorf("Strip 1 mismatch: expected %q, got %q", strip1, decoded1)
	}
	if decoded2 != strip2 {
		t.Errorf("Strip 2 mismatch: expected %q, got %q", strip2, decoded2)
	}
	if decoded3 != strip3 {
		t.Errorf("Strip 3 mismatch: expected %q, got %q", strip3, decoded3)
	}
}

func TestDecodeISODataBlockErrors(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		expectError bool
	}{
		{
			name:        "Invalid header",
			data:        "INVALID",
			expectError: true,
		},
		{
			name:        "Missing end code",
			data:        "\x1bs\x1b\x01test\x1b\x02test\x1b\x03test",
			expectError: true,
		},
		{
			name:        "Too short",
			data:        "\x1b",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := decodeISODataBlock(tt.data)
			if tt.expectError && err == nil {
				t.Errorf("Expected error for %s, but got none", tt.name)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.name, err)
			}
		})
	}
}

func TestTrackMappings(t *testing.T) {
	// Test Track1Map contains expected characters
	if len(Track1Map) == 0 {
		t.Error("Track1Map should not be empty")
	}

	// Should contain space and alphanumeric characters
	if Track1Map[0] != ' ' {
		t.Error("Track1Map should start with space character")
	}

	// Test Track23Map contains numeric characters
	if len(Track23Map) == 0 {
		t.Error("Track23Map should not be empty")
	}
	if Track23Map[0] != '0' {
		t.Error("Track23Map should start with '0'")
	}
}

func TestConstants(t *testing.T) {
	if EscapeCode != "\x1B" {
		t.Errorf("EscapeCode should be \\x1B, got %q", EscapeCode)
	}
	if EndCode != "\x1C" {
		t.Errorf("EndCode should be \\x1C, got %q", EndCode)
	}
	if HiCo != true {
		t.Error("HiCo should be true")
	}
	if LoCo != false {
		t.Error("LoCo should be false")
	}
	if HiBPI != true {
		t.Error("HiBPI should be true")
	}
	if LoBPI != false {
		t.Error("LoBPI should be false")
	}
}

func TestTrackData(t *testing.T) {
	td := &TrackData{
		Track1: "test1",
		Track2: "test2",
		Track3: "test3",
	}

	if td.Track1 != "test1" {
		t.Error("TrackData.Track1 not set correctly")
	}
	if td.Track2 != "test2" {
		t.Error("TrackData.Track2 not set correctly")
	}
	if td.Track3 != "test3" {
		t.Error("TrackData.Track3 not set correctly")
	}
}

func TestRawData(t *testing.T) {
	rd := RawData{
		Data:         "test",
		TotalLength:  4,
		ParityErrors: "    ",
		LRCError:     false,
	}

	if rd.Data != "test" {
		t.Error("RawData.Data not set correctly")
	}
	if rd.TotalLength != 4 {
		t.Error("RawData.TotalLength not set correctly")
	}
	if rd.LRCError != false {
		t.Error("RawData.LRCError not set correctly")
	}
}

func TestPackUnpackRawSimplified(t *testing.T) {
	// Test simplified versions
	data := "TEST123"
	packed := PackRaw(data, Track1Map, 6, 8)
	unpacked := UnpackRaw(packed, Track1Map, 6, 8)

	// For simplified version, should return original data
	if unpacked.Data != data {
		t.Errorf("Simplified PackRaw/UnpackRaw: expected %q, got %q", data, unpacked.Data)
	}
}

// Test that MSR struct can be created (compilation test)
func TestMSRCreation(t *testing.T) {
	// This test just ensures the MSR struct and its methods compile correctly
	// We can't actually test device functionality without hardware

	// Test that NewMSR function signature works
	_, err := NewMSR("/dev/null")
	// We expect this to fail since /dev/null isn't a serial port
	if err == nil {
		t.Error("Expected error when opening /dev/null as serial port")
	}
}

// Benchmark tests
func BenchmarkEncodeISODataBlock(b *testing.B) {
	strip1 := "TRACK1BENCHMARKDATA"
	strip2 := "TRACK2BENCHMARKDATA"
	strip3 := "TRACK3BENCHMARKDATA"

	for i := 0; i < b.N; i++ {
		encodeISODataBlock(strip1, strip2, strip3)
	}
}

func BenchmarkDecodeISODataBlock(b *testing.B) {
	strip1 := "TRACK1BENCHMARKDATA"
	strip2 := "TRACK2BENCHMARKDATA"
	strip3 := "TRACK3BENCHMARKDATA"
	encoded := encodeISODataBlock(strip1, strip2, strip3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decodeISODataBlock(encoded)
	}
}
