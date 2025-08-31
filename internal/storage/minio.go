package storage

import (
	"context"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Minio struct {
	//MinioEndpoint	string
	//MinioAccessKey	string
	//MinioSecretKey	string
	//MinioUseSSL		bool
	MinioBucket string        //Bucket name
	MinioClient *minio.Client //Minio client
	Error       error         //Error
}

func (m *Minio) GetError() error {
	return m.Error
}

func NewMinio(endpoint, access, passkey, bucket string, ssl bool) *Minio {
	cl, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(access, passkey, ""),
		Secure: ssl,
	})
	if err != nil {
		return &Minio{
			MinioBucket: bucket,
			MinioClient: cl,
			Error:       err,
		}
	}
	_, err = cl.BucketExists(context.Background(), bucket)
	if err != nil {
		return &Minio{
			MinioBucket: bucket,
			MinioClient: cl,
			Error:       err,
		}
	}
	return &Minio{
		//MinioEndpoint: endpoint,
		//MinioAccessKey: access,
		//MinioSecretKey: passkey,
		//MinioUseSSL: ssl,
		MinioBucket: bucket,
		MinioClient: cl,
		Error:       err,
	}
}

func (m *Minio) PutObject(c context.Context, key string, r *ProgressReader) (*FileInfo, error) {
	uploadInfo, err := m.MinioClient.PutObject(c, m.MinioBucket, key, r, -1,
		minio.PutObjectOptions{
			UserMetadata: map[string]string{
				"filename": r.Filename,
			},
		})
	return &FileInfo{
		Key:      uploadInfo.Key,
		Filename: r.Filename,
		Size:     uploadInfo.Size,
	}, err
}

func (m *Minio) GetObject(c context.Context, key string) (*FileInfo, error) {
	object, err := m.MinioClient.GetObject(c, m.MinioBucket, key,
		minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	stat, err := object.Stat()
	if err != nil {
		return nil, err
	}
	ret := &FileInfo{
		Object:  object,
		Content: stat.ContentType,
		Size:    stat.Size,
	}
	if meta := stat.Metadata["X-Amz-Meta-Filename"]; meta != nil {
		ret.Filename = meta[0]
	}

	return ret, nil
}

func (m *Minio) GetUserSpace(c context.Context, u string) (int64, error) {
	_, size, err := m.ListFiles(c, u)
	if err != nil {
		return 0, err
	}
	return size, nil
}

func (m *Minio) ListFiles(c context.Context, u string) ([]FileInfo, int64, error) {
	var files []FileInfo
	var size int64
	objects := m.MinioClient.ListObjects(c, m.MinioBucket, minio.ListObjectsOptions{
		Prefix:    u,
		Recursive: true,
	})

	for obj := range objects {
		if obj.Err != nil {
			continue
		}
		var file FileInfo
		stat, err := m.MinioClient.StatObject(c, m.MinioBucket, obj.Key, minio.StatObjectOptions{})
		if err != nil {
			continue
		}

		if filename := stat.Metadata["X-Amz-Meta-Filename"]; filename != nil {
			file.Filename = filename[0]
		}

		if key := strings.Split(obj.Key, "/"); len(key) > 1 {
			file.ID = key[1]
		}

		file.Content = obj.ContentType
		file.Size = obj.Size
		file.ExpiresAt = obj.LastModified.Add(7 * 24 * time.Hour)

		size += obj.Size
		files = append(files, file)
	}

	return files, size, nil
}

func (m *Minio) DeleteObject(c context.Context, key string) (*FileInfo, error) {
	err := m.MinioClient.RemoveObject(c, m.MinioBucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return nil, err
	}
	return nil, nil
}
