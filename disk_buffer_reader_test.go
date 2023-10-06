package diskbufferreader

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

func TestReadAll(t *testing.T) {
	tests := map[string]struct {
		content string
		record  bool
		reset   bool
	}{
		"RecordOnNoReset": {
			"OneTwoThree",
			true,
			false,
		},
		"RecordOffNoReset": {
			"OneTwoThree",
			false,
			false,
		},
		"RecordOnReset": {
			"OneTwoThree",
			true,
			true,
		},
		"RecordOffReset": {
			"OneTwoThree",
			false,
			true,
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

		if testCase.reset {
			chunk := make([]byte, 3)
			dbr.Read(chunk)
			dbr.Reset()
		}

		if !testCase.record {
			dbr.Stop()
		}

		baseBytes, baseErr := io.ReadAll(bytesReader)
		testBytes, testErr := io.ReadAll(dbr)

		if string(testBytes) != string(baseBytes) {
			t.Fatalf("%s: Unexpected read result. Got: %v, expected: %v", testName, testBytes, baseBytes)
		}

		if !errors.Is(testErr, baseErr) {
			t.Fatalf("%s: Unexpected error. Got: %s, expected: %s", testName, testErr, baseErr)
		}
	}
}

func TestReadAllLarge(t *testing.T) {
	resp, err := http.Get("https://raw.githubusercontent.com/bill-rich/bad-secrets/master/FifteenMB.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	readBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	bytesReader := bytes.NewBuffer(readBytes)
	tmpReader := bytes.NewBuffer(readBytes)
	dbr, err := New(tmpReader)
	if err != nil {
		t.Fatal(err)
	}
	defer dbr.Close()

	chunk := make([]byte, 3)
	dbr.Read(chunk)
	dbr.Reset()

	dbr.Stop()

	baseBytes, baseErr := io.ReadAll(bytesReader)
	testBytes, testErr := io.ReadAll(dbr)

	if len(testBytes) != len(baseBytes) {
		t.Fatalf("Wrong number of bytes read. Got: %d, expected: %d", len(testBytes), len(baseBytes))
	}

	if string(testBytes) != string(baseBytes) {
		t.Fatalf("Unexpected read result. Got: %v..%v, expected: %v..%v", testBytes[:1024], testBytes[len(testBytes)-16:], baseBytes[:1024], baseBytes[len(baseBytes)-16:])
	}

	if !errors.Is(testErr, baseErr) {
		t.Fatalf("Unexpected error. Got: %s, expected: %s", testErr, baseErr)
	}
}

func TestTmpDir(t *testing.T) {
	tmpDir := "/tmp/dbrtest"

	err := os.Mkdir(tmpDir, 0755)
	if !os.IsExist(err) {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("TMPDIR", tmpDir)

	reader := strings.NewReader("TestString")
	dbr, err := New(reader)
	if err != nil {
		t.Fatal(err)
	}

	defer dbr.Close()

	testBytes := make([]byte, 1)
	_, err = dbr.Read(testBytes)
	if err != nil {
		t.Fatal(err)
	}

	tmpFileName := filepath.Base(dbr.tmpFile.Name())
	_, err = os.Stat(fmt.Sprintf("%s/%s", tmpDir, tmpFileName))
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkSeek(b *testing.B) {
	// Use a fixed data source for consistent benchmarking.
	data := make([]byte, 100000) // Example: 100KB of data
	for i := range data {
		data[i] = byte(i % 256)
	}

	tmpReader := bytes.NewBuffer(data)

	dbr, err := New(tmpReader)
	if err != nil {
		b.Fatal(err)
	}
	defer dbr.Close()

	whenceOptions := []int{io.SeekStart, io.SeekCurrent, io.SeekEnd}
	offset := int64(100) // Example offset value

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		whence := whenceOptions[i%len(whenceOptions)] // Vary whence to cover different cases
		_, err := dbr.Seek(offset, whence)
		if err != nil {
			b.Fatal(err)
		}
	}
}
