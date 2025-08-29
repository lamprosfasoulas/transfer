package storage

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Minio struct {
	//MinioEndpoint	string
	//MinioAccessKey	string
	//MinioSecretKey	string
	//MinioUseSSL		bool
	MinioBucket		string //Bucket name
	MinioClient 	*minio.Client //Minio client
	Error 			error //Error
}

func (m *Minio) GetError () error {
	return m.Error
}

func NewMinio(endpoint, access, passkey, bucket string, ssl bool) *Minio {
	cl, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(access, passkey, ""),
		Secure: ssl,
	}) 
	if err != nil {
		return &Minio{
			MinioBucket: bucket,
			MinioClient: cl,
			Error: err,
		}
	}
	_, err = cl.BucketExists(context.Background(), bucket)
	if err != nil {
		return &Minio{
			MinioBucket: bucket,
			MinioClient: cl,
			Error: err,
		}
	}
	return &Minio{
		//MinioEndpoint: endpoint,
		//MinioAccessKey: access,
		//MinioSecretKey: passkey,
		//MinioUseSSL: ssl,
		MinioBucket: bucket,
		MinioClient: cl,
		Error: err,
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
			Key: uploadInfo.Key,
			Filename: r.Filename,
			Size: uploadInfo.Size,
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
		Object: 	object,
		Content: 	stat.ContentType,
		Size: 		stat.Size,
	}
	if meta := stat.Metadata["X-Amz-Meta-Filename"]; meta != nil {
		ret.Filename = meta[0]
	}

	return ret, nil
}
//i need work
// not in use
//func (m *Minio) ListObjects(c context.Context, prefix string) *[]FileInfo{
//	objCh := m.MinioClient.ListObjects(c, m.MinioBucket, minio.ListObjectsOptions{
//		Prefix: prefix,
//		Recursive: true,
//	})
//	var files []FileInfo
//	for o := range objCh {
//		var file FileInfo
//		if o.Err != nil {
//			log.Printf("Error listing objects: %v\n", o.Err)
//			continue
//		}
//		parts := strings.SplitN(o.Key, "/", 2)
//		if len(parts) != 2 {
//			continue
//		}
//
//		file.ID = parts[1]
//		statInfo, err := m.MinioClient.StatObject(c, m.MinioBucket, o.Key, minio.StatObjectOptions{})
//		if err != nil {
//			return nil, err
//		}
//		meta := statInfo.Metadata["X-Amz-Meta-Filename"]
//		if meta != nil {
//			file.Filename = meta[0]
//		}
//		file.ExpireAt  	= int(time.Duration(o.LastModified.Add(7 * 24 * time.Hour).Sub(time.Now()).Hours()))
//		file.Key 		= o.Key
//		file.LastMod 	= o.LastModified
//		file.DispSize 	= bites(o.Size)
//		file.Size 		= o.Size
//		//This could go in the template insted
//		file.URL 		= "/download/" + prefix + "/" + file.ID
//		files 			= append(files, file)
//	}
//	return &files
//}

func (m *Minio) DeleteObject(c context.Context, key string) (*FileInfo, error){
	err := m.MinioClient.RemoveObject(c, m.MinioBucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return nil, err
	}
	return nil, nil
}
