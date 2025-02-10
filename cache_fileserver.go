package symple

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type cacheFile struct {
	data []byte
	pos  *int64
	info os.FileInfo
}

func (cf cacheFile) clone() cacheFile {
	var pos int64
	return cacheFile{data: cf.data, pos: &pos, info: cf.info}
}

func (cf cacheFile) Read(p []byte) (n int, err error) {
	if *cf.pos >= int64(len(cf.data)) {
		return 0, io.EOF
	}

	n = copy(p, (cf.data)[*cf.pos:])
	*cf.pos += int64(n)
	return n, nil
}

func (cf cacheFile) Seek(offset int64, whence int) (int64, error) {
	var newPos int64

	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = *cf.pos + offset
	case io.SeekEnd:
		newPos = int64(len(cf.data)) + offset
	default:
		return 0, errors.New("invalid whence value")
	}

	if newPos < 0 || newPos > int64(len(cf.data)) {
		return 0, errors.New("seek out of range")
	}

	*cf.pos = newPos
	return *cf.pos, nil
}

func (cf cacheFile) Close() error {
	*cf.pos = 0
	return nil
}

func (bf cacheFile) Readdir(count int) ([]os.FileInfo, error) {
	if count > 0 {
		return nil, nil
	}
	return nil, nil
}

func (cf cacheFile) Stat() (os.FileInfo, error) {
	return cf.info, nil
}

type cacheFileSystem struct {
	store map[string]cacheFile
}

func (cfs cacheFileSystem) Open(name string) (http.File, error) {
	val, ok := cfs.store[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return val.clone(), nil
}

func CacheFileSystem(path string) (http.FileSystem, error) {
	cfs := cacheFileSystem{store: make(map[string]cacheFile)}

	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			fileBytes, err := io.ReadAll(file)
			if err != nil {
				return err
			}
			var pos int64
			cfs.store[fmt.Sprintf("/%s", path)] = cacheFile{data: fileBytes, pos: &pos, info: info}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return cfs, nil
}

// Be carefull that tralling slash is automaticaly added to the pattern so do not
// add one.
func (rs *routerState) WithCacheFileServer(path string, pattern string, cacheControl string, expire time.Duration) routerOption {
	return func(rb *routerBuilder) error {
		cacheFS, err := CacheFileSystem(path)
		if err != nil {
			return err
		}
		rb.routeStack = append(
			rb.routeStack,
			routeDefinition{
				id:              rs.getSequence(),
				pattern:         fmt.Sprintf("GET %s/", pattern),
				handler:         fileServerWithCacheControl(http.FileServer(cacheFS), cacheControl, expire),
				middlewareStack: []Middleware{},
			},
		)
		rs.setExtra(rs.getSequence(), routeExtra{options: setBool{}, sitemap: setBool{}})
		rs.nextSequence()
		return nil
	}
}

func fileServerWithCacheControl(fsFunc http.Handler, cacheControl string, expire time.Duration) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Cache-Control", cacheControl)
		w.Header().Set("Expires", time.Now().Add(expire).Format(http.TimeFormat))
		w.Header().Set("ETag", fmt.Sprintf(`"%d"`, time.Now().Unix()))

		fsFunc.ServeHTTP(w, r)
		return nil
	}
}
