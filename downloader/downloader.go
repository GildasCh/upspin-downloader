package downloader

import (
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type Downloader struct {
	downloads map[string]*download
	sync.Mutex
}

func (d *Downloader) Add(url string, out io.WriteCloser) (string, bool) {
	if d.downloads == nil {
		d.downloads = make(map[string]*download)
	}

	d.Lock()
	defer d.Unlock()

	ref := fmt.Sprintf("%x", sha1.Sum([]byte(url)))

	if _, ok := d.downloads[ref]; ok {
		return ref, false
	}

	d.downloads[ref] = &download{url: url}
	go d.downloads[ref].start(out)

	return ref, true
}

func (d *Downloader) Status(ref string) map[string]string {
	if dl, ok := d.downloads[ref]; ok {
		return map[string]string{
			"URL":      dl.url,
			"Progress": fmt.Sprintf("%.2f", dl.progress),
			"Finished": fmt.Sprintf("%v", dl.finished),
			"Error":    fmt.Sprintf("%v", dl.err),
		}
	}

	fmt.Printf("ref %q not found in %#v\n", ref, d.downloads)
	return nil
}

type download struct {
	url      string
	progress float64
	finished bool
	err      error
}

func (d *download) start(out io.WriteCloser) {
	defer out.Close()

	resp, err := http.Get(d.url)
	if err != nil {
		d.err = err
		return
	}
	defer resp.Body.Close()

	fmt.Printf("response from get %q: %#v\n", d.url, resp)

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		d.err = err
	}

	d.progress = float64(n)
	d.finished = true
}
