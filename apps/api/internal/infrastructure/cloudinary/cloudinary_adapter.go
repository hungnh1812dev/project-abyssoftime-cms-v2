package cloudinary

import (
	"context"
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
	cld, err := cloudinarygo.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, err
	}
	return &adapter{cld: cld}, nil
}

func boolPtr(b bool) *bool { return &b }

func (a *adapter) Upload(ctx context.Context, file io.Reader, filename string) (*repository.UploadResult, error) {
	res, err := a.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID:       filename,
		Overwrite:      boolPtr(true),
		UniqueFilename: boolPtr(false),
	})
	if err != nil {
		return nil, err
	}
	return &repository.UploadResult{
		URL:      res.SecureURL,
		PublicID: res.PublicID,
	}, nil
}

func (a *adapter) Delete(ctx context.Context, publicID string) error {
	_, err := a.cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
	return err
}
