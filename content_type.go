package symple

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
)

type ContentType string

const (
	ContentTypeTextPlain   ContentType = "text/plain"
	ContentTypeTextHtml    ContentType = "text/html"
	ContentTypeJson        ContentType = "application/json"
	ContentTypeXml         ContentType = "application/xml"
	ContentTypeFormEncoded ContentType = "application/x-www-form-urlencoded"
	ContentTypeFormData    ContentType = "multipart/form-data"
)

func getContentType(r *http.Request) string {
	contentType := r.Header.Get("Content-Type")
	if i := strings.Index(contentType, ";"); i > -1 {
		contentType = contentType[0:i]
	}
	return contentType
}

func requestContentType(cts []string) func(HandlerFunc) HandlerFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			if r.ContentLength != 0 {
				contentType := getContentType(r)
				if !slices.Contains(cts, contentType) {
					return fmt.Errorf("%w invalid Content-Type, found %s, wanted %s", ErrUnsupportedMediaType, contentType, strings.Join(cts, " or "))
				}
			}
			return next(w, r)
		}
	}
}

func (rs *routerState) WithRequestContentType(cts ...ContentType) routerOption {
	return func(rb *routerBuilder) error {
		stringContentType := make([]string, 0, len(cts))
		for _, ct := range cts {
			if !slices.Contains(stringContentType, string(ct)) {
				stringContentType = append(stringContentType, string(ct))
			} else {
				return fmt.Errorf("found duplicate request Content-Type config %s", string(ct))
			}
		}

		rb.middlewareStack = append(rb.middlewareStack, requestContentType(stringContentType))
		return nil
	}
}

func responseContentType(ct ContentType) func(HandlerFunc) HandlerFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			w.Header().Add("Content-Type", string(ct))
			return next(w, r)
		}
	}
}

func (rs *routerState) WithResponseContentType(ct ContentType) routerOption {
	return func(rb *routerBuilder) error {
		rb.middlewareStack = append(rb.middlewareStack, responseContentType(ct))
		return nil
	}
}
