package ezcomm

import (
	"io"
	"os"
	"testing"

	"gitee.com/bon-ami/eztools/v5"
)

func TestFlow(t *testing.T) {
	FlowReaderNew = func(p string) (io.ReadCloser, error) {
		return os.OpenFile(p, os.O_RDONLY,
			eztools.FileCreatePermission)
	}
	FlowWriterNew = func(p string) (io.WriteCloser, error) {
		return os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
			eztools.FileCreatePermission)
	}
	Init4Tests(t)
	fl, err := ReadFlowFile("sample_desktop.xml")
	if err != nil {
		t.Fatal(err)
	}
	RunFlow(fl)
}
