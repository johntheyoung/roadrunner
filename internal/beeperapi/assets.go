package beeperapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	beeperdesktopapi "github.com/beeper/desktop-api-go"
	"github.com/beeper/desktop-api-go/option"
)

// AssetsService handles asset operations.
type AssetsService struct {
	client *Client
}

// AssetUploadParams configures multipart upload requests.
type AssetUploadParams struct {
	FilePath string
	FileName string
	MimeType string
}

// AssetUploadBase64Params configures base64 upload requests.
type AssetUploadBase64Params struct {
	Content  string
	FileName string
	MimeType string
}

// AssetUploadResult represents upload metadata.
type AssetUploadResult struct {
	UploadID string  `json:"upload_id"`
	SrcURL   string  `json:"src_url,omitempty"`
	FileName string  `json:"file_name,omitempty"`
	MimeType string  `json:"mime_type,omitempty"`
	FileSize int64   `json:"file_size,omitempty"`
	Duration float64 `json:"duration,omitempty"`
	Width    int     `json:"width,omitempty"`
	Height   int     `json:"height,omitempty"`
}

// AssetServeResult represents streamed response metadata.
type AssetServeResult struct {
	ContentType   string `json:"content_type,omitempty"`
	ContentLength int64  `json:"content_length,omitempty"`
	BytesWritten  int64  `json:"bytes_written"`
}

// Download retrieves a local file URL for an asset.
func (s *AssetsService) Download(ctx context.Context, url string) (string, error) {
	ctx, cancel := s.client.contextWithTimeout(ctx)
	defer cancel()

	resp, err := s.client.SDK.Assets.Download(ctx, beeperdesktopapi.AssetDownloadParams{
		URL: url,
	})
	if err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", fmt.Errorf("download failed: %s", resp.Error)
	}
	if resp.SrcURL == "" {
		return "", fmt.Errorf("empty src_url in response")
	}
	return resp.SrcURL, nil
}

// Serve streams an asset response to the provided writer.
func (s *AssetsService) Serve(ctx context.Context, url string, dst io.Writer) (AssetServeResult, error) {
	if dst == nil {
		return AssetServeResult{}, fmt.Errorf("serve failed: destination writer is nil")
	}

	ctx, cancel := s.client.contextWithTimeout(ctx)
	defer cancel()

	var resp *http.Response
	if err := s.client.SDK.Get(
		ctx,
		"v1/assets/serve",
		beeperdesktopapi.AssetServeParams{URL: url},
		nil,
		option.WithHeader("Accept", "*/*"),
		option.WithResponseInto(&resp),
	); err != nil {
		return AssetServeResult{}, err
	}
	if resp == nil || resp.Body == nil {
		return AssetServeResult{}, fmt.Errorf("serve failed: empty response body")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	n, err := io.Copy(dst, resp.Body)
	if err != nil {
		return AssetServeResult{}, fmt.Errorf("stream asset: %w", err)
	}

	return AssetServeResult{
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: resp.ContentLength,
		BytesWritten:  n,
	}, nil
}

// Upload stores a local file in temporary upload storage and returns an upload ID.
func (s *AssetsService) Upload(ctx context.Context, params AssetUploadParams) (AssetUploadResult, error) {
	ctx, cancel := s.client.contextWithTimeout(ctx)
	defer cancel()

	f, err := os.Open(params.FilePath)
	if err != nil {
		return AssetUploadResult{}, fmt.Errorf("open %s: %w", params.FilePath, err)
	}
	defer func() {
		_ = f.Close()
	}()

	sdkParams := beeperdesktopapi.AssetUploadParams{
		File: f,
	}
	if params.FileName != "" {
		sdkParams.FileName = beeperdesktopapi.String(params.FileName)
	}
	if params.MimeType != "" {
		sdkParams.MimeType = beeperdesktopapi.String(params.MimeType)
	}

	resp, err := s.client.SDK.Assets.Upload(ctx, sdkParams)
	if err != nil {
		return AssetUploadResult{}, err
	}
	if resp.Error != "" {
		return AssetUploadResult{}, fmt.Errorf("upload failed: %s", resp.Error)
	}
	if resp.UploadID == "" {
		return AssetUploadResult{}, fmt.Errorf("upload failed: empty upload_id in response")
	}

	return AssetUploadResult{
		UploadID: resp.UploadID,
		SrcURL:   resp.SrcURL,
		FileName: resp.FileName,
		MimeType: resp.MimeType,
		FileSize: int64(resp.FileSize),
		Duration: resp.Duration,
		Width:    int(resp.Width),
		Height:   int(resp.Height),
	}, nil
}

// UploadBase64 uploads a file payload encoded as base64.
func (s *AssetsService) UploadBase64(ctx context.Context, params AssetUploadBase64Params) (AssetUploadResult, error) {
	ctx, cancel := s.client.contextWithTimeout(ctx)
	defer cancel()

	sdkParams := beeperdesktopapi.AssetUploadBase64Params{
		Content: params.Content,
	}
	if params.FileName != "" {
		sdkParams.FileName = beeperdesktopapi.String(params.FileName)
	}
	if params.MimeType != "" {
		sdkParams.MimeType = beeperdesktopapi.String(params.MimeType)
	}

	resp, err := s.client.SDK.Assets.UploadBase64(ctx, sdkParams)
	if err != nil {
		return AssetUploadResult{}, err
	}
	if resp.Error != "" {
		return AssetUploadResult{}, fmt.Errorf("upload failed: %s", resp.Error)
	}
	if resp.UploadID == "" {
		return AssetUploadResult{}, fmt.Errorf("upload failed: empty upload_id in response")
	}

	return AssetUploadResult{
		UploadID: resp.UploadID,
		SrcURL:   resp.SrcURL,
		FileName: resp.FileName,
		MimeType: resp.MimeType,
		FileSize: int64(resp.FileSize),
		Duration: resp.Duration,
		Width:    int(resp.Width),
		Height:   int(resp.Height),
	}, nil
}
