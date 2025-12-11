package middlewarerequest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	wminio "bytetrade.io/web3os/tapr/pkg/workload/minio"

	"github.com/minio/madmin-go"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
)

func (c *controller) createOrUpdateMinioRequest(req *aprv1.MiddlewareRequest) error {
	adminUser, adminPassword, err := c.findMinioAdminCredentials(req.Namespace)
	if err != nil {
		return fmt.Errorf("failed to find minio admin credentials: %w", err)
	}

	endpoint, err := c.getMinioEndpoint()
	if err != nil {
		return fmt.Errorf("failed to get minio endpoint: %w", err)
	}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(adminUser, adminPassword, ""),
		Secure: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create minio client: %w", err)
	}

	madminClient, err := madmin.New(endpoint, adminUser, adminPassword, false)
	if err != nil {
		return fmt.Errorf("failed to create minio admin client: %v", err)
	}

	klog.Info("create minio user and buckets, ", req.Spec.Minio.User)

	userPassword, err := req.Spec.Minio.Password.GetVarValue(c.ctx, c.k8sClientSet, req.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get user password: %w", err)
	}

	err = c.createOrUpdateMinioUser(c.ctx, madminClient, req.Spec.Minio.User, userPassword)
	if err != nil {
		return fmt.Errorf("failed to create or update minio user: %w", err)
	}

	bucketList := make([]string, 0, len(req.Spec.Minio.Buckets))
	for _, bucket := range req.Spec.Minio.Buckets {
		bucketName := wminio.GetBucketName(req.Spec.AppNamespace, bucket.Name)
		klog.Info("create bucket for user, ", bucketName, ", ", req.Spec.Minio.User)
		bucketList = append(bucketList, bucketName)
		err = minioClient.MakeBucket(c.ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			exists, errBucketExists := minioClient.BucketExists(c.ctx, bucketName)
			if errBucketExists != nil {
				return fmt.Errorf("failed to check bucket name: %s, existence: %w", bucketName, errBucketExists)
			}
			if !exists {
				return fmt.Errorf("failed to create bucket %s: %w", bucketName, err)
			}
			klog.Info("bucket already exists, ", bucketName)
		}
	}

	err = c.setBucketPolicyForUser(c.ctx, madminClient, bucketList, req)
	if err != nil {
		return fmt.Errorf("failed to set bucket policy: %w", err)
	}

	return nil
}

func (c *controller) deleteMinioRequest(req *aprv1.MiddlewareRequest) error {
	adminUser, adminPassword, err := c.findMinioAdminCredentials(req.Namespace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// MinIO admin secret is gone, likely because MinIO middleware was already removed.
			// Nothing to clean up; treat as successful no-op.
			klog.Infof("minio admin secret not found, skipping deletion for user %s", req.Spec.Minio.User)
			return nil
		}
		return fmt.Errorf("failed to find minio admin credentials: %w", err)
	}

	endpoint, err := c.getMinioEndpoint()
	if err != nil {
		return fmt.Errorf("failed to get minio endpoint: %w", err)
	}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(adminUser, adminPassword, ""),
		Secure: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create minio client: %w", err)
	}
	madminClient, err := madmin.New(endpoint, adminUser, adminPassword, false)
	if err != nil {
		return fmt.Errorf("failed to create minio admin client: %v", err)
	}

	klog.Info("delete minio user and buckets, ", req.Spec.Minio.User)

	for _, bucket := range req.Spec.Minio.Buckets {
		bucketName := wminio.GetBucketName(req.Spec.AppNamespace, bucket.Name)
		klog.Info("delete bucket, ", bucketName)

		err = c.removeAllObjectsInBucket(c.ctx, minioClient, bucketName)
		if err != nil {
			klog.Warning("failed to remove objects in bucket ", bucketName, ": ", err)
		}

		err = minioClient.RemoveBucket(c.ctx, bucketName)
		if err != nil {
			klog.Warning("failed to remove bucket ", bucketName, ": ", err)
		}
	}

	// Additionally delete any buckets created with the namespace prefix if allowed
	if req.Spec.Minio.AllowNamespaceBuckets {
		prefix := req.Spec.AppNamespace + "-"
		buckets, err := minioClient.ListBuckets(c.ctx)
		if err != nil {
			klog.Warning("failed to list buckets: ", err)
		} else {
			for _, b := range buckets {
				if strings.HasPrefix(b.Name, prefix) {
					klog.Info("delete prefixed bucket, ", b.Name)
					if err := c.removeAllObjectsInBucket(c.ctx, minioClient, b.Name); err != nil {
						klog.Warning("failed to remove objects in bucket ", b.Name, ": ", err)
					}
					if err := minioClient.RemoveBucket(c.ctx, b.Name); err != nil {
						klog.Warning("failed to remove bucket ", b.Name, ": ", err)
					}
				}
			}
		}
	}

	err = c.deleteMinioUser(c.ctx, madminClient, req.Spec.Minio.User)
	if err != nil {
		return fmt.Errorf("failed to delete minio user: %w", err)
	}

	// Remove the canned policy associated with the user
	policyName := fmt.Sprintf("%s-policy", req.Spec.Minio.User)
	if err := madminClient.RemoveCannedPolicy(c.ctx, policyName); err != nil {
		klog.Warning("failed to remove user policy ", policyName, ": ", err)
	}

	return nil
}

