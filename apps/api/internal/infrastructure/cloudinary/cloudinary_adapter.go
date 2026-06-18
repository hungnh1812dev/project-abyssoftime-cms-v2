package cloudinary

import (
	"context"
	"fmt"
	"io"

	cloudinarygo "github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"

	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

var _ repository.StorageAdapter = (*adapter)(nil)

type adapter struct {
	cld *cloudinarygo.Cloudinary
}

func NewCloudinaryAdapter(cloudName, apiKey, apiSecret string) (repository.StorageAdapter, error) {
	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("cloudinary: cloudName, apiKey, and apiSecret are all required")
	}
	cld, err := cloudinarygo.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, err
	}
	return &adapter{cld: cld}, nil
}

func boolPtr(b bool) *bool { return &b }

func (a *adapter) Upload(ctx context.Context, file io.Reader, filename string, generateThumbnail bool) (*repository.UploadResult, error) {
	params := uploader.UploadParams{
		PublicID:       filename,
		Overwrite:      boolPtr(true),
		UniqueFilename: boolPtr(false),
	}
	if generateThumbnail {
		params.Eager = "c_thumb,w_300,h_300"
		params.EagerAsync = boolPtr(false)
	}
	res, err := a.cld.Upload.Upload(ctx, file, params)
	if err != nil {
		return nil, err
	}
	thumbnailURL := res.SecureURL
	if generateThumbnail && len(res.Eager) > 0 {
		thumbnailURL = res.Eager[0].SecureURL
	}
	return &repository.UploadResult{
		URL:          res.SecureURL,
		ThumbnailURL: thumbnailURL,
		PublicID:     res.PublicID,
	}, nil
}

func (a *adapter) Delete(ctx context.Context, publicID string) error {
	_, err := a.cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
	return err
}
