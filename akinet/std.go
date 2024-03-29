package akinet

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/akitasoftware/akita-libs/buffer_pool"
	"github.com/akitasoftware/go-utils/sets"
	"github.com/akitasoftware/go-utils/slices"
	"github.com/google/uuid"
)

func FromStdRequest(streamID uuid.UUID, seq int, src *http.Request, body buffer_pool.Buffer) HTTPRequest {
	return HTTPRequest{
		StreamID:   streamID,
		Seq:        seq,
		Method:     src.Method,
		ProtoMajor: src.ProtoMajor,
		ProtoMinor: src.ProtoMinor,
		URL:        src.URL,
		Host:       src.Host,
		Cookies:    src.Cookies(),
		Header:     src.Header,
		Body:       body.Bytes(),

		buffer: body,
	}
}

func (r HTTPRequest) ToStdRequest() *http.Request {
	result := &http.Request{
		Method:        r.Method,
		URL:           r.URL,
		Proto:         fmt.Sprintf("HTTP/%d.%d", r.ProtoMajor, r.ProtoMinor),
		ProtoMajor:    r.ProtoMajor,
		ProtoMinor:    r.ProtoMinor,
		Host:          r.Host,
		Header:        r.Header,
		ContentLength: int64(r.Body.Len()),
		Body:          ioutil.NopCloser(r.Body.CreateReader()),
	}

	// Add any cookies in r.Cookies not already in r.Header.
	existingCookies := sets.NewSet[string](slices.Map(result.Cookies(), func(c *http.Cookie) string {
		return c.String()
	})...)
	for _, c := range r.Cookies {
		if v := c.String(); !existingCookies.Contains(v) {
			result.AddCookie(c)
			existingCookies.Insert(v)
		}
	}

	return result
}

func FromStdResponse(streamID uuid.UUID, seq int, src *http.Response, body buffer_pool.Buffer) HTTPResponse {
	return HTTPResponse{
		StreamID:   streamID,
		Seq:        seq,
		StatusCode: src.StatusCode,
		ProtoMajor: src.ProtoMajor,
		ProtoMinor: src.ProtoMinor,
		Cookies:    src.Cookies(),
		Header:     src.Header,
		Body:       body.Bytes(),

		buffer: body,
	}
}

func (r HTTPResponse) ToStdResponse() *http.Response {
	response := &http.Response{
		Status:        http.StatusText(r.StatusCode),
		StatusCode:    r.StatusCode,
		Proto:         fmt.Sprintf("HTTP/%d.%d", r.ProtoMajor, r.ProtoMinor),
		ProtoMajor:    r.ProtoMajor,
		ProtoMinor:    r.ProtoMinor,
		Header:        r.Header,
		ContentLength: int64(r.Body.Len()),
		Body:          ioutil.NopCloser(r.Body.CreateReader()),
	}

	// Add any cookies in r.Cookies not already in r.Header.
	existingCookies := sets.NewSet[string](slices.Map(response.Cookies(), func(c *http.Cookie) string {
		return c.String()
	})...)
	for _, c := range r.Cookies {
		if v := c.String(); v != "" && !existingCookies.Contains(v) {
			response.Header.Add("Set-Cookie", v)
			existingCookies.Insert(v)
		}
	}

	return response
}
