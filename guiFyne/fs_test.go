package main

import (
	"io/fs"
	"net/http"
	"path/filepath"
	"testing"

	"gitee.com/bon-ami/eztools/v5"
	"gitlab.com/bon-ami/ezcomm"
)

var tstCntFile int

// TestFS uses TstRoot and TstMsgCount
func TestFS(t *testing.T) {
	ezcomm.Init4Tests(t)
	hfs := httpFS(*ezcomm.TstRoot)
	tstCntFile = *ezcomm.TstMsgCount
	if tstCntFile == 0 {
		tstCntFile--
	}
	if err := tstFSRead(t, hfs, ""); err != nil {
		t.Fatal(err)
	}
}

func tstFSRead(t *testing.T, hfs httpFS, chld string) (err error) {
	var ff http.File
	sz := -1
	for i := -1; i < 3 && err == nil; i++ {
		var sz1 int
		ff, err = hfs.Open(chld)
		//t.Log("open", chld, err)
		if err == nil {
			sz1, err = tstFFRead(t, hfs, ff, chld, sz, i)
			if sz < 0 {
				sz = sz1
			}
		}
		if err != nil {
			t.Error(chld, "under", hfs)
		}
		ff.Close()
		if sz < 1 || (sz == 1 && i == 1) {
			if eztools.Verbose > 0 {
				t.Log(chld, "size", sz, "read", sz1)
			}
			break
		}
	}
	return err
}

func tstFFRead(t *testing.T, hfs httpFS, ff http.File,
	fn string, sz, readParam int) (int, error) {
	fi, err := ff.Stat()
	if err != nil {
		return -1, err
	}
	cnt := int(fi.Size())
	isDir := fi.IsDir()
	if eztools.Verbose > 1 {
		t.Log(fi.Name(), "is dir =", isDir, "size =", cnt)
	}
	rdFile := func(fd []fs.FileInfo) error {
		for _, f1 := range fd {
			if tstCntFile == 0 {
				break
			}
			if err = tstFSRead(t, httpFS(filepath.Join(string(hfs), fn)),
				f1.Name()); err != nil {
				return err
			}
		}
		return nil
	}
	switch isDir {
	case true:
		if readParam > 1 {
			readParam = sz
		}
		if eztools.Verbose > 1 {
			t.Log("readdir", readParam)
		}
		fd, err := ff.Readdir(readParam)
		if err != nil {
			break
		}
		fdLen := len(fd)
		switch readParam {
		case -1:
			err = rdFile(fd)
		case 1:
			if fdLen != 1 {
				return fdLen, eztools.ErrOutOfBound
			}
			for ; sz > 1; sz-- {
				fd, err = ff.Readdir(1)
				if err != nil {
					return sz, err
				}
				sz1 := len(fd)
				if sz1 != 1 {
					return sz1, eztools.ErrOutOfBound
				}
			}
			fd, err = ff.Readdir(1)
			sz1 := len(fd)
			if err == nil || sz1 != 0 {
				return sz1, eztools.ErrInExistence
			}
			err = nil
		default: // 0 or > 1
			if fdLen != sz {
				err = eztools.ErrAccess
			}
		}
		return fdLen, err

	case false:
		if tstCntFile == 0 {
			return -1, nil
		}
		tstCntFile--
		if eztools.Verbose > 1 {
			t.Log("readfile", readParam, "/", sz)
		}
		switch readParam {
		case -1:
			buf := make([]byte, cnt)
			sz, err = ff.Read(buf)
			if err != nil {
				break
			}
			if eztools.Verbose > 2 {
				t.Log(buf[:sz])
			}
			if cnt != sz {
				err = eztools.ErrAccess
			}
		case 1:
			buf := make([]byte, 1)
			for ; sz > 0; sz-- {
				cnt, err = ff.Read(buf)
				if err != nil {
					break
				}
				if eztools.Verbose > 2 {
					t.Log(buf[:cnt])
				}
				if cnt != 1 {
					err = eztools.ErrAccess
					break
				}
			}
			if err == nil {
				cnt, err = ff.Read(buf)
				if err == nil || cnt != 0 {
					return cnt, eztools.ErrInExistence
				}
				err = nil
			}
		}
	}
	return cnt, err
}
