package html

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

type ctxKey string

const assetsKey ctxKey = "assets"

//go:embed dist/*
var assetFs embed.FS

type asset struct {
	Path        string
	Name        string
	ContentType string
	Hash        string
}

type AssetMap map[string]asset

func (a AssetMap) Add(asset asset) {
	a[asset.Name] = asset
}

func MakeAssets() (AssetMap, error) {
	assets := make(AssetMap)

	walkFn := func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() || err != nil {
			return err
		}

		bytes, err := fs.ReadFile(assetFs, path)
		if err != nil {
			return fmt.Errorf("read %q asset: %w", path, err)
		}

		name := strings.TrimPrefix(path, "dist/")
		contentType := mime.TypeByExtension(filepath.Ext(path))

		hasher := sha256.New()
		if _, err := hasher.Write(bytes); err != nil {
			return fmt.Errorf("calculate asset %q hash: %w", path, err)
		}
		hash := hex.EncodeToString(hasher.Sum(nil))

		assets.Add(asset{
			Path:        path,
			Name:        name,
			ContentType: contentType,
			Hash:        hash[:8],
		})

		return nil
	}

	if err := fs.WalkDir(assetFs, ".", walkFn); err != nil {
		return nil, fmt.Errorf("load assets: %w", err)
	}

	return assets, nil
}

func NewAssetsMiddleware(logger *slog.Logger, assets AssetMap) func(http.Handler) http.Handler {
	const week = 7 * 24 * 60 * 60
	cacheControl := fmt.Sprintf("max-age=%d, public", week)

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			name := strings.TrimPrefix(r.URL.Path, "/")
			asset, ok := assets[name]
			if r.Method != http.MethodGet || !ok {
				r = r.WithContext(context.WithValue(r.Context(), assetsKey, assets))

				next.ServeHTTP(w, r)

				return
			}

			w.Header().Add("ETag", asset.Hash)
			w.Header().Add("Content-Type", asset.ContentType)
			w.Header().Add("Cache-Control", cacheControl)

			if r.Header.Get("If-None-Match") == asset.Hash {
				w.WriteHeader(http.StatusNotModified)

				return
			}

			f, err := assetFs.Open(asset.Path)
			if err != nil {
				logger.LogAttrs(r.Context(), slog.LevelError, "Failed to open asset",
					slog.String("asset", asset.Path),
					slog.String("error", err.Error()),
				)
				w.WriteHeader(http.StatusInternalServerError)

				return
			}
			defer f.Close()

			if _, err := io.Copy(w, f); err != nil {
				logger.LogAttrs(r.Context(), slog.LevelError, "Failed to deliver asset",
					slog.String("asset", asset.Path),
					slog.String("error", err.Error()),
				)
			}
		}

		return http.HandlerFunc(fn)
	}
}

var (
	errContextMissingAssetMap = errors.New("context doesn't contain AssetMap")
	errAssetNotFound          = errors.New("not found")
)

func assetPath(ctx context.Context, name string) (string, error) {
	assets, ok := ctx.Value(assetsKey).(AssetMap)
	if !ok {
		return "", errContextMissingAssetMap
	}

	asset, ok := assets[name]
	if !ok {
		return "", fmt.Errorf("asset %q: %w", name, errAssetNotFound)
	}

	return fmt.Sprintf("/%s?v=%s", asset.Name, asset.Hash), nil
}
