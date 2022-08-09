# Disk Buffer Reader
Are you tired of dealing with how to use a reader more than once? Me too! Rather than teeing a reader, reading it all, or other messy methods, try using your disk!

## What it does
Disk buffer reader uses takes everything read from a reader, writes it to a temporary file on disk, and resets the reader to read from the start again. Once you're done reusing your reader, stop the recording function and use it as normal.

## Example
```
func main() {
  readerOriginal := bytes.NewBuffer([]bytes("OneTwoThr")
  dbr, _ := diskBufferReader.NewReader(readerOriginal)
  defer dbr.Close() // Close reader and delete tmp file.
  message := make([]byte, 3)
  for i := 0; i < 3; i++ {
    dbr.Read(message)
    fmt.Println(message)
  }
  for i := 0; i < 3; i++ {
    dbr.Reset() // Reset reader to 0.
    dbr.Read(message)
    fmt.Println(message)
  }
  dbr.Stop() // Stop storing read bytes in tmp file.
}
```
```
One
Two
Thr
One
One
One
```
