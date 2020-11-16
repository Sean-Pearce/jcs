package s3proxy

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/Sean-Pearce/jcs/service/httpserver/dao"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/klauspost/reedsolomon"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

func (p *Proxy) upload(bucket *dao.Bucket, key string) error {
	if bucket.Mode == "ec" {
		return p.uploadECMode(bucket, key)
	} else if bucket.Mode == "replica" {
		return p.uploadReplicaMode(bucket, key)
	}

	return nil
}

func (p *Proxy) uploadReplicaMode(bucket *dao.Bucket, key string) error {
	// download to disk
	dir := path.Join(p.tmpPath, uuid.NewV4().String())
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.WithError(err).Errorf("os.MkDirAll(%s, 0755) failed.", dir)
		return err
	}
	file, err := os.Create(path.Join(dir, key))
	if err != nil {
		log.WithError(err).Errorf("os.Create(%s) failed.", path.Join(dir, key))
		return err
	}
	defer file.Close()

	src := p.s3Map[minioName]
	downloader := s3manager.NewDownloaderWithClient(src)

	_, err = downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(bucket.Name),
		Key:    aws.String(key),
	})
	if err != nil {
		log.WithError(err).Errorf("Download object %s from bucket %s failed.", key, bucket.Name)
		return err
	}

	// upload to clouds
	for _, cloud := range bucket.Locations {
		dst := p.s3Map[cloud]
		_, err = dst.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(getBucketName(cloud, bucket.Name)),
			Key:    aws.String(key),
			Body:   file,
		})
		if err != nil {
			log.WithError(err).Errorf("Put object %s to bucket %s failed.", key, getBucketName(cloud, bucket.Name))
			continue
		}
	}

	return nil
}

func (p *Proxy) uploadECMode(bucket *dao.Bucket, key string) error {
	// download to disk
	dir := path.Join(p.tmpPath, uuid.NewV4().String())
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.WithError(err).Errorf("os.MkDirAll(%s, 0755) failed.", dir)
		return err
	}
	file, err := os.Create(path.Join(dir, key))
	if err != nil {
		log.WithError(err).Errorf("os.Create(%s) failed.", path.Join(dir, key))
		return err
	}
	defer file.Close()

	src := p.s3Map[minioName]
	downloader := s3manager.NewDownloaderWithClient(src)

	_, err = downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(bucket.Name),
		Key:    aws.String(key),
	})
	if err != nil {
		log.WithError(err).Errorf("Download object %s from bucket %s failed.", key, bucket.Name)
		return err
	}

	// Create encoding matrix.
	enc, err := reedsolomon.NewStream(bucket.N, bucket.K)
	if err != nil {
		log.WithError(err).Errorf("Create new encoder(n=%d,k=%d) failed.", bucket.N, bucket.K)
		return err
	}

	// Create the resulting files.
	out := make([]*os.File, bucket.N+bucket.K)
	for i := range out {
		outfn := fmt.Sprintf("%s.%d", key, i)
		out[i], err = os.Create(path.Join(dir, outfn))
		if err != nil {
			log.WithError(err).Errorf("os.Create(%s) failed.", path.Join(dir, outfn))
			return err
		}
	}

	// Split into files.
	instat, err := file.Stat()
	if err != nil {
		log.WithError(err).Errorf("file.Stat() failed: %s.", file.Name())
		return err
	}
	data := make([]io.Writer, bucket.N)
	for i := range data {
		data[i] = out[i]
	}
	err = enc.Split(file, data, instat.Size())
	if err != nil {
		log.WithError(err).Errorf("Split file %s(%sB) failed.", file.Name(), instat.Size())
		return err
	}

	// Close and re-open the files.
	input := make([]io.Reader, bucket.N)
	for i := range data {
		out[i].Close()
		f, err := os.Open(out[i].Name())
		if err != nil {
			log.WithError(err).Errorf("Open file %s failed.", out[i].Name())
			return err
		}
		input[i] = f
		out[i] = f
	}

	// Create parity output writers
	parity := make([]io.Writer, bucket.K)
	for i := range parity {
		parity[i] = out[bucket.N+i]
	}

	// Encode parity
	err = enc.Encode(input, parity)

	// upload to clouds
	for i, cloud := range bucket.Locations {
		out[i].Close()
		f, err := os.Open(out[i].Name())
		if err != nil {
			log.WithError(err).Errorf("Open file %s failed.", out[i].Name())
			return err
		}
		defer f.Close()

		dst := p.s3Map[cloud]
		_, err = dst.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(getBucketName(cloud, bucket.Name)),
			Key:    aws.String(key),
			Body:   f,
		})
		if err != nil {
			log.WithError(err).Errorf("Put object %s to bucket %s failed.", key, getBucketName(cloud, bucket.Name))
			return err
		}
	}

	return nil
}

func (p *Proxy) download(bucket *dao.Bucket, key string) error {
	if bucket.Mode == "ec" {
		return p.downloadECMode(bucket, key)
	} else if bucket.Mode == "replica" {
		return p.downloadReplicaMode(bucket, key)
	}

	return nil
}

func (p *Proxy) downloadReplicaMode(bucket *dao.Bucket, key string) error {
	// download to disk
	dir := path.Join(p.tmpPath, uuid.NewV4().String())
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	file, err := os.Create(path.Join(dir, key))
	if err != nil {
		return err
	}
	defer file.Close()

	// TODO: choose best cloud
	cloud := bucket.Locations[0]

	src := p.s3Map[cloud]
	downloader := s3manager.NewDownloaderWithClient(src)

	_, err = downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(getBucketName(cloud, bucket.Name)),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}

	// upload to minio
	dst := p.s3Map[minioName]
	_, err = dst.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket.Name),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *Proxy) downloadECMode(bucket *dao.Bucket, key string) error {
	// download to disk
	dir := path.Join(p.tmpPath, uuid.NewV4().String())
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	// TODO: choose best cloud
	clouds := bucket.Locations
	blocks := make([]*os.File, len(clouds))
	for i, cloud := range clouds {
		fname := fmt.Sprintf("%s.%d", key, i)
		file, err := os.Create(path.Join(dir, fname))
		if err != nil {
			return err
		}
		defer file.Close()

		blocks[i] = file
		src := p.s3Map[cloud]
		downloader := s3manager.NewDownloaderWithClient(src)

		_, err = downloader.Download(file, &s3.GetObjectInput{
			Bucket: aws.String(getBucketName(cloud, bucket.Name)),
			Key:    aws.String(key),
		})
		if err != nil {
			return err
		}
	}

	inputs := make([]io.Reader, len(clouds))
	for i := range inputs {
		inputs[i] = blocks[i]
	}

	file, err := os.Create(path.Join(dir, key))
	if err != nil {
		return err
	}

	enc, err := reedsolomon.NewStream(bucket.N, bucket.K)
	if err != nil {
		return err
	}

	// ok, err := enc.Verify(inputs)
	// logrus.WithError(err).Info(ok)

	err = enc.Join(file, inputs, 1024)
	if err != nil {
		log.WithError(err).Error("reconsruct failed")
		return err
	}

	// reopen file
	file.Close()
	file, err = os.Open(file.Name())
	if err != nil {
		return err
	}
	defer file.Close()

	// upload to minio
	dst := p.s3Map[minioName]
	_, err = dst.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket.Name),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return err
	}

	return nil
}
