package downloader

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownload(t *testing.T) {
	content := &mockContent{size: 1000}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "some file.pdf", time.Now(), content)
	}))
	defer ts.Close()

	downloader := Downloader{}
	ref, ok := downloader.Add(ts.URL+"/some file.pdf", &mockWriteCloser{}, "dummy dest")
	require.True(t, ok)
	assert.NotNil(t, ref)

	var status map[string]string
	for status["Finished"] != "true" {
		status = downloader.Status(ref)
		time.Sleep(1 * time.Millisecond)
	}

	assert.Equal(t, "1.00", status["Progress"])
	assert.Equal(t, content.size, downloader.Downloads[ref].Size)
}

type mockContent struct {
	size   int64
	offset int64
}

func (mc *mockContent) Read(p []byte) (n int, err error) {
	if mc.offset+int64(len(p)) > mc.size {
		mc.offset = mc.size
		return int(mc.size - mc.offset), nil
	}
	mc.offset += int64(len(p))
	return len(p), nil
}

func (mc *mockContent) Seek(offset int64, whence int) (int64, error) {
	var calc int64
	switch whence {
	case io.SeekStart:
		calc = offset
	case io.SeekCurrent:
		calc = mc.offset + offset
	case io.SeekEnd:
		calc = mc.size + offset
	default:
		calc = -1
	}

	if calc < 0 || calc >= mc.size {
		return 0, errors.New("offset outside of file")
	}

	mc.offset = calc
	fmt.Printf("call %v, %v returning %v, nil", offset, whence, calc)
	return mc.offset, nil
}

type mockWriteCloser struct{}

func (*mockWriteCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (*mockWriteCloser) Close() error {
	return nil
}
