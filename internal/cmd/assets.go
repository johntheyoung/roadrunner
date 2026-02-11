package cmd

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/johntheyoung/roadrunner/internal/beeperapi"
	"github.com/johntheyoung/roadrunner/internal/config"
	"github.com/johntheyoung/roadrunner/internal/errfmt"
	"github.com/johntheyoung/roadrunner/internal/outfmt"
	"github.com/johntheyoung/roadrunner/internal/ui"
)

// AssetsCmd is the parent command for asset subcommands.
type AssetsCmd struct {
	Download     AssetsDownloadCmd     `cmd:"" help:"Download an asset by mxc:// URL"`
	Upload       AssetsUploadCmd       `cmd:"" help:"Upload an asset and return upload ID"`
	UploadBase64 AssetsUploadBase64Cmd `cmd:"" name:"upload-base64" help:"Upload base64 data and return upload ID"`
}

// AssetsDownloadCmd downloads an asset from a Matrix URL.
type AssetsDownloadCmd struct {
	URL  string `arg:"" name:"url" help:"Matrix content URL (mxc:// or localmxc://)"`
	Dest string `help:"Destination file or directory (optional)" name:"dest"`
}

// AssetsUploadCmd uploads a local file.
type AssetsUploadCmd struct {
	FilePath string `arg:"" name:"file" help:"File path to upload"`
	FileName string `help:"Filename to send in metadata (optional)" name:"file-name"`
	MimeType string `help:"MIME type override (optional)" name:"mime-type"`
}

// AssetsUploadBase64Cmd uploads base64 content.
type AssetsUploadBase64Cmd struct {
	Content     string `arg:"" optional:"" help:"Base64-encoded content"`
	ContentFile string `help:"Read base64 content from file ('-' for stdin)" name:"content-file"`
	Stdin       bool   `help:"Read base64 content from stdin" name:"stdin"`
	FileName    string `help:"Filename to send in metadata (optional)" name:"file-name"`
	MimeType    string `help:"MIME type override (optional)" name:"mime-type"`
}

// Run executes the assets download command.
func (c *AssetsDownloadCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	if c.URL == "" {
		return errfmt.UsageError("asset URL is required")
	}

	token, _, err := config.GetToken()
	if err != nil {
		return err
	}

	timeout := time.Duration(flags.Timeout) * time.Second
	client, err := beeperapi.NewClient(token, flags.BaseURL, timeout)
	if err != nil {
		return err
	}

	srcURL, err := client.Assets().Download(ctx, c.URL)
	if err != nil {
		return err
	}

	destPath := ""
	if c.Dest != "" {
		destPath, err = assetsDownloadDestination(srcURL, c.Dest)
		if err != nil {
			return errfmt.UsageError("invalid --dest: %s", err.Error())
		}
		if err := copyAsset(srcURL, destPath); err != nil {
			return err
		}
	}

	// JSON output
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, map[string]any{
			"src_url": srcURL,
			"dest":    destPath,
			"copied":  destPath != "",
		}, "assets download")
	}

	// Plain output
	if outfmt.IsPlain(ctx) {
		if destPath != "" {
			u.Out().Printf("%s", destPath)
		} else {
			u.Out().Printf("%s", srcURL)
		}
		return nil
	}

	// Human-readable output
	if destPath != "" {
		u.Out().Successf("Saved to %s", destPath)
	} else {
		u.Out().Successf("Downloaded to %s", srcURL)
	}

	return nil
}

