package downloader

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "file", time.Now(), strings.NewReader(`fssfhsrhfffihsh38h3w98r3289tfhiw;kqd29eq0oajd`))
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
	assert.Equal(t, int64(45), downloader.Downloads[ref].Size)
}

type mockWriteCloser struct{}

func (*mockWriteCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (*mockWriteCloser) Close() error {
	return nil
}
