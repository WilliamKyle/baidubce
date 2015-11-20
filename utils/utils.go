package utils

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

func GetHttpHeadTimeStamp() string {
	gmt := time.Now().Add(-1000 * 1000 * 1000 * 60 * 60 * 8)
	return gmt.Format("2006-01-02T15:04:05Z")
}

func IsStringInSlice(s string, slice []string) bool {
	for _, v := range slice {
		if s == v {
			return true
		}
	}
	return false
}

func UriEncode(s string) string {
	result := []string{}
	ss := strings.Split(s, "/")
	for _, v := range ss {
		value, _ := url.Parse(v)
		result = append(result, fmt.Sprint(value))
	}
	return strings.Join(result, "%2F")
}

func UriEncodeExceptSlash(s string) string {
	result := []string{}
	ss := strings.Split(s, "/")
	for _, v := range ss {
		value, _ := url.Parse(v)
		result = append(result, fmt.Sprint(value))
	}
	return strings.Join(result, "/")
}

