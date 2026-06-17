package s3

import (
	"bytes"
	"context"
	"errors"
	"testing"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

type fakeS3API struct {
	putObjectFn    func(ctx context.Context, params *awss3.PutObjectInput, optFns ...func(*awss3.Options)) (*awss3.PutObjectOutput, error)
	deleteObjectFn func(ctx context.Context, params *awss3.DeleteObjectInput, optFns ...func(*awss3.Options)) (*awss3.DeleteObjectOutput, error)
}

func (f *fakeS3API) PutObject(ctx context.Context, params *awss3.PutObjectInput, optFns ...func(*awss3.Options)) (*awss3.PutObjectOutput, error) {
	return f.putObjectFn(ctx, params, optFns...)
}

func (f *fakeS3API) DeleteObject(ctx context.Context, params *awss3.DeleteObjectInput, optFns ...func(*awss3.Options)) (*awss3.DeleteObjectOutput, error) {
	return f.deleteObjectFn(ctx, params, optFns...)
}

func TestAdapter_Upload_ReturnsURLAndPublicID(t *testing.T) {
	var gotBucket, gotKey string
	fake := &fakeS3API{
		putObjectFn: func(_ context.Context, params *awss3.PutObjectInput, _ ...func(*awss3.Options)) (*awss3.PutObjectOutput, error) {
			gotBucket = *params.Bucket
			gotKey = *params.Key
			return &awss3.PutObjectOutput{}, nil
		},
	}
	a := &adapter{client: fake, bucket: "my-bucket", region: "us-east-1"}

	result, err := a.Upload(context.Background(), bytes.NewReader([]byte("data")), "photo.jpg", false)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if gotBucket != "my-bucket" || gotKey != "photo.jpg" {
		t.Errorf("Upload() called PutObject with bucket=%q key=%q, want my-bucket/photo.jpg", gotBucket, gotKey)
	}
	if result.PublicID != "photo.jpg" {
		t.Errorf("Upload() PublicID = %q, want photo.jpg", result.PublicID)
	}
	wantURL := "https://my-bucket.s3.us-east-1.amazonaws.com/photo.jpg"
	if result.URL != wantURL {
		t.Errorf("Upload() URL = %q, want %q", result.URL, wantURL)
	}
}

func TestAdapter_Upload_AlwaysSetsThumbnailURLEqualToURL(t *testing.T) {
	fake := &fakeS3API{
		putObjectFn: func(_ context.Context, _ *awss3.PutObjectInput, _ ...func(*awss3.Options)) (*awss3.PutObjectOutput, error) {
			return &awss3.PutObjectOutput{}, nil
		},
	}
	a := &adapter{client: fake, bucket: "my-bucket", region: "us-east-1"}

	for _, generateThumbnail := range []bool{true, false} {
		result, err := a.Upload(context.Background(), bytes.NewReader([]byte("data")), "photo.jpg", generateThumbnail)
		if err != nil {
			t.Fatalf("Upload(generateThumbnail=%v) error = %v", generateThumbnail, err)
		}
		if result.ThumbnailURL != result.URL {
			t.Errorf("Upload(generateThumbnail=%v) ThumbnailURL=%q, want same as URL=%q", generateThumbnail, result.ThumbnailURL, result.URL)
		}
	}
}

func TestAdapter_Upload_PropagatesError(t *testing.T) {
	fake := &fakeS3API{
		putObjectFn: func(_ context.Context, _ *awss3.PutObjectInput, _ ...func(*awss3.Options)) (*awss3.PutObjectOutput, error) {
			return nil, errors.New("boom")
		},
	}
	a := &adapter{client: fake, bucket: "b", region: "r"}

	if _, err := a.Upload(context.Background(), bytes.NewReader(nil), "f.jpg", false); err == nil {
		t.Error("Upload() error = nil, want error from PutObject")
	}
}

func TestAdapter_Delete_RemovesObject(t *testing.T) {
	var gotKey string
	fake := &fakeS3API{
		deleteObjectFn: func(_ context.Context, params *awss3.DeleteObjectInput, _ ...func(*awss3.Options)) (*awss3.DeleteObjectOutput, error) {
			gotKey = *params.Key
			return &awss3.DeleteObjectOutput{}, nil
		},
	}
	a := &adapter{client: fake, bucket: "my-bucket", region: "us-east-1"}

	if err := a.Delete(context.Background(), "photo.jpg"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if gotKey != "photo.jpg" {
		t.Errorf("Delete() called DeleteObject with key=%q, want photo.jpg", gotKey)
	}
}

func TestAdapter_Delete_PropagatesError(t *testing.T) {
	fake := &fakeS3API{
		deleteObjectFn: func(_ context.Context, _ *awss3.DeleteObjectInput, _ ...func(*awss3.Options)) (*awss3.DeleteObjectOutput, error) {
			return nil, errors.New("boom")
		},
	}
	a := &adapter{client: fake, bucket: "b", region: "r"}

	if err := a.Delete(context.Background(), "f.jpg"); err == nil {
		t.Error("Delete() error = nil, want error from DeleteObject")
	}
}
