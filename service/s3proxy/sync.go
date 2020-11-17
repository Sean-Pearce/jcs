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

func (p *Proxy) download(bucket *dao.Bucket, key string) error {
	if bucket.Mode == "ec" {
		return p.downloadECMode(bucket, key)
	} else if bucket.Mode == "replica" {
		return p.downloadReplicaMode(bucket, key)
	}

	return nil
}

func (p *Proxy) uploadReplicaMode(bucket *dao.Bucket, key string) error {
	// download to disk
	dir := path.Join(p.tmpPath, uuid.NewV4().String())
	filename := path.Join(dir, key)
	err := p.download2Disk(minioName, bucket.Name, key, dir, filename)
	if err != nil {
		log.Errorf("Download to disk failed.")
		return err
	}

	// upload to clouds
	for _, cloud := range bucket.Locations {
		err := p.upload2Cloud(cloud, bucket.Name, key, filename)
		if err != nil {
			// TODO
			continue
		}
	}

	return nil
}

func (p *Proxy) uploadECMode(bucket *dao.Bucket, key string) error {
	// download to disk
	dir := path.Join(p.tmpPath, uuid.NewV4().String())
	filename := path.Join(dir, key)
	err := p.download2Disk(minioName, bucket.Name, key, dir, filename)
	if err != nil {
		log.Errorf("Download to disk failed.")
		return err
	}

	// encode file to shards
	shards := make([]string, bucket.N+bucket.K)
	for i := range shards {
		shards[i] = path.Join(dir, fmt.Sprintf("%s.%d", key, i))
	}
	err = encode(filename, shards, bucket.N, bucket.K)
	if err != nil {
		log.Errorf("Encode file %s failed.", filename)
		return err
	}

	// upload to clouds
	for i, cloud := range bucket.Locations {
		err := p.upload2Cloud(cloud, bucket.Name, key, shards[i])
		if err != nil {
			log.Errorf("Upload to cloud %s failed.", cloud)
			// TODO
			continue
		}
	}

	return nil
}

func (p *Proxy) downloadReplicaMode(bucket *dao.Bucket, key string) error {
	// download to disk
	dir := path.Join(p.tmpPath, uuid.NewV4().String())
	filename := path.Join(dir, key)
	// TODO
	cloud := bucket.Locations[0]
	err := p.download2Disk(cloud, bucket.Name, key, dir, filename)
	if err != nil {
		log.Errorf("Download to disk failed.")
		return err
	}

	// upload to minio
	err = p.upload2Cloud(minioName, bucket.Name, key, filename)
	if err != nil {
		log.Errorf("Upload to %s failed.", minioName)
		return err
	}

	return nil
}

func (p *Proxy) downloadECMode(bucket *dao.Bucket, key string) error {
	// download to disk
	dir := path.Join(p.tmpPath, uuid.NewV4().String())
	filename := path.Join(dir, key)
	// TODO
	clouds := bucket.Locations
	shards := make([]string, len(clouds))
	for i, cloud := range clouds {
		shards[i] = path.Join(dir, fmt.Sprintf("%s.%d", key, i))
		err := p.download2Disk(cloud, bucket.Name, key, dir, shards[i])
		if err != nil {
			log.Errorf("Download from %s failed.", cloud)
			return err
		}
	}

	// decode to disk
	err := decode(filename, shards, bucket.N, bucket.K)
	if err != nil {
		log.Debugf("Decode %s failed.", filename)
		return err
	}

	// upload to minio
	err = p.upload2Cloud(minioName, bucket.Name, key, filename)
	if err != nil {
		log.Debugf("Upload to %s failed.", minioName)
		return err
	}

	return nil
}

