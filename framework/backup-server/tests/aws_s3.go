package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/pflag"
	"olares.com/backup-server/pkg/util"
	"olares.com/backup-server/pkg/util/log"
	"olares.com/backup-server/pkg/util/pointer"
)

var (
	endpoint  = "http://192.168.50.35:9000"
	region    = "minio"
	ak        = "minioadmin"
	sk        = "RAzvq6LXRjw39zcp"
	bucket    = "olares"
	prefix    = "system-backups"
	delimiter = "/"

	listBuckets     bool
	listObjects     bool
	uploadFilePath  string
	uploadKeyPrefix = "osdata-backups"
)

func createS3Client() (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithRetryMaxAttempts(1),
	)

	if endpoint != "" {
		cfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               endpoint,
				SigningRegion:     region,
				HostnameImmutable: true,
			}, nil
		})
	}

	if ak != "" && sk != "" {
		cfg.Credentials = credentials.NewStaticCredentialsProvider(ak, sk, "")
	}

	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg), nil
}

func init() {
	log.InitLog("debug")
}

func main() {
	pflag.StringVar(&endpoint, "endpoint", "", "s3 endpoint")
	pflag.StringVar(&region, "region", "", "s3 region")
	pflag.StringVar(&ak, "ak", "", "s3 access key")
	pflag.StringVar(&sk, "sk", "", "s3 secret key")
	pflag.StringVar(&bucket, "bucket", "", "s3 bucket")
	pflag.StringVar(&prefix, "prefix", "", "s3 prefix")
	pflag.StringVar(&delimiter, "delimiter", "/", "s3 objects list delimiter")
	pflag.BoolVar(&listBuckets, "list-buckets", false, "whether to list buckets")
	pflag.BoolVar(&listObjects, "list-objects", false, "whether to list objects")
	pflag.StringVar(&uploadKeyPrefix, "upload-keyprefix", "osdata-backups", "upload object key prefix")
	pflag.StringVar(&uploadFilePath, "upload-filepath", "", "upload file path")

	pflag.Parse()

	log.Debugw("flags", "endpoint", endpoint,
		"region", region,
		"ak", ak,
		"sk", sk,
		"bucket", bucket,
		"prefix", prefix,
		"delimiter", delimiter,
		"listBuckets", listBuckets,
		"uploadKeyPrefix", uploadKeyPrefix,
		"uploadSourceFilepath", uploadFilePath,
		"listObjects", listObjects,
	)

	s3Client, err := createS3Client()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	if listBuckets {
		log.Debug("list buckets")
		// list buckets
		buckets, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("list buckets:\n%s\n", util.PrettyJSON(buckets))
	}

	// objects
	if listObjects {

		log.Debug("list dir objects")

		objects, err := s3Client.ListObjects(ctx, &s3.ListObjectsInput{
			Bucket:    pointer.String(bucket),
			Prefix:    pointer.String(prefix),
			Delimiter: pointer.String(delimiter),
		})
		if err != nil {
			log.Fatal(err)
		}

		log.Infof("list %q bucket %d dirs:", bucket, len(objects.CommonPrefixes))

		for _, o := range objects.CommonPrefixes {
			fmt.Printf("%s\n", *o.Prefix)
		}
	}

	// upload file
	if uploadFilePath != "" {
		log.Debug("upload file")
		fileName := filepath.Base(uploadFilePath)
		keyName := strings.TrimRight(uploadKeyPrefix, "/") + "/" + fileName

		f, err := os.Open(uploadFilePath)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		uploader := manager.NewUploader(s3Client)
		result, err := uploader.Upload(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(keyName),
			Body:   f,
		})
		if err != nil {
			log.Fatal(err)
		}

		log.Infof("s3 upload file result: %s", util.PrettyJSON(result))
	}
}
