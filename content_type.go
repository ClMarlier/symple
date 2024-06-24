package symple

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
)

type contentTypeConfig struct {
	types []string
}

type contentTypeOption func(*contentTypeConfig) error

// WithContentType restricts the access to the current router and all it's
// children to the specified set of Content-Type
func WithContentType(opts ...contentTypeOption) routerOption {
	return func(rb *routerBuilder) error {
		config := &contentTypeConfig{
			types: []string{},
		}

		for _, option := range opts {
			if err := option(config); err != nil {
				return err
			}
		}
		rb.middlewareStack = append(rb.middlewareStack, config.contentType)
		return nil
	}
}

func WithApplicationJSON(ctc *contentTypeConfig) error {
	if slices.Contains(ctc.types, "application/json") {
		return fmt.Errorf("duplicate Content-Type")
	}
	ctc.types = append(ctc.types, "application/json")
	return nil
}

func WithApplicationXML(ctc *contentTypeConfig) error {
	if slices.Contains(ctc.types, "application/xml") {
		return fmt.Errorf("duplicate Content-Type")
	}
	ctc.types = append(ctc.types, "application/xml")
	return nil
}

func WithFormEncoded(ctc *contentTypeConfig) error {
	if slices.Contains(ctc.types, "application/x-www-form-urlencoded") {
		return fmt.Errorf("duplicate Content-Type")
	}
	ctc.types = append(ctc.types, "application/x-www-form-urlencoded")
	return nil
}

func WithFormData(ctc *contentTypeConfig) error {
	if slices.Contains(ctc.types, "multipart/form-data") {
		return fmt.Errorf("duplicate Content-Type")
	}
	ctc.types = append(ctc.types, "multipart/form-data")
	return nil
}

func getContentType(r *http.Request) string {
	contentType := r.Header.Get("Content-Type")
	if i := strings.Index(contentType, ";"); i > -1 {
		contentType = contentType[0:i]
	}
	return contentType
}

func (ctc *contentTypeConfig) contentType(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength != 0 {
			contentType := getContentType(r)
			if !slices.Contains(ctc.types, contentType) {
				ErrorResponse(
					w,
					fmt.Errorf(
						"invalid Content-Type, found %s, wanted %s",
						contentType,
						strings.Join(ctc.types, ", ")),
					http.StatusUnsupportedMediaType)
				return
			}
		}
		next(w, r)
	}
}
