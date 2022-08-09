package diskbufferreader

import (
	"bytes"
	"errors"
	"testing"
)

func TestDiskBufferReader(t *testing.T) {
	tests := map[string]struct {
		iterations int
		readSize   int
		content    string
	}{
		"SingleReadSizeSmall": {
			1,
			3,
			"OneTwoThree",
		},
		"SingleReadSizeEqual": {
			1,
			11,
			"OneTwoThree",
		},
		"SingleReadSizeLarge": {
			1,
			64,
			"OneTwoThree",
		},
		"DoubleReadSizeSmall": {
			2,
			3,
			"OneTwoThree",
		},
		"DoubleReadSizeEqual": {
			2,
			11,
			"OneTwoThree",
		},
		"DoubleReadSizeLarge": {
			2,
			64,
			"OneTwoThree",
		},
	}

	for testName, testCase := range tests {

		readBytes := []byte(testCase.content)
		bytesReader := bytes.NewBuffer(readBytes)
		tmpReader := bytes.NewBuffer(readBytes)
		dbr, err := New(tmpReader)
		if err != nil {
			t.Fatal(err)
		}
		defer dbr.Close()
		testBytes := make([]byte, testCase.readSize)
		baseBytes := make([]byte, testCase.readSize)
		testN, testErr := dbr.Read(testBytes)
		baseN, baseErr := bytesReader.Read(baseBytes)

		for i := 0; i < testCase.iterations; i++ {
			if string(testBytes) != string(baseBytes) {
				t.Fatalf("%s: Unexpected read result. Got: %v, expected: %v", testName, testBytes, baseBytes)
			}

			if testN != baseN {
				t.Fatalf("%s: Wrong number of bytes read. Got: %d, expected: %d", testName, testN, baseN)
			}

			if !errors.Is(testErr, baseErr) {
				t.Fatalf("%s: Unexpected error. Got: %s, expected: %s", testName, testErr, baseErr)
			}
		}
	}
}
