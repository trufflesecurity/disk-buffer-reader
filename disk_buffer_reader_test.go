package diskbufferreader

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestReadExactReaderSize(t *testing.T) {
	readBytes := []byte("OneTwoThreeFourFive")
	reader := bytes.NewBuffer(readBytes)
	dbr, err := New(reader)
	if err != nil {
		t.Fatal(err)
	}
	defer dbr.Close()
	outBytes := make([]byte, len(readBytes))
	n, err := dbr.Read(outBytes)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(readBytes) {
		t.Fatalf("Wrong number of bytes read. Expected: %d, got: %d", len(readBytes), n)
	}
}

func TestReadMoreThanReaderSize(t *testing.T) {
	readBytes := []byte("OneTwoThreeFourFive")
	reader := bytes.NewBuffer(readBytes)
	dbr, err := New(reader)
	if err != nil {
		t.Fatal(err)
	}
	defer dbr.Close()
	outBytes := make([]byte, len(readBytes)+1)
	n, err := dbr.Read(outBytes)
	if !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	if n != len(readBytes) {
		t.Fatalf("Wrong number of bytes read. Expected: %d, got: %d", len(readBytes), n)
	}
}

func TestReadTwiceNoEOF(t *testing.T) {
	readBytes := []byte("OneTwoThreeFourFive")
	reader := bytes.NewBuffer(readBytes)
	dbr, err := New(reader)
	if err != nil {
		t.Fatal(err)
	}
	defer dbr.Close()
	outBytes := make([]byte, 3)
	_, err = dbr.Read(outBytes)
	if err != nil {
		t.Fatal(err)
	}
	if string(outBytes) != string(readBytes[:3]) {
		t.Fatalf("Wrong byte content. Expected: %s, got: %s", readBytes[:3], outBytes)
	}

	outBytes = make([]byte, 6)
	_, err = dbr.Read(outBytes)
	if err != nil {
		t.Fatal(err)
	}
	if string(outBytes) != string(readBytes[3:9]) {
		t.Fatalf("Wrong byte content. Expected: %s, got: %s", readBytes[3:9], outBytes)
	}
}

func TestReadTwiceReset(t *testing.T) {
	readBytes := []byte("OneTwoThreeFourFive")
	reader := bytes.NewBuffer(readBytes)
	dbr, err := New(reader)
	if err != nil {
		t.Fatal(err)
	}
	defer dbr.Close()
	outBytes := make([]byte, 3)
	_, err = dbr.Read(outBytes)
	if err != nil {
		t.Fatal(err)
	}
	if string(outBytes) != string(readBytes[:3]) {
		t.Fatalf("Wrong byte content. Expected: %s, got: %s", readBytes[:3], outBytes)
	}

	dbr.Reset()

	outBytes = make([]byte, 3)
	_, err = dbr.Read(outBytes)
	if err != nil {
		t.Fatal(err)
	}
	if string(outBytes) != string(readBytes[:3]) {
		t.Fatalf("Wrong byte content. Expected: %s, got: %s", readBytes[:3], outBytes)
	}
}

func TestReadTwiceEOF(t *testing.T) {
	readBytes := []byte("OneTwoThreeFourFive")
	reader := bytes.NewBuffer(readBytes)
	dbr, err := New(reader)
	if err != nil {
		t.Fatal(err)
	}
	defer dbr.Close()
	outBytes := make([]byte, 3)
	_, err = dbr.Read(outBytes)
	if err != nil {
		t.Fatal(err)
	}
	if string(outBytes) != string(readBytes[:3]) {
		t.Fatalf("Wrong byte content. Expected: %s, got: %s", readBytes[:3], outBytes)
	}

	outBytes = make([]byte, len(readBytes)+2)
	n, err := dbr.Read(outBytes)
	if !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}

	if n != len(readBytes)-3 {
		t.Fatalf("Wrong read length. Expected: %d, got: %d", len(readBytes)-3, n)
	}
}

func TestNoRecordingEOF(t *testing.T) {
	readBytes := []byte("OneTwoThreeFourFive")
	reader := bytes.NewBuffer(readBytes)
	dbr, err := New(reader)
	if err != nil {
		t.Fatal(err)
	}
	defer dbr.Close()
	outBytes := make([]byte, 3)
	_, err = dbr.Read(outBytes)
	if err != nil {
		t.Fatal(err)
	}
	if string(outBytes) != string(readBytes[:3]) {
		t.Fatalf("Wrong byte content. Expected: %s, got: %s", readBytes[:3], outBytes)
	}

	dbr.Stop()

	outBytes = make([]byte, 3)
	_, err = dbr.Read(outBytes)
	if err != nil {
		t.Fatal(err)
	}
	if string(outBytes) != string(readBytes[3:6]) {
		t.Fatalf("Wrong byte content. Expected: %s, got %s", outBytes, readBytes[3:6])
	}
}