// Run executes the assets upload command.
func (c *AssetsUploadCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	if strings.TrimSpace(c.FilePath) == "" {
		return errfmt.UsageError("file path is required")
	}

	token, _, err := config.GetToken()
	if err != nil {
		return err
	}

	timeout := time.Duration(flags.Timeout) * time.Second
	client, err := beeperapi.NewClient(token, flags.BaseURL, timeout)
	if err != nil {
		return err
	}

	resp, err := client.Assets().Upload(ctx, beeperapi.AssetUploadParams{
		FilePath: c.FilePath,
		FileName: c.FileName,
		MimeType: c.MimeType,
	})
	if err != nil {
		if beeperapi.IsUnsupportedRoute(err, "POST", "/assets/upload") {
			return fmt.Errorf("asset upload is not supported by this Beeper Desktop API version (requires a newer Beeper Desktop build)")
		}
		return err
	}

	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, resp, "assets upload")
	}

	if outfmt.IsPlain(ctx) {
		u.Out().Printf("%s\t%s\t%s\t%s\t%d", resp.UploadID, resp.FileName, resp.MimeType, resp.SrcURL, resp.FileSize)
		return nil
	}

	u.Out().Success("Asset uploaded")
	u.Out().Printf("Upload ID: %s", resp.UploadID)
	if resp.FileName != "" {
		u.Out().Printf("File:      %s", resp.FileName)
	}
	if resp.MimeType != "" {
		u.Out().Printf("MIME:      %s", resp.MimeType)
	}

	return nil
}

// Run executes the assets upload-base64 command.
func (c *AssetsUploadBase64Cmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	content, err := resolveTextInput(c.Content, c.ContentFile, c.Stdin, true, "base64 content", "--content-file", "--stdin")
	if err != nil {
		return err
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return errfmt.UsageError("base64 content is required")
	}

	token, _, err := config.GetToken()
	if err != nil {
		return err
	}

	timeout := time.Duration(flags.Timeout) * time.Second
	client, err := beeperapi.NewClient(token, flags.BaseURL, timeout)
	if err != nil {
		return err
	}

	resp, err := client.Assets().UploadBase64(ctx, beeperapi.AssetUploadBase64Params{
		Content:  content,
		FileName: c.FileName,
		MimeType: c.MimeType,
	})
	if err != nil {
		if beeperapi.IsUnsupportedRoute(err, "POST", "/assets/upload/base64") {
			return fmt.Errorf("base64 asset upload is not supported by this Beeper Desktop API version (requires a newer Beeper Desktop build)")
		}
		return err
	}

	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, resp, "assets upload-base64")
	}

	if outfmt.IsPlain(ctx) {
		u.Out().Printf("%s\t%s\t%s\t%s\t%d", resp.UploadID, resp.FileName, resp.MimeType, resp.SrcURL, resp.FileSize)
		return nil
	}

	u.Out().Success("Asset uploaded")
	u.Out().Printf("Upload ID: %s", resp.UploadID)
	if resp.FileName != "" {
		u.Out().Printf("File:      %s", resp.FileName)
	}
	if resp.MimeType != "" {
		u.Out().Printf("MIME:      %s", resp.MimeType)
	}

	return nil
}

func assetsDownloadDestination(srcURL string, dest string) (string, error) {
	path, err := assetLocalPath(srcURL)
	if err != nil {
		return "", err
	}
	base := filepath.Base(path)
	if base == "." || base == string(filepath.Separator) {
		return "", fmt.Errorf("invalid source filename")
	}
	info, err := os.Stat(dest)
	if err == nil && info.IsDir() {
		return filepath.Join(dest, base), nil
	}
	if err != nil && os.IsNotExist(err) {
		if strings.HasSuffix(dest, string(filepath.Separator)) {
			return filepath.Join(dest, base), nil
		}
	}
	return dest, nil
}

func assetLocalPath(srcURL string) (string, error) {
	if strings.HasPrefix(srcURL, "file://") {
		parsed, err := url.Parse(srcURL)
		if err != nil {
			return "", err
		}
		if parsed.Path == "" {
			return "", fmt.Errorf("empty file URL path")
		}
		path, err := url.PathUnescape(parsed.Path)
		if err != nil {
			return "", err
		}
		return filepath.FromSlash(path), nil
	}
	return srcURL, nil
}

func copyAsset(srcURL string, dest string) error {
	path, err := assetLocalPath(srcURL)
	if err != nil {
		return err
	}
	in, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer func() {
		_ = in.Close()
	}()

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create %s: %w", dest, err)
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy asset: %w", err)
	}

	return nil
}
