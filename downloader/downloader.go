package downloader

import (
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type Downloader struct {
	Downloads map[string]*Download
	sync.Mutex
}

func (d *Downloader) Add(url string, out io.WriteCloser) (string, bool) {
	if d.Downloads == nil {
		d.Downloads = make(map[string]*Download)
	}

	d.Lock()
	defer d.Unlock()

	ref := fmt.Sprintf("%x", sha1.Sum([]byte(url)))

	if _, ok := d.Downloads[ref]; ok {
		return ref, false
	}

	d.Downloads[ref] = &Download{URL: url}
	go d.Downloads[ref].start(out)

	return ref, true
}

func (d *Downloader) Status(ref string) map[string]string {
	if dl, ok := d.Downloads[ref]; ok {
		return map[string]string{
			"URL":      dl.URL,
			"Progress": fmt.Sprintf("%.2f", dl.Progress),
			"Finished": fmt.Sprintf("%v", dl.Finished),
			"Error":    fmt.Sprintf("%v", dl.Err),
		}
	}

	fmt.Printf("ref %q not found in %#v\n", ref, d.Downloads)
	return nil
}

type Download struct {
	URL      string
	Progress float64
	Finished bool
	Err      error
}

func (d *Download) start(out io.WriteCloser) {
	defer out.Close()

	resp, err := http.Get(d.URL)
	if err != nil {
		d.Err = err
		return
	}
	defer resp.Body.Close()

	fmt.Printf("response from get %q: %#v\n", d.URL, resp)

	buf := make([]byte, 32*1024)
	for {
		nr, er := resp.Body.Read(buf)
		if nr > 0 {
			nw, ew := out.Write(buf[0:nr])
			if nw > 0 {
				d.Progress += float64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	if err != nil {
		d.Err = err
	}

	d.Finished = true
}
