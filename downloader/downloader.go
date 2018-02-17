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

	buf := make([]byte, 32*1024)

	var err2 error
	var n int
	for err != nil && err2 != nil {
		_, err = resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			d.err = err
			return
		}
		n, err2 = out.Write(buf)
		d.progress += float64(n)
	}

	fmt.Printf("out of read loop for url %q with errors %v, %v\n", d.url, err, err2)

	if err != nil && err != io.EOF {
		d.err = err
		return
	}
	if err2 != nil && err2 != io.EOF {
		d.err = err2
		return
	}

	d.finished = true
}
