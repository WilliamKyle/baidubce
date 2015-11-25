package bos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/WilliamKyle/baidubce/auth"
	"github.com/WilliamKyle/baidubce/httplib"
	"github.com/WilliamKyle/baidubce/utils"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

const (
	DefaultLocation   = "bj"
	DefaultAPIVersion = "v1"
)

type Client struct {
	Credential *auth.BceCredentials
	Location   string
	APIVersion string
	Host       string
	Debug      bool
}

func NewClient(credential *auth.BceCredentials) (Client, error) {
	return Client{
		Credential: credential,
		Location:   DefaultLocation,
		APIVersion: DefaultAPIVersion,
		Debug:      false,
	}, nil
}

func (c Client) GetBaseURL() string {
	return fmt.Sprintf("%s/%s", c.GetEndpoint(), c.APIVersion)
}

func (c Client) GetEndpoint() string {
	return fmt.Sprintf("http://%s", c.GetHost())
}

func (c Client) GetHost() string {
	if c.Host != "" {
		return c.Host
	}
	return fmt.Sprintf("%s.bcebos.com", c.Location)
}

type ErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"requestId"`
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("Service returned error: Code=%s, RequestId=%s, Message=%s", e.Code, e.RequestId, e.Message)
}

func (c Client) doRequest(req *httplib.Request) (*http.Response, error) {
	if req.BaseUrl == "" {
		req.BaseUrl = c.GetBaseURL()
	}
	req.Headers[httplib.HOST] = c.GetHost()

	timestamp := utils.GetHttpHeadTimeStamp()
	auth.Debug = c.Debug
	authorization := auth.Sign(c.Credential, timestamp, req.Method, req.Path, req.Query, req.Headers)

	req.Headers[httplib.BCE_DATE] = timestamp
	req.Headers[httplib.AUTHORIZATION] = authorization

	httplib.Debug = c.Debug
	res, err := httplib.Run(req, nil)
	if err != nil {
		return res, err
	}

	if res.StatusCode != 200 {
		errR := &ErrorResponse{}
		if req.Method == httplib.HEAD || req.Method == httplib.DELETE {
			errR.Code = fmt.Sprintf("%d", res.StatusCode)
			errR.Message = res.Status
			errR.RequestId = "EMPTY"
		} else {
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return res, err
			}
			j := json.NewDecoder(strings.NewReader(string(body)))
			j.Decode(&errR)
		}
		return res, errR
	}
	return res, err
}

/*************************************************************************************************

Bucket Opreation Method

*************************************************************************************************/

/*
 * Name: GetBucketLocation
 * URL: http://bce.baidu.com/doc/BOS/API.html#GetBucketLocation.E6.8E.A5.E5.8F.A3
 */

type BucketLocationResponse struct {
	LocationConstraint string `json:"locationConstraint"`
}

func (c Client) GetBucketLocation(bucketName string) (output BucketLocationResponse, err error) {
	req := &httplib.Request{
		Method:  httplib.GET,
		Headers: map[string]string{},
		Query:   "location",
		Path:    c.APIVersion + "/" + bucketName,
	}

	res, err := c.doRequest(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	j := json.NewDecoder(strings.NewReader(string(body)))
	j.Decode(&output)
	return
}

/*
 * Name: GetService/ListBucket
 * URL: http://bce.baidu.com/doc/BOS/API.html#GetService.2FListBucket.E6.8E.A5.E5.8F.A3
 */

type OwnerInfo struct {
	DisplayName string
	Id          string
}

type BucketInfo struct {
	CreationDate string
	Location     string
	Name         string
}

type ListBucketResponse struct {
	Owner   OwnerInfo
	Buckets []BucketInfo
}

func (c Client) ListBucket() (output ListBucketResponse, err error) {
	req := &httplib.Request{
		Method:  httplib.GET,
		Headers: map[string]string{},
		Path:    c.APIVersion + "/",
	}

	res, err := c.doRequest(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	j := json.NewDecoder(strings.NewReader(string(body)))
	j.Decode(&output)
	return
}

/*
 * Name: PutBucket
 * URL: http://bce.baidu.com/doc/BOS/API.html#PutBucket.E6.8E.A5.E5.8F.A3
 */

func (c Client) PutBucket(bucketName string) (err error) {
	req := &httplib.Request{
		Method:  httplib.PUT,
		Headers: map[string]string{},
		Path:    c.APIVersion + "/" + bucketName,
	}

	_, err = c.doRequest(req)
	return
}

/*
 * Name: GetBucket/ListObjects
 * URL: http://bce.baidu.com/doc/BOS/API.html#GetBucket.2FListObjects.E6.8E.A5.E5.8F.A3
 */

type ObjectInfo struct {
	ObjectName   string `json:"key"`
	LastModified string
	ETag         string
	Size         int64
	Owner        OwnerInfo
}

type ListObjectsResponse struct {
	Name        string
	Prefix      string
	Delimiter   string
	Marker      string
	MaxKeys     int
	IsTruncated bool
	Contents    []ObjectInfo
}

// TODO: Need test
func (c Client) ListObjects(bucketName string, delimiter, marker, maxKeys, prefix interface{}) (output ListObjectsResponse, err error) {
	query := []string{}
	if delimiter != nil {
		query = append(query, "delimiter="+delimiter.(string))
	}
	if marker != nil {
		query = append(query, "marker="+marker.(string))
	}
	if maxKeys != nil {
		query = append(query, "maxKeys="+maxKeys.(string))
	}
	if prefix != nil {
		prefix = c.formatPath(prefix.(string))
		query = append(query, "prefix="+prefix.(string))
	}
	req := &httplib.Request{
		Method:  httplib.GET,
		Headers: map[string]string{},
		Query:   strings.Join(query, "&"),
		Path:    c.APIVersion + "/" + bucketName,
	}

	res, err := c.doRequest(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	j := json.NewDecoder(strings.NewReader(string(body)))
	j.Decode(&output)
	return
}

/*
 * Name: HeadBucket
 * URL: http://bce.baidu.com/doc/BOS/API.html#HeadBucket.E6.8E.A5.E5.8F.A3
 */

func (c Client) HeadBucket(bucketName string) (err error) {
	req := &httplib.Request{
		Method:  httplib.HEAD,
		Headers: map[string]string{},
		Path:    c.APIVersion + "/" + bucketName,
	}

	_, err = c.doRequest(req)
	return
}

/*
 * Name: DeleteBucket
 * URL: http://bce.baidu.com/doc/BOS/API.html#DeleteBucket.E6.8E.A5.E5.8F.A3
 */

func (c Client) DeleteBucket(bucketName string) (err error) {
	req := &httplib.Request{
		Method:  httplib.DELETE,
		Headers: map[string]string{},
		Path:    c.APIVersion + "/" + bucketName,
	}

	_, err = c.doRequest(req)
	if err != nil {
		if strings.HasPrefix(err.Error(), "Service returned error: Code=204, ") {
			return nil
		}
	}
	return
}

/*
 * Name: GetBucketAcl
 * URL: http://bce.baidu.com/doc/BOS/API.html#GetBucketAcl.E6.8E.A5.E5.8F.A3-1
 */

type GranteeInfo struct {
	Id string
}

type GranteeGroup struct {
	Grantee    []GranteeInfo
	Permission []string
}

type BucketAclResponse struct {
	Owner             GranteeInfo
	AccessControlList []GranteeGroup
}

func (c Client) GetBucketAcl(bucketName string) (output BucketAclResponse, err error) {
	req := &httplib.Request{
		Method:  httplib.GET,
		Headers: map[string]string{},
		Query:   "acl",
		Path:    c.APIVersion + "/" + bucketName,
	}

	res, err := c.doRequest(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	j := json.NewDecoder(strings.NewReader(string(body)))
	j.Decode(&output)
	return
}

/*
 * Name: SetBucketAcl
 * URL: http://bce.baidu.com/doc/BOS/API.html#SetBucketAcl.E6.8E.A5.E5.8F.A3
 */

// TODO: Upload ACL filed not implement
func (c Client) SetBucketAcl(bucketName string, cannedAcl string) (err error) {
	req := &httplib.Request{
		Method:  httplib.PUT,
		Headers: map[string]string{httplib.BCE_ACL: cannedAcl},
		Query:   "acl",
		Path:    c.APIVersion + "/" + bucketName,
	}

	_, err = c.doRequest(req)
	return
}

/*************************************************************************************************

Object Opreation Method

*************************************************************************************************/

/*
 * Name: PutObject
 * URL: http://bce.baidu.com/doc/BOS/API.html#PutObject.E6.8E.A5.E5.8F.A3
 */

func (c Client) PutObject(bucketName, objectName string, body *bytes.Reader, contentMD5, contentSHA256 string, metaInfo map[string]string) (eTag string, err error) {
	objectName = c.formatPath(objectName)
	req := &httplib.Request{
		Method:  httplib.PUT,
		Headers: map[string]string{},
		Path:    c.APIVersion + "/" + bucketName + "/" + objectName,
	}

	if contentMD5 != "" {
		req.Headers[httplib.CONTENT_MD5] = contentMD5
	}

	if contentSHA256 != "" {
		req.Headers[httplib.BCE_CONTENT_SHA256] = contentSHA256
	}

	if metaInfo != nil {
		for k, v := range metaInfo {
			bceMeta := fmt.Sprintf("%s%s", httplib.BCE_USER_METADATA_PREFIX, k)
			req.Headers[bceMeta] = v
		}
	}

	req.Body = body

	res, err := c.doRequest(req)
	if err == nil {
		eTag = strings.Replace(res.Header.Get("Etag"), "\"", "", -1)
	}
	return
}

/*
 * Name: InitiateMultipartUpload
 * URL: http://bce.baidu.com/doc/BOS/API.html#InitiateMultipartUpload.E6.8E.A5.E5.8F.A3
 */

type MultipartUploadResponse struct {
	BucketName string `json:"bucket"`
	ObjectName string `json:"key"`
	UploadId   string `json:"uploadId"`
}

func (c Client) InitiateMultipartUpload(bucketName, objectName, contentType string) (output MultipartUploadResponse, err error) {
	objectName = c.formatPath(objectName)
	req := &httplib.Request{
		Method:  httplib.POST,
		Headers: map[string]string{},
		Path:    c.APIVersion + "/" + bucketName + "/" + objectName,
		Query:   "uploads",
	}

	req.Headers[httplib.CONTENT_TYPE] = httplib.OCTET_STREAM
	if contentType != "" {
		req.Headers[httplib.CONTENT_TYPE] = contentType
	}

	res, err := c.doRequest(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	j := json.NewDecoder(strings.NewReader(string(body)))
	j.Decode(&output)

	return
}

/*
 * Name: UploadPart
 * URL: http://bce.baidu.com/doc/BOS/API.html#UploadPart.E6.8E.A5.E5.8F.A3
 */

func (c Client) UploadPart(bucketName, objectName, uploadId, partNumber string, body *bytes.Reader) (eTag string, err error) {
	objectName = c.formatPath(objectName)
	req := &httplib.Request{
		Method:  httplib.PUT,
		Headers: map[string]string{},
		Path:    c.APIVersion + "/" + bucketName + "/" + objectName,
		Query:   "uploadId=" + uploadId + "&partNumber=" + partNumber,
	}

	req.Body = body

	res, err := c.doRequest(req)
	if err == nil {
		eTag = strings.Replace(res.Header.Get("Etag"), "\"", "", -1)
	}
	return
}

/*
 * Name: CompleteMultipartUpload
 * URL: http://bce.baidu.com/doc/BOS/API.html#CompleteMultipartUpload.E6.8E.A5.E5.8F.A3
 */

type PartInfo struct {
	PartNumber   int    `json:"partNumber"`
	ETag         string `json:"eTag"`
	LastModified string `json:"lastModified"`
	Size         int    `json:"size"`
}

type CompleteMultipartUploadResponse struct {
	Location   string `json:"location"`
	BuketName  string `json:"bucket"`
	ObjectName string `json:"key"`
	ETag       string `json:"eTag"`
}

func (c Client) CompleteMultipartUpload(bucketName, objectName, uploadId string, parts []PartInfo) (output CompleteMultipartUploadResponse, err error) {
	objectName = c.formatPath(objectName)
	req := &httplib.Request{
		Method:  httplib.POST,
		Headers: map[string]string{},
		Path:    c.APIVersion + "/" + bucketName + "/" + objectName,
		Query:   "uploadId=" + uploadId,
	}

	uploadInfo := map[string][]PartInfo{"parts": parts}
	jstring, err := json.Marshal(uploadInfo)
	req.Body = bytes.NewReader(jstring)
	req.Type = httplib.TEXT

	res, err := c.doRequest(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	j := json.NewDecoder(strings.NewReader(string(body)))
	j.Decode(&output)
	return
}

/*
 * Name: AbortMultipartUpload
 * URL: http://bce.baidu.com/doc/BOS/API.html#AbortMultipartUpload.E6.8E.A5.E5.8F.A3
 */

func (c Client) AbortMultipartUpload(bucketName, objectName, uploadId string) (err error) {
	objectName = c.formatPath(objectName)
	req := &httplib.Request{
		Method:  httplib.DELETE,
		Headers: map[string]string{},
		Path:    c.APIVersion + "/" + bucketName + "/" + objectName,
		Query:   "uploadId=" + uploadId,
	}

	_, err = c.doRequest(req)
	if err != nil {
		if strings.HasPrefix(err.Error(), "Service returned error: Code=204, ") {
			return nil
		}
	}
	return
}

/*
 * Name: ListParts
 * URL: http://bce.baidu.com/doc/BOS/API.html#ListParts.E6.8E.A5.E5.8F.A3
 */

type ListPartsResponse struct {
	BucketName           string `json:"bucket"`
	ObjectName           string `json:"key"`
	UploadId             string `json:"uploadId"`
	Initiated            string `json:"initiated"`
	Owner                OwnerInfo
	PartNumberMarker     int  `json:"partNumberMarker"`
	NextPartNumberMarker int  `json:"nextPartNumberMarker"`
	MaxParts             int  `json:"maxParts"`
	IsTruncated          bool `json:"isTruncated"`
	Parts                []PartInfo
}

func (c Client) ListParts(bucketName, objectName, uploadId string, partNumberMarker, maxParts interface{}) (output ListPartsResponse, err error) {
	objectName = c.formatPath(objectName)
	query := []string{}
	query = append(query, "uploadId="+uploadId)
	if partNumberMarker != nil {
		query = append(query, "partNumberMarker="+partNumberMarker.(string))
	}
	if maxParts != nil {
		query = append(query, "maxParts="+maxParts.(string))
	}
	req := &httplib.Request{
		Method:  httplib.GET,
		Headers: map[string]string{},
		Path:    c.APIVersion + "/" + bucketName + "/" + objectName,
		Query:   strings.Join(query, "&"),
	}

	res, err := c.doRequest(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	j := json.NewDecoder(strings.NewReader(string(body)))
	j.Decode(&output)

	return
}

/*
 * Name: ListMultipartUploads
 * URL: http://bce.baidu.com/doc/BOS/API.html#ListMultipartUploads.E6.8E.A5.E5.8F.A3
 */

type UploadInfo struct {
	Owner      OwnerInfo
	ObjectName string `json:"key"`
	UploadId   string
	Initiated  string
}
type ListMultipartUploadsResponse struct {
	BucketName     string `json:"bucket"`
	CommonPrefixes string `json:"commonPrefixes"`
	Prefix         string `json:"prefix"`
	KeyMarker      string `json:"keyMarker"`
	NextKeyMarker  string `json:"nextMarker"`
	MaxUploads     int64  `json:"maxUploads"`
	IsTruncated    bool   `json:"isTruncated"`
	Uploads        []UploadInfo
}

func (c Client) ListMultipartUploads(bucketName string, delimiter, keyMarker, maxUploads, prefix interface{}) (output ListMultipartUploadsResponse, err error) {
	query := []string{}
	query = append(query, "uploads=")
	if delimiter != nil {
		query = append(query, "delimiter="+delimiter.(string))
	}
	if delimiter != nil {
		query = append(query, "delimiter="+delimiter.(string))
	}
	if keyMarker != nil {
		query = append(query, "keyMarker="+keyMarker.(string))
	}
	if maxUploads != nil {
		query = append(query, "maxUploads="+maxUploads.(string))
	}
	if prefix != nil {
		prefix = c.formatPath(prefix.(string))
		query = append(query, "prefix="+prefix.(string))
	}
	req := &httplib.Request{
		Method:  httplib.GET,
		Headers: map[string]string{},
		Path:    c.APIVersion + "/" + bucketName,
		Query:   strings.Join(query, "&"),
	}

	res, err := c.doRequest(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	j := json.NewDecoder(strings.NewReader(string(body)))
	j.Decode(&output)

	return
}

/*
 * Name: CopyObject
 * URL: http://bce.baidu.com/doc/BOS/API.html#CopyObject.E6.8E.A5.E5.8F.A3
 */

type CopyObjectResponse struct {
	LastModified string `json:"lastModified"`
	ETag         string `json:"eTag"`
}

func (c Client) CopyObject(srcBucketName, srcObjectName, destBucketName, destObjectName, eTag, metaDirect string) (output CopyObjectResponse, err error) {
	destObjectName = c.formatPath(destObjectName)
	req := &httplib.Request{
		Method:  httplib.PUT,
		Headers: map[string]string{},
		Path:    c.APIVersion + "/" + destBucketName + "/" + destObjectName,
	}
	req.Headers[httplib.BCE_COPY_SOURCE] = utils.UriEncodeExceptSlash("/" + srcBucketName + "/" + srcObjectName)

	if eTag != "" {
		req.Headers[httplib.BCE_COPY_SOURCE_IF_MATCH] = eTag
	}

	if metaDirect == "copy" || metaDirect == "replace" {
		req.Headers[httplib.BCE_COPY_METADATA_DIRECTIVE] = metaDirect
	}

	res, err := c.doRequest(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	j := json.NewDecoder(strings.NewReader(string(body)))
	j.Decode(&output)

	return
}

/*
 * Name: GetObject
 * URL: http://bce.baidu.com/doc/BOS/API.html#GetObject.E6.8E.A5.E5.8F.A3
 */
type GetObjectResponse struct {
	Body io.ReadCloser
	ETag string
	Size int
	Meta map[string]string
}

func (c Client) GetObject(bucketName, objectName string, startPos, endPos int64) (output GetObjectResponse, err error) {
	objectName = c.formatPath(objectName)
	req := &httplib.Request{
		Method:  httplib.GET,
		Headers: map[string]string{},
		Path:    c.APIVersion + "/" + bucketName + "/" + objectName,
	}
	if startPos >= 0 && endPos > 0 {
		if endPos > startPos {
			req.Headers[httplib.RANGE] = fmt.Sprintf("bytes=%d-%d", startPos, endPos)
		} else {
			req.Headers[httplib.RANGE] = fmt.Sprintf("bytes=%d-", startPos)
		}
	}

	res, err := c.doRequest(req)
	if err != nil {
		return
	}
	output = GetObjectResponse{Body: res.Body, Meta: map[string]string{}}
	eTag := strings.Replace(res.Header.Get("Etag"), "\"", "", -1)
	output.ETag = eTag
	size, _ := strconv.Atoi(res.Header.Get("Content-Length"))
	output.Size = size

	for k, v := range res.Header {
		if strings.HasPrefix(strings.ToLower(k), httplib.BCE_USER_METADATA_PREFIX) {
			output.Meta[k] = v[0]
		}
	}
	return
}

/*
 * Name: GetObjectMeta
 * URL: http://bce.baidu.com/doc/BOS/API.html#GetObjectMeta.E6.8E.A5.E5.8F.A3
 */

func (c Client) GetObjectMeta(bucketName, objectName string) (output map[string]string, err error) {
	objectName = c.formatPath(objectName)
	req := &httplib.Request{
		Method:  httplib.HEAD,
		Headers: map[string]string{},
		Path:    c.APIVersion + "/" + bucketName + "/" + objectName,
	}

	res, err := c.doRequest(req)
	if err != nil {
		return
	}
	output = map[string]string{}
	output["Size"] = res.Header.Get("Content-Length")
	output["eTag"] = res.Header.Get("ETag")
	for k, v := range res.Header {
		if strings.HasPrefix(strings.ToLower(k), httplib.BCE_USER_METADATA_PREFIX) {
			output[k] = v[0]
		}
	}

	return
}

/*
 * Name: DeleteObject
 * URL: http://bce.baidu.com/doc/BOS/API.html#DeleteObject.E6.8E.A5.E5.8F.A3
 */

func (c Client) DeleteObject(bucketName, objectName string) (err error) {
	objectName = c.formatPath(objectName)
	req := &httplib.Request{
		Method:  httplib.DELETE,
		Headers: map[string]string{},
		Path:    c.APIVersion + "/" + bucketName + "/" + objectName,
	}

	_, err = c.doRequest(req)
	if err != nil {
		if strings.HasPrefix(err.Error(), "Service returned error: Code=204, ") {
			return nil
		}
	}
	return
}

func (c Client) formatPath(objectName string) string {
	if objectName[0] == '/' {
		return objectName[1:]
	}
	return objectName
}
