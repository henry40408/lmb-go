package sync_reader

import (
	"bufio"
	"io"
	"sync"
)

type SyncReader struct {
	bufio.Reader
	m sync.Mutex
}

func NewSyncReader(reader io.Reader) SyncReader {
	r := bufio.NewReader(reader)
	return SyncReader{Reader: *r}
}

func (sr *SyncReader) ReadAll() ([]byte, error) {
	sr.m.Lock()
	defer sr.m.Unlock()
	return io.ReadAll(&sr.Reader)
}
func (sr *SyncReader) ReadLine() (string, error) {
	sr.m.Lock()
	defer sr.m.Unlock()
	return sr.Reader.ReadString('\n')
}

func (sr *SyncReader) Read(p []byte) (int, error) {
	sr.m.Lock()
	defer sr.m.Unlock()
	return sr.Reader.Read(p)
}

func (sr *SyncReader) ReadRune() (rune, int, error) {
	sr.m.Lock()
	defer sr.m.Unlock()
	return sr.Reader.ReadRune()
}
