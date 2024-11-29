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

func requestContentType(cts []string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength != 0 {
				contentType := getContentType(r)
				if !slices.Contains(cts, contentType) {
					ErrorResponse(
						w,
						fmt.Errorf(
							"invalid Content-Type, found %s, wanted %s",
							contentType,
							strings.Join(cts, ", ")),
						http.StatusUnsupportedMediaType)
					return
				}
			}
			next(w, r)
		}
	}
}

func WithRequestContentType(cts []ContentType) routerOption {
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

func responseContentType(ct ContentType) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", string(ct))
			next(w, r)
		}
	}
}

func WithResponseContentType(ct ContentType) routerOption {
	return func(rb *routerBuilder) error {
		rb.middlewareStack = append(rb.middlewareStack, responseContentType(ct))
		return nil
	}
}
