package diskbufferreader

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// DiskBufferReader uses an io.Reader and stores read bytes to a tmp file so the reader
// can be reset to the start.
type DiskBufferReader struct {
	recording bool
	reader    io.Reader
	bytesRead int64
	tmpFile   *os.File
	index     int64
}

// New takes an io.Reader and creates returns an initialized DiskBufferReader.
func New(r io.Reader) (*DiskBufferReader, error) {
	tmpFile, err := ioutil.TempFile("", "disk-buffer-file")
	if err != nil {
		return nil, err
	}
	return &DiskBufferReader{
		recording: true,
		reader:    r,
		bytesRead: 0,
		tmpFile:   tmpFile,
		index:     0,
	}, nil
}

// Read from the len(out) bytes from the reader starting at the current index.
func (dbr *DiskBufferReader) Read(out []byte) (int, error) {
	outLen := len(out)

	if outLen == 0 {
		return 0, nil
	}

	if int64(outLen)+dbr.index > dbr.bytesRead && dbr.recording {
		// Go to end of file so writes go at the end.
		_, err := dbr.tmpFile.Seek(0, io.SeekEnd)
		if err != nil {
			return 0, err
		}

		// Will need the difference of the requested bytes and how many are read.
		bytesToRead := int(int64(outLen) + dbr.index - dbr.bytesRead)
		readerBytes := make([]byte, bytesToRead)

		// Read the bytes from the reader.
		n, err := dbr.reader.Read(readerBytes)
		if err != nil && !errors.Is(err, io.EOF) {
			return 0, err
		}

		// Write the read bytes to the tmp file.
		m, err := dbr.tmpFile.Write(readerBytes[:n])
		if err != nil {
			return 0, err
		}

		// Update the number of bytes read.
		dbr.bytesRead += int64(m)

		// Go back to the beginning of the tmp file so reads start from the beginning.
		dbr.tmpFile.Seek(dbr.index, io.SeekStart)
	}

	// Read from the multireader of the tmp file and the reader.
	mr := io.MultiReader(dbr.tmpFile, dbr.reader)
	bytesRead := 0
	outBuffer := bytes.NewBuffer([]byte{})
	outMulti := make([]byte, len(out))
	var outErr error
	if dbr.index <= dbr.bytesRead {
		dbr.tmpFile.Seek(dbr.index, io.SeekStart)
	}
	for {
		n, err := mr.Read(outMulti)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return 0, err
			}
			outErr = err
		}

		outBuffer.Write(outMulti[:n])
		bytesRead += n

		if errors.Is(outErr, io.EOF) || bytesRead >= outLen {
			if bytesRead > 0 && bytesRead < outLen && errors.Is(err, io.EOF) {
				fmt.Printf("%d", bytesRead)
				outErr = nil
			}

			break
		}
	}
	outStart := dbr.index
	if int64(bytesRead) < dbr.index {
		outStart = int64(bytesRead)
	}
	copy(out, outBuffer.Bytes()[outStart:])
	dbr.index = int64(bytesRead)
	return bytesRead, outErr
}

// Reset the reader position to the start.
func (dbr *DiskBufferReader) Reset() error {
	if !dbr.recording {
		return fmt.Errorf("can not reset disk buffer reader after disk buffering is stopped")
	}
	dbr.index = 0
	return nil
}

// Stop storing the read bytes in the tmp file.
func (dbr *DiskBufferReader) Stop() {
	dbr.recording = false
}

// Close the reader and delete the tmp file.
func (dbr *DiskBufferReader) Close() error {
	err := os.Remove(dbr.tmpFile.Name())
	if err != nil {
		return err
	}
	return dbr.tmpFile.Close()
}