func (c *controller) findMinioAdminCredentials(namespace string) (string, string, error) {
	return wminio.FindMinioAdminUser(c.ctx, c.k8sClientSet, "minio-middleware")
}

func (c *controller) getMinioEndpoint() (string, error) {
	return fmt.Sprintf("minio-minio.%s.svc.cluster.local:9000", "minio-middleware"), nil
}

func (c *controller) createOrUpdateMinioUser(ctx context.Context, madminClient *madmin.AdminClient, username, password string) error {

	err := madminClient.AddUser(ctx, username, password)
	if err != nil {
		return fmt.Errorf("failed to add/update user: %v", err)
	}
	klog.Info("creating or updating minio user: ", username)
	return nil
}

func (c *controller) setBucketPolicyForUser(ctx context.Context, madminClient *madmin.AdminClient, buckets []string, mr *aprv1.MiddlewareRequest) error {
	resources := make([]string, 0, len(buckets)*2)
	for _, bucketName := range buckets {
		resources = append(resources, fmt.Sprintf("arn:aws:s3:::%s", bucketName))
		resources = append(resources, fmt.Sprintf("arn:aws:s3:::%s/*", bucketName))
	}

	// Build base statements
	statements := []map[string]interface{}{
		{
			"Effect":   "Allow",
			"Action":   []string{"s3:ListAllMyBuckets", "s3:HeadBucket", "s3:GetBucketLocation"},
			"Resource": []string{"arn:aws:s3:::*"},
		},
		{
			"Effect":   "Allow",
			"Action":   "s3:*",
			"Resource": resources,
		},
	}

	// If allowed, grant full access on all buckets with AppNamespace prefix and allow creating them

	if mr.Spec.Minio.AllowNamespaceBuckets {
		prefix := mr.Spec.AppNamespace
		statements = append(statements,
			map[string]interface{}{
				"Effect": "Allow",
				"Action": []string{
					"s3:*",
				},
				"Resource": []string{
					fmt.Sprintf("arn:aws:s3:::%s-*", prefix),
				},
			},
			map[string]interface{}{
				"Effect":   "Allow",
				"Action":   "s3:*",
				"Resource": []string{fmt.Sprintf("arn:aws:s3:::%s-*/*", prefix)},
			},
		)
	}

	policy := map[string]interface{}{
		"Version":   "2012-10-17",
		"Statement": statements,
	}

	policyBytes, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal bucket policy: %w", err)
	}

	policyName := fmt.Sprintf("%s-policy", mr.Spec.Minio.User)
	if err := madminClient.AddCannedPolicy(ctx, policyName, policyBytes); err != nil {
		return fmt.Errorf("failed to add canned policy: %w", err)
	}

	if err := madminClient.SetPolicy(ctx, policyName, mr.Spec.Minio.User, false); err != nil {
		return fmt.Errorf("failed to set policy: %s for user: %s, err %v", policyName, mr.Spec.Minio.User, err)
	}

	klog.Infof("set bucket policy for user %s on buckets %v", mr.Spec.Minio.User, buckets)
	return nil
}

func (c *controller) removeAllObjectsInBucket(ctx context.Context, client *minio.Client, bucketName string) error {
	objectsCh := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{Recursive: true})
	for object := range objectsCh {
		if object.Err != nil {
			return object.Err
		}
		err := client.RemoveObject(ctx, bucketName, object.Key, minio.RemoveObjectOptions{})
		if err != nil {
			klog.Warning("failed to remove object ", object.Key, " from bucket ", bucketName, ": ", err)
		}
	}
	return nil
}

func (c *controller) deleteMinioUser(ctx context.Context, madminClient *madmin.AdminClient, username string) error {
	users, err := madminClient.ListUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list users: %v", err)
	}
	if _, exists := users[username]; !exists {
		klog.Infof("User %s does not exist, skipping deletion", username)
		return nil
	}

	err = madminClient.RemoveUser(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to remove user %s: %v", username, err)
	}
	klog.Infof("Deleted minio user: %s", username)
	return nil
}
