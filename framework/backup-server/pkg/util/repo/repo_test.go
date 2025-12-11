package repo

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestFormatCosUrl(t *testing.T) {

	var url = "https://MYBUCKET.cos.ap-tokyo.myqcloud.com/subfolder/olares-backups/name-00000000-0000-0000-0000-000000000000"
	var raw, _ = FormatCosByRawUrl(url)
	assert.Equal(t, raw.Bucket, "MYBUCKET")
	assert.Equal(t, raw.Region, "ap-tokyo")
	assert.Equal(t, raw.Prefix, "subfolder/olares-backups/name-00000000-0000-0000-0000-000000000000")
	assert.Equal(t, raw.Endpoint, "https://cos.ap-tokyo.myqcloud.com/MYBUCKET/subfolder/olares-backups/name-00000000-0000-0000-0000-000000000000")

	url = "https://MYBUCKET.cos.ap-tokyo.myqcloud.com/olares-backups/name-00000000-0000-0000-0000-000000000000"
	raw, _ = FormatCosByRawUrl(url)
	assert.Equal(t, raw.Bucket, "MYBUCKET")
	assert.Equal(t, raw.Region, "ap-tokyo")
	assert.Equal(t, raw.Prefix, "olares-backups/name-00000000-0000-0000-0000-000000000000")
	assert.Equal(t, raw.Endpoint, "https://cos.ap-tokyo.myqcloud.com/MYBUCKET/olares-backups/name-00000000-0000-0000-0000-000000000000")

	//
	url = "https://cos.ap-tokyo.myqcloud.com/MYBUCKET/olares-backups/name-00000000-0000-0000-0000-000000000000"
	raw, _ = FormatCosByRawUrl(url)
	assert.Equal(t, raw.Bucket, "MYBUCKET")
	assert.Equal(t, raw.Region, "ap-tokyo")
	assert.Equal(t, raw.Prefix, "olares-backups/name-00000000-0000-0000-0000-000000000000")
	assert.Equal(t, raw.Endpoint, "https://cos.ap-tokyo.myqcloud.com/MYBUCKET/olares-backups/name-00000000-0000-0000-0000-000000000000")
}

func TestFormatCosEndpointA(t *testing.T) {
	var schema = "https"
	var host = "cos.ap-beijing.myqcloud.com"
	var path = "MYBUCKET/folder1/olares-backups/name-00000000-0000-0000-0000-000000000000"
	var ep, _ = FormatCos(schema, host, path)
	assert.Equal(t, ep.Endpoint, "https://cos.ap-beijing.myqcloud.com/MYBUCKET/folder1/olares-backups/name-00000000-0000-0000-0000-000000000000")
}

func TestFormatCosEndpointB(t *testing.T) {
	var schema = "https"
	var host = "MYBUCKET.cos.ap-beijing.myqcloud.com"
	var path = "folder1/olares-backups/name-00000000-0000-0000-0000-000000000000"
	var ep, _ = FormatCos(schema, host, path)
	assert.Equal(t, ep.Endpoint, "https://cos.ap-beijing.myqcloud.com/MYBUCKET/folder1/olares-backups/name-00000000-0000-0000-0000-000000000000")
}
