package main

import (
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"gitee.com/bon-ami/eztools/v4"
)

const fyneFileBufSize = 1024 * 1024 * 10

type fyneFileInfo struct {
	uri fyne.URI
}

func (ff fyneFileInfo) Name() string {
	eztools.LogPrint("fileinf name", ff.uri)
	return ff.uri.Name()
}

// Size reads the whole file and figures the size
func (ff fyneFileInfo) Size() int64 {
	eztools.LogPrint("fileinf sizing", ff.uri)
	// TODO: by fyne
	if ok, err := storage.CanRead(ff.uri); !ok || err != nil {
		return 0
	}
	rdr, err := storage.Reader(ff.uri)
	if err != nil {
		return 0
	}
	defer rdr.Close()
	buf := make([]byte, fyneFileBufSize)
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
func (ff fyneFileInfo) Mode() fs.FileMode {
	if ok, err := storage.CanList(ff.uri); ok && err == nil {
		eztools.LogPrint("fileinf is dir", ff.uri)
		return fs.ModeDir
	}
	eztools.LogPrint("fileinf is device", ff.uri)
	return fs.ModeDevice
}

// ModTime returns nil
func (ff fyneFileInfo) ModTime() time.Time { return time.Now() }

// IsDir checks whether listable
func (ff fyneFileInfo) IsDir() bool {
	if ok, err := storage.CanList(ff.uri); ok && err == nil {
		eztools.LogPrint("fileinf isdir", ff.uri)
		return true
	}
	eztools.LogPrint("fileinf not isdir", ff.uri)
	return false
}

// Sys returns nil
func (ff fyneFileInfo) Sys() any { return nil }

type fyneHTTPFile struct {
	uri   fyne.URI
	files []fyne.URI
	fileN int
	rdr   io.ReadCloser
}

func (fh *fyneHTTPFile) Open() error {
	eztools.LogPrint("file opening", fh.uri)
	if fh.rdr != nil {
		return nil
	}
	if ok, err := storage.CanRead(fh.uri); !ok || err != nil {
		return fs.ErrPermission
	}
	var err error
	fh.rdr, err = storage.Reader(fh.uri)
	return err
}

func (fh fyneHTTPFile) Readdir(n int) ([]fs.FileInfo, error) {
	eztools.LogPrint("listing", fh.uri, fh.files)
	var rdN int
	if n <= 0 || fh.files == nil {
		if ok, err := storage.CanList(fh.uri); !ok || err != nil {
			if err == nil {
				err = fs.ErrPermission
			}
			return nil, err
		}
		var err error
		fh.files, err = storage.List(fh.uri)
		if err != nil {
			return nil, err
		}
		rdN = len(fh.files)
		if n > 0 {
			if n < len(fh.files) {
				rdN = n
			}
		} else {
			fh.fileN = 0
		}
	} else {
		rdN = len(fh.files) - fh.fileN
		if n < rdN {
			rdN = n
		}
	}
	ret := make([]fs.FileInfo, rdN)
	for i := 0; i < rdN; i++ {
		ret[i] = fyneFileInfo{fh.files[fh.fileN+i]}
		eztools.LogPrint("listed", ret[i].Name())
	}
	fh.fileN += rdN
	return ret, nil
}

func (fh fyneHTTPFile) Read(b []byte) (int, error) {
	if err := fh.Open(); err != nil {
		return 0, err
	}
	eztools.LogPrint("file reading", fh.uri)
	defer fh.rdr.Close()
	return fh.rdr.Read(b)
}
func (fh fyneHTTPFile) Close() error {
	eztools.LogPrint("file closing", fh.uri)
	if fh.rdr == nil {
		return nil
	}
	return fh.rdr.Close()
}

func (fh fyneHTTPFile) Seek(offset int64, whence int) (int64, error) {
	eztools.LogPrint("file seeking", fh.uri, offset, whence)
	// TODO: by fyne
	switch whence {
	case io.SeekStart:
		if offset == 0 {
			return 0, nil
		}
	}
	return 0, fs.ErrInvalid
}

func (fh fyneHTTPFile) Stat() (fs.FileInfo, error) {
	eztools.LogPrint("file stating", fh.uri)
	return fyneFileInfo{fh.uri}, nil
}

type fyneHTTPFS string

func (fh fyneHTTPFS) Open(name string) (http.File, error) {
	eztools.LogPrint("fs opening", fh, name)
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
	ret := fyneHTTPFile{uri: uri}
	/*if err := fh.fhf.Open(); err != nil {
		fh.fhf.uri = nil
		return nil, err
	}*/
	return ret, nil
}
