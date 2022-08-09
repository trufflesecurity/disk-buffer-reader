package diskBufferReader

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
)

type DiskBufferReader struct {
	recording bool
	reader    io.Reader
	bytesRead int64
	tmpFile   *os.File
}

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
	}, nil
}

func (dbr *DiskBufferReader) Read(out []byte) (int, error) {
	if dbr.recording {
		defer dbr.tmpFile.Seek(0, io.SeekStart)
	}
	outLen := len(out)

	if outLen == 0 {
		return 0, nil
	}

	if int64(outLen) > dbr.bytesRead && dbr.recording {
		// Go to end of file so writes go at the end.
		_, err := dbr.tmpFile.Seek(0, io.SeekEnd)
		if err != nil {
			return 0, err
		}

		// Will need the difference of the requested bytes and how many are read.
		bytesToRead := int(int64(outLen) - dbr.bytesRead)
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
		dbr.tmpFile.Seek(0, io.SeekStart)
	}

	// Read from the multireader of the tmp file and the reader.
	mr := io.MultiReader(dbr.tmpFile, dbr.reader)
	bytesRead := 0
	outBuffer := bytes.NewBuffer([]byte{})
	outMulti := make([]byte, len(out))
	var outErr error
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
			break
		}
	}
	copy(out, outBuffer.Bytes())
	return bytesRead, outErr
}

func (dbr *DiskBufferReader) Stop() {
	dbr.recording = false
}

func (dbr *DiskBufferReader) Close() error {
	err := os.Remove(dbr.tmpFile.Name())
	if err != nil {
		return err
	}
	return dbr.tmpFile.Close()
}
