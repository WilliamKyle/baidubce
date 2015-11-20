package httplib

// Standard HTTP Headers
const (
	AUTHORIZATION       = "Authorization"
	CONTENT_DISPOSITION = "Content-Disposition"
	CONTENT_ENCODING    = "Content-Encoding"
	CONTENT_LENGTH      = "Content-Length"
	CONTENT_MD5         = "Content-MD5"
	CONTENT_RANGE       = "Content-Range"
	CONTENT_TYPE        = "Content-Type"
	DATE                = "Date"
	ETAG                = "ETag"
	EXPIRES             = "Expires"
	HOST                = "Host"
	LAST_MODIFIED       = "Last-Modified"
	RANGE               = "Range"
	SERVER              = "Server"
	USER_AGENT          = "User-Agent"
)

// BCE Common HTTP Headers

const (
	BCE_PREFIX                  = "x-bce-"
	BCE_ACL                     = "x-bce-acl"
	BCE_CONTENT_SHA256          = "x-bce-content-sha256"
	BCE_COPY_METADATA_DIRECTIVE = "x-bce-metadata-directive"
	BCE_COPY_SOURCE             = "x-bce-copy-source"
	BCE_COPY_SOURCE_IF_MATCH    = "x-bce-copy-source-if-match"
	BCE_DATE                    = "x-bce-date"
	BCE_USER_METADATA_PREFIX    = "x-bce-meta-"
	BCE_REQUEST_ID              = "x-bce-request-id"
)

// BOS HTTP Headers

const BOS_DEBUG_ID = "x-bce-bos-debug-id"
