package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	Serve        AssetsServeCmd        `cmd:"" help:"Stream an asset by URL (raw bytes)"`
	Upload       AssetsUploadCmd       `cmd:"" help:"Upload an asset and return upload ID"`
	UploadBase64 AssetsUploadBase64Cmd `cmd:"" name:"upload-base64" help:"Upload base64 data and return upload ID"`
}

// AssetsDownloadCmd downloads an asset from a Matrix URL.
type AssetsDownloadCmd struct {
	URL  string `arg:"" name:"url" help:"Matrix content URL (mxc:// or localmxc://)"`
	Dest string `help:"Destination file or directory (optional)" name:"dest"`
}

// AssetsServeCmd streams an asset by URL.
type AssetsServeCmd struct {
	URL    string `arg:"" name:"url" help:"Asset URL to stream (mxc://, localmxc://, or file://)"`
	Dest   string `help:"Destination file path (writes raw bytes to file)" name:"dest"`
	Stdout bool   `help:"Force writing raw bytes to stdout (even on a terminal)" name:"stdout"`
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

// Run executes the assets serve command.
func (c *AssetsServeCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	if c.URL == "" {
		return errfmt.UsageError("asset URL is required")
	}
	if c.Dest != "" && c.Stdout {
		return errfmt.UsageError("cannot use --dest with --stdout")
	}
	if outfmt.IsJSON(ctx) || outfmt.IsPlain(ctx) {
		if c.Dest == "" {
			return errfmt.UsageError("--dest is required with --json or --plain for assets serve")
		}
		if c.Stdout {
			return errfmt.UsageError("--stdout cannot be used with --json or --plain")
		}
	}

	writeToStdout := c.Stdout || c.Dest == ""
	if writeToStdout && !c.Stdout && stdoutIsTerminal() {
		return errfmt.UsageError("refusing to write binary data to terminal stdout; use --dest or --stdout")
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

	var writer io.Writer
	var outFile *os.File
	if writeToStdout {
		writer = os.Stdout
	} else {
		if err := os.MkdirAll(filepath.Dir(c.Dest), 0755); err != nil {
			return fmt.Errorf("create dir: %w", err)
		}
		f, err := os.Create(c.Dest)
		if err != nil {
			return fmt.Errorf("create %s: %w", c.Dest, err)
		}
		outFile = f
		writer = f
	}

	serve, err := client.Assets().Serve(ctx, c.URL, writer)
	if outFile != nil {
		_ = outFile.Close()
	}
	if err != nil {
		return err
	}

	if writeToStdout {
		return nil
	}

	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, map[string]any{
			"url":            c.URL,
			"dest":           c.Dest,
			"content_type":   serve.ContentType,
			"content_length": serve.ContentLength,
			"bytes_written":  serve.BytesWritten,
		}, "assets serve")
	}

	if outfmt.IsPlain(ctx) {
		u.Out().Printf("%s\t%d\t%s", c.Dest, serve.BytesWritten, serve.ContentType)
		return nil
	}

	u.Out().Successf("Streamed to %s", c.Dest)
	u.Out().Printf("Bytes: %d", serve.BytesWritten)
	if serve.ContentType != "" {
		u.Out().Printf("Content-Type: %s", serve.ContentType)
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
	if err := checkAndRememberNonIdempotentDuplicate(ctx, flags, "assets upload", struct {
		FilePath string `json:"file_path"`
		FileName string `json:"file_name"`
		MimeType string `json:"mime_type"`
	}{
		FilePath: c.FilePath,
		FileName: c.FileName,
		MimeType: c.MimeType,
	}); err != nil {
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
	contentSum := sha256.Sum256([]byte(content))
	if err := checkAndRememberNonIdempotentDuplicate(ctx, flags, "assets upload-base64", struct {
		ContentSHA256 string `json:"content_sha256"`
		FileName      string `json:"file_name"`
		MimeType      string `json:"mime_type"`
	}{
		ContentSHA256: hex.EncodeToString(contentSum[:]),
		FileName:      c.FileName,
		MimeType:      c.MimeType,
	}); err != nil {
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

func stdoutIsTerminal() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
