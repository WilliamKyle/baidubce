package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"github.com/WilliamKyle/baidubce/httplib"
	"github.com/WilliamKyle/baidubce/utils"
	"sort"
	"strings"
)

var Debug bool

/*
 * 生成规范header
 */
func getCanonicalHeaders(headers map[string]string, headersToSign []string) (string, string) {
	// 没有指定 headersToSign 的情况下，默认使用：
	//   1.host
	//   2.content-md5
	//   3.content-length
	//   4.content-type
	//   5.所有以x-bce-开头的header项
	// 生成规范header
	if headersToSign == nil {
		headersToSign = []string{"host", "content-md5", "content-length", "content-type"}
	}

	result := []string{}
	signedHeaders := []string{}
	for k, v := range headers {
		k = strings.ToLower(k)
		if strings.HasPrefix(k, httplib.BCE_PREFIX) || utils.IsStringInSlice(k, headersToSign) {
			key := utils.UriEncode(k)
			value := utils.UriEncode(v)
			strTmp := fmt.Sprintf("%s:%s", key, value)
			result = append(result, strTmp)
			signedHeaders = append(signedHeaders, fmt.Sprint(key))
		}
	}
	sort.Strings(result)

	return strings.Join(result, "\n"), strings.Join(signedHeaders, ";")
}

func getCannonicalQuery(query string) string {
	if query == "" {
		return ""
	}

	result := []string{}
	for _, v := range strings.Split(query, "&") {
		tags := strings.Split(v, "=")
		if len(tags) == 2 {
			key := utils.UriEncode(tags[0])
			value := utils.UriEncode(tags[1])
			result = append(result, fmt.Sprintf("%s=%s", key, value))
		} else if len(tags) == 1 {
			key := utils.UriEncode(tags[0])
			result = append(result, fmt.Sprintf("%s=", key))
		}
	}
	sort.Strings(result)

	return strings.Join(result, "&")
}

func Sign(credentials *BceCredentials, timestamp, httpMethod, path, query string,
	headers map[string]string) string {

	if path[0] != '/' {
		path = "/" + path
	}

	var expirationPeriodInSeconds = 1800
	authStringPrefix := fmt.Sprintf("bce-auth-v1/%s/%s/%d", credentials.AccessKeyId,
		timestamp, expirationPeriodInSeconds)
	//fmt.Println(authStringPrefix)

	mac := hmac.New(sha256.New, []byte(credentials.SecretAccessKey))
	mac.Write([]byte(authStringPrefix))
	signingKey := fmt.Sprintf("%x", mac.Sum(nil))
	//fmt.Printf(signingKey)

	CanonicalURI := utils.UriEncodeExceptSlash(path)
	CanonicalQueryString := getCannonicalQuery(query)
	CanonicalHeaders, signedHeaders := getCanonicalHeaders(headers, nil)
	CanonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s", httpMethod, CanonicalURI,
		CanonicalQueryString, CanonicalHeaders)

	mac = hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(CanonicalRequest))
	signature := fmt.Sprintf("%x", mac.Sum(nil))
	//fmt.Println(signature)

	authorization := fmt.Sprintf("%s/%s/%s", authStringPrefix, signedHeaders, signature)
	if Debug {
		fmt.Println(CanonicalRequest)
		fmt.Println(authorization)
	}
	return authorization
}

/* vim: set expandtab ts=4 sw=4 sts=4 tw=100: */
