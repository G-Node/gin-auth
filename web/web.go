package web

import (
	"fmt"
	"net/url"

	"github.com/G-Node/gin-auth/conf"
)

// MakeUrl makes a URL for other resources provided by gin-auth using
// the base url from the server config file.
func MakeUrl(pathFormat string, param ...interface{}) string {
	baseUrl := conf.GetServerConfig().BaseURL
	for i, s := range param {
		switch s := s.(type) {
		case string:
			param[i] = url.QueryEscape(s)
		}
	}
	pathFormat = fmt.Sprintf(pathFormat, param...)
	return baseUrl + pathFormat
}
