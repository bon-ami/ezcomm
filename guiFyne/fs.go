package main

import (
	"io"
	"io/fs"
	"math"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

const fileBufSize = 1024 * 1024 * 10

type fileInfo struct {
	uri fyne.URI
}

func (ff fileInfo) Name() string {
	return ff.uri.Name()
}

// Size reads the whole file and figures the size
func (ff fileInfo) Size() int64 {
	// TODO: by fyne
	if ok, err := storage.CanList(ff.uri); err == nil && ok {
		l, err := storage.List(ff.uri)
		if err != nil {
			return -1
		}
		return int64(len(l))
	}
	if ok, err := storage.CanRead(ff.uri); !ok || err != nil {
		return 0
	}
	rdr, err := storage.Reader(ff.uri)
	if err != nil {
		return 0
	}
	defer rdr.Close()
	buf := make([]byte, fileBufSize)
	var ret int64
	for {
		sz, err := rdr.Read(buf)
		if err != nil || sz < 1 {
			break
		}
		ret += int64(sz)
	}
	return ret
}

// Mode returns either ModeDir or ModeDevice
func (ff fileInfo) Mode() fs.FileMode {
	if ok, err := storage.CanList(ff.uri); ok && err == nil {
		return fs.ModeDir
	}
	return fs.ModeDevice
}

// ModTime returns nil
func (ff fileInfo) ModTime() time.Time { return time.Now() }

// IsDir checks whether listable
func (ff fileInfo) IsDir() bool {
	if ok, err := storage.CanList(ff.uri); ok && err == nil {
		return true
	}
	return false
}

// Sys returns nil
func (ff fileInfo) Sys() any { return nil }

type httpEntity struct {
	Files []fyne.URI
	FileN int
	Rdr   io.ReadCloser
}

var httpPool struct {
	Lock sync.Mutex
	Cnt  int
	Ent  map[int]httpEntity
}

type httpFile struct {
	id  int
	uri fyne.URI
}

func (fh httpFile) Readdir(n int) ([]fs.FileInfo, error) {
	ent, ok := httpPool.Ent[fh.id]
	if !ok {
		return nil, fs.ErrClosed
	}
	//eztools.Log(ent)
	var rdN int
	if n <= 0 || ent.Files == nil {
		if ok, err := storage.CanList(fh.uri); !ok || err != nil {
			if err == nil {
				err = fs.ErrPermission
			}
			return nil, err
		}
		var err error
		ent.Files, err = storage.List(fh.uri)
		if err != nil {
			return nil, err
		}
		rdN = len(ent.Files)
		if n > 0 {
			if n < len(ent.Files) {
				rdN = n
			}
		} else {
			ent.FileN = 0
		}
	} else {
		rdN = len(ent.Files) - ent.FileN
		//eztools.Log(rdN)
		if n < rdN {
			rdN = n
		}
		//eztools.Log(rdN)
		if rdN < 1 {
			return nil, io.EOF
		}
	}
	ret := make([]fs.FileInfo, rdN)
	for i := 0; i < rdN; i++ {
		ret[i] = fileInfo{ent.Files[ent.FileN+i]}
	}
	ent.FileN += rdN
	httpPool.Ent[fh.id] = ent
	return ret, nil
}

func (fh httpFile) Read(b []byte) (int, error) {
	ent := httpPool.Ent[fh.id]
	if ent.Rdr == nil {
		if ok, err := storage.CanRead(fh.uri); !ok || err != nil {
			return 0, fs.ErrPermission
		}
		var err error
		ent.Rdr, err = storage.Reader(fh.uri)
		if err != nil {
			return 0, err
		}
	}
	c, e := ent.Rdr.Read(b)
	httpPool.Ent[fh.id] = ent
	return c, e
}

func (fh httpFile) Close() error {
	ent, ok := httpPool.Ent[fh.id]
	httpPool.Lock.Lock()
	delete(httpPool.Ent, fh.id)
	switch {
	case len(httpPool.Ent) == 0:
		httpPool.Cnt = 0
	case httpPool.Cnt == fh.id:
		httpPool.Cnt--
	}
	httpPool.Lock.Unlock()
	if !ok || ent.Rdr == nil {
		return nil
	}
	return ent.Rdr.Close()
}

func (fh httpFile) Seek(offset int64, whence int) (int64, error) {
	// TODO: by fyne
	switch whence {
	case io.SeekStart:
		if offset == 0 {
			return 0, nil
		}
	}
	return 0, fs.ErrInvalid
}

func (fh httpFile) Stat() (fs.FileInfo, error) {
	return fileInfo{fh.uri}, nil
}

type httpFS string

func (fh httpFS) Open(name string) (http.File, error) {
	if name == "" {
		name = "."
	}
	uri := storage.NewFileURI(filepath.Join(string(fh), name))
	if ok, err := storage.Exists(uri); !ok || err != nil {
		if err == nil {
			err = fs.ErrNotExist
		}
		return nil, err
	}
	httpPool.Lock.Lock()
	// TODO: no limit on cnt
	if httpPool.Cnt == math.MaxInt {
		httpPool.Cnt = 0
	} else {
		httpPool.Cnt++
	}
	if httpPool.Ent == nil {
		httpPool.Ent = make(map[int]httpEntity, 1)
	}
	httpPool.Ent[httpPool.Cnt] = httpEntity{}
	ret := httpFile{id: httpPool.Cnt, uri: uri}
	httpPool.Lock.Unlock()
	return ret, nil
}
