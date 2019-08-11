package utils

import (
	"bytes"
	"encoding/base32"
	"net/http"
	"time"

	"github.com/pborman/uuid"
)

// StringInterface ...
type StringInterface map[string]interface{}

// StringMap ...
type StringMap map[string]string

// StringArray ...
type StringArray []string

var (
	encoding = base32.NewEncoding("ybndrfg8ejkmcpqxot1uwisza345h769")
)

// NewID ...
func NewID() string {
	var b bytes.Buffer
	encoder := base32.NewEncoder(encoding, &b)
	encoder.Write(uuid.NewRandom())
	encoder.Close()
	b.Truncate(26) // removes the '==' padding
	return b.String()
}

// GetMillis is a convenience method to get milliseconds since epoch.
func GetMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// GetProtocol ...
func GetProtocol(r *http.Request) string {
	if r.Header.Get("X-Forwarded-Proto") == "https" || r.TLS != nil {
		return "https"
	}
	return "http"
}