func (p *Proxy) download2Disk(cloud string, bucket string, key string, dir string, filename string) error {
	log.Debugf("cloud: %s, bucket: %s, key: %s, dir: %s, filename: %s", cloud, bucket, key, dir, filename)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.WithError(err).Errorf("os.MkDirAll(%s, 0755) failed.", dir)
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		log.WithError(err).Errorf("os.Create(%s) failed.", filename)
		return err
	}
	defer file.Close()

	src := p.s3Map[cloud]
	downloader := s3manager.NewDownloaderWithClient(src)

	_, err = downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(getBucketName(cloud, bucket)),
		Key:    aws.String(key),
	})
	if err != nil {
		log.WithError(err).Errorf("Download object %s from bucket %s failed.", key, bucket)
		return err
	}

	return nil
}

func (p *Proxy) upload2Cloud(cloud string, bucket string, key string, filename string) error {
	log.Debugf("cloud: %s, bucket: %s, key: %s, filename: %s", cloud, bucket, key, filename)
	// open file
	file, err := os.Open(filename)
	if err != nil {
		log.WithError(err).Errorf("Open %s failed.", filename)
		return err
	}
	defer file.Close()

	// upload to cloud
	dst := p.s3Map[cloud]
	_, err = dst.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(getBucketName(cloud, bucket)),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		log.WithError(err).Errorf("Put object %s to bucket %s failed.", key, getBucketName(cloud, bucket))
		return err
	}

	return nil
}

func encode(filename string, shards []string, n, k int) error {
	log.Debugf("filename: %s, shards: %v, n: %d, k: %d", filename, shards, n, k)

	// open file
	file, err := os.Open(filename)
	if err != nil {
		log.WithError(err).Errorf("Open %s failed.", filename)
		return err
	}
	defer file.Close()

	// Create encoding matrix
	enc, err := reedsolomon.NewStream(n, k)
	if err != nil {
		log.WithError(err).Errorf("Create new encoder(n=%d,k=%d) failed.", n, k)
		return err
	}

	// Create the resulting files
	out := make([]*os.File, n+k)
	for i := range out {
		out[i], err = os.Create(shards[i])
		if err != nil {
			log.WithError(err).Errorf("os.Create(%s) failed.", shards[i])
			return err
		}
	}

	// Split into files.
	instat, err := file.Stat()
	if err != nil {
		log.WithError(err).Errorf("file.Stat() failed: %s.", file.Name())
		return err
	}
	data := make([]io.Writer, n)
	for i := range data {
		data[i] = out[i]
	}
	err = enc.Split(file, data, instat.Size())
	if err != nil {
		log.WithError(err).Errorf("Split file %s(%sB) failed.", file.Name(), instat.Size())
		return err
	}

	// Close and re-open the files.
	input := make([]io.Reader, n)
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
	parity := make([]io.Writer, k)
	for i := range parity {
		parity[i] = out[n+i]
	}

	// Encode parity
	err = enc.Encode(input, parity)
	if err != nil {
		log.WithError(err).Errorf("Encode parity shards failed.")
	}

	// Close result files
	for _, f := range out {
		f.Close()
	}

	return nil
}

func decode(filename string, shards []string, n, k int) error {
	log.Debugf("filename: %s, shards: %v, n: %d, k: %d", filename, shards, n, k)

	// read shards
	inputs := make([]io.Reader, n+k)
	for i, s := range shards {
		f, err := os.Open(s)
		if err != nil {
			log.WithError(err).Errorf("Open file %s failed.", s)
			return err
		}
		inputs[i] = f
		defer f.Close()
	}

	// create file
	file, err := os.Create(filename)
	if err != nil {
		log.WithError(err).Errorf("Create file %s failed.", filename)
		return err
	}
	defer file.Close()

	enc, err := reedsolomon.NewStream(n, k)
	if err != nil {
		log.WithError(err).Errorf("Create new encoder failed.")
		return err
	}

	// ok, err := enc.Verify(inputs)
	// logrus.WithError(err).Info(ok)

	err = enc.Join(file, inputs, 1024)
	if err != nil {
		log.WithError(err).Error("reconsruct failed")
		return err
	}

	return nil
}
