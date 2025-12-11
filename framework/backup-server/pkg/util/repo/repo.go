package repo

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"olares.com/backup-server/pkg/util"
)

type RepositoryInfo struct {
	Region   string `json:"region"`
	Bucket   string `json:"bucket"`
	Prefix   string `json:"prefix"`
	Endpoint string `json:"endpoint"`
	Suffix   string `json:"suffix"`
}

func FormatSpace(schema, host, path string) (*RepositoryInfo, error) {
	var bucket, region, prefix, endpoint, suffix string
	var err error
	var hosts = strings.Split(host, ".")
	if len(hosts) != 4 {
		return nil, fmt.Errorf("invalid space host: %s", host)
	}
	var hostPrefix = hosts[0]
	if hostPrefix != "cos" && hostPrefix != "s3" {
		return nil, fmt.Errorf("invalid space host: %s", host)
	}
	region = hosts[1]
	var p = strings.TrimPrefix(path, "/")
	var pt = strings.Split(p, "/")
	if pt == nil || len(pt) == 0 {
		return nil, fmt.Errorf("invalid space path: %s", path)
	}

	suffix, err = util.GetSuffix(pt[1], "-")
	if err != nil {
		return nil, fmt.Errorf("invalid space path: %s", path)
	}

	bucket = pt[0]
	if len(pt) > 1 {
		prefix = strings.Join(pt[1:], "/")
	}

	if hostPrefix == "s3" {
		endpoint = fmt.Sprintf("%s://s3.%s.amazonaws.com/%s%s", schema, region, bucket, prefix)
	} else {
		endpoint = fmt.Sprintf("%s://cos.%s.myqcloud.com/%s%s", schema, region, bucket, prefix)
	}

	var result = &RepositoryInfo{
		Region:   region,
		Bucket:   bucket,
		Prefix:   prefix,
		Endpoint: endpoint,
		Suffix:   suffix,
	}

	return result, nil
}

func FormatCosByRawUrl(rawurl string) (*RepositoryInfo, error) {
	var region, bucket, prefix, endpoint string
	var err error
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	var host = u.Host
	var schema = u.Scheme
	var path = u.Path

	var hosts = strings.Split(host, ".")
	if len(hosts) != 4 && len(hosts) != 5 {
		return nil, fmt.Errorf("invalid cos host: %s", host)
	}

	// if hosts[0] != "cos" {
	// 	return nil, fmt.Errorf("invalid cos host: %s", host)
	// }

	// region = hosts[1]

	var p = strings.TrimPrefix(path, "/")
	var pt = strings.Split(p, "/")
	if pt == nil || len(pt) == 0 {
		return nil, fmt.Errorf("invalid cos path: %s", path)
	}

	if len(hosts) == 4 {
		bucket = pt[0]
		region = hosts[1]
		// if len(pt) > 1 {
		prefix = strings.Join(pt[1:], "/")
		// }
	} else {
		bucket = hosts[0]
		region = hosts[2]
		// if len(pt) > 0 {
		prefix = strings.Join(pt[0:], "/")
		// }
	}

	// bucket = pt[0]
	// if len(hosts) == 4 {

	// } else {

	// }

	if prefix != "" {
		endpoint = fmt.Sprintf("%s://cos.%s.myqcloud.com/%s/%s", schema, region, bucket, strings.TrimRight(prefix, "/"))
	} else {
		endpoint = fmt.Sprintf("%s://cos.%s.myqcloud.com/%s", schema, region, bucket)
	}

	var result = &RepositoryInfo{
		Region:   region,
		Bucket:   bucket,
		Prefix:   prefix,
		Endpoint: endpoint,
	}
	return result, nil
}

func FormatCos(schema, host, path string) (*RepositoryInfo, error) {
	var region, bucket, prefix, endpoint string

	var hosts = strings.Split(host, ".")
	if len(hosts) != 4 && len(hosts) != 5 {
		return nil, fmt.Errorf("invalid cos host: %s", host)
	}

	// if hosts[0] != "cos" {
	// 	return nil, fmt.Errorf("invalid cos host: %s", host)
	// }

	var p = strings.TrimPrefix(path, "/")
	var pt = strings.Split(p, "/")
	if pt == nil || len(pt) == 0 {
		return nil, fmt.Errorf("invalid cos path: %s", path)
	}

	if len(hosts) == 4 {
		if len(pt) > 1 {
			prefix = strings.Join(pt[1:], "/")
		}
	} else {
		if len(pt) > 0 {
			prefix = strings.Join(pt[0:], "/")
		}
	}

	if len(hosts) == 4 {
		bucket = pt[0]
		region = hosts[1]
	} else {
		bucket = hosts[0]
		region = hosts[2]
	}

	if prefix != "" {
		endpoint = fmt.Sprintf("%s://cos.%s.myqcloud.com/%s/%s", schema, region, bucket, strings.TrimRight(prefix, "/"))
	} else {
		endpoint = fmt.Sprintf("%s://cos.%s.myqcloud.com/%s", schema, region, bucket)
	}

	var result = &RepositoryInfo{
		Region:   region,
		Bucket:   bucket,
		Prefix:   prefix,
		Endpoint: endpoint,
	}
	return result, nil
}

func FormatS3(rawurl string) (*RepositoryInfo, error) {
	var region, bucket, prefix, endpoint string
	var err error

	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	host := u.Host
	path := strings.TrimPrefix(u.Path, "/")

	parts := strings.Split(host, ".")

	if len(parts) < 3 || parts[len(parts)-2] != "amazonaws" || parts[len(parts)-1] != "com" {
		return nil, errors.New("host is not a valid amazonaws.com domain")
	}

	switch {
	case strings.HasPrefix(host, "s3."):
		if len(parts) < 4 {
			return nil, errors.New("host format invalid for s3.region.amazonaws.com")
		}
		region = parts[1]
		pathParts := strings.SplitN(path, "/", 2)
		if len(pathParts) < 1 || pathParts[0] == "" {
			return nil, errors.New("bucket not found in path")
		}
		bucket = pathParts[0]
		if len(pathParts) == 2 {
			prefix = pathParts[1]
		} else {
			prefix = ""
		}

	case len(parts) >= 5 && parts[1] == "s3":
		bucket = parts[0]
		region = parts[2]
		prefix = path

	case len(parts) >= 4:
		bucket = parts[0]
		region = parts[1]
		prefix = path

	default:
		return nil, errors.New("host format not recognized")
	}

	if prefix != "" {
		endpoint = fmt.Sprintf("s3:https://s3.%s.amazonaws.com/%s/%s", region, bucket, strings.TrimRight(prefix, "/"))
	} else {
		endpoint = fmt.Sprintf("s3:https://s3.%s.amazonaws.com/%s", region, bucket)
	}

	endpoint = strings.TrimRight(endpoint, "/")

	var result = &RepositoryInfo{
		Region:   region,
		Bucket:   bucket,
		Prefix:   prefix,
		Endpoint: endpoint,
	}

	return result, nil
}
