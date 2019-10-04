package matroska

import (
	"fmt"
	"io"
	"reflect"
	"sync"
	"unsafe"
)

// #cgo CFLAGS: -std=gnu99
// #include "io.h"
import "C"

// Here be dragons.
//
// You cannot stash Go pointers in C-land, since Go's garbage collection can
// move things around between C calls, and this makes it rather hard to keep
// local state around when registering callbacks.
//
// A hack around this is to have a package-global map (wait, stop running away!)
// that is indexed with a inique key (unique to a given Demuxer object) that stores
// the io.ReadSeeker so that it can be grabbed when the Go callbacks are called
// from with in C (MatroskaParser's I/O), by passing the key as the opaque pointer
// instead of the structs. This, of couse, necessitates a bit of locking, but only
// during adding and deleting entries (i.e. when NewDemuxer() or Close() are called).
// Reads should not cause contention, since we use a sync.RWMutex here.
//
// For more rationale, see (the rather old... pre CGO docs):
//     https://gist.github.com/dwbuiten/c9865c4afb38f482702e
var ioTable = make(map[string]io.ReadSeeker)
var ioTableMu sync.RWMutex

func addReader(key string, r io.ReadSeeker) {
	ioTableMu.Lock()
	ioTable[key] = r
	ioTableMu.Unlock()
}

func delReader(key string) {
	ioTableMu.Lock()
	delete(ioTable, key)
	ioTableMu.Unlock()
}

func getReader(key string) (io.ReadSeeker, error) {
	ioTableMu.RLock()
	ret, ok := ioTable[key]
	ioTableMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("invalid reader key")
	}

	return ret, nil
}

//export cReadCallback
func cReadCallback(ckey *C.char, buf unsafe.Pointer, size C.int) C.int {
	key := C.GoString(ckey)

	r, err := getReader(key)
	if err != nil {
		return C.int(-1)
	}

	var gb []byte
	gbuf := (*reflect.SliceHeader)(unsafe.Pointer(&gb))
	gbuf.Data = uintptr(buf)
	gbuf.Len = int(size)
	gbuf.Cap = int(size)

	n, err := r.Read(gb)
	if err != nil && err != io.EOF {
		return C.int(-1)
	}

	return C.int(n)
}

//export cSeekCallback
func cSeekCallback(ckey *C.char, cpos C.ulonglong) C.int {
	key := C.GoString(ckey)

	r, err := getReader(key)
	if err != nil {
		return C.int(-1)
	}

	pos := int64(cpos)

	_, err = r.Seek(pos, io.SeekStart)
	if err != nil {
		return C.int(-1)
	}

	return C.int(0)
}

//export cSizeCallback
func cSizeCallback(ckey *C.char) C.longlong {
	key := C.GoString(ckey)

	r, err := getReader(key)
	if err != nil {
		return C.longlong(-1)
	}

	pos, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return C.longlong(-1)
	}

	endPos, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return C.longlong(-1)
	}

	_, err = r.Seek(pos, io.SeekStart)
	if err != nil {
		return C.longlong(-1)
	}

	return C.longlong(endPos)
}
