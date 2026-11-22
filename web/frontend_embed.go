package web

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"yatori-go-console/frontend"
)

const (
	embeddedFrontendDir  = "assets/frontend"
	embeddedFrontendHash = "assets/frontend.hash"
)

// prepareEmbeddedFrontend extracts the frontend bundled into the binary to
// assets/frontend. A content hash is stored next to the directory so an
// upgraded binary replaces stale files while an unchanged binary skips the
// extraction on later starts.
func prepareEmbeddedFrontend() error {
	if !frontend.Available() {
		return nil
	}
	src := frontend.FS()
	hash, err := hashEmbeddedFS(src)
	if err != nil {
		return fmt.Errorf("计算内嵌前端哈希失败: %w", err)
	}
	if current, readErr := os.ReadFile(embeddedFrontendHash); readErr == nil &&
		strings.TrimSpace(string(current)) == hash {
		if info, statErr := os.Stat(filepath.Join(embeddedFrontendDir, "index.html")); statErr == nil && !info.IsDir() {
			return nil
		}
	}

	tmpDir := embeddedFrontendDir + ".tmp"
	_ = os.RemoveAll(tmpDir)
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("创建前端临时目录失败: %w", err)
	}
	walkErr := fs.WalkDir(src, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(tmpDir, filepath.FromSlash(name))
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, readErr := fs.ReadFile(src, name)
		if readErr != nil {
			return readErr
		}
		if mkErr := os.MkdirAll(filepath.Dir(target), 0o755); mkErr != nil {
			return mkErr
		}
		return os.WriteFile(target, data, 0o644)
	})
	if walkErr != nil {
		_ = os.RemoveAll(tmpDir)
		return fmt.Errorf("释放内嵌前端失败: %w", walkErr)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "index.html")); err != nil {
		_ = os.RemoveAll(tmpDir)
		return fmt.Errorf("内嵌前端缺少 index.html")
	}
	_ = os.RemoveAll(embeddedFrontendDir)
	if err := os.Rename(tmpDir, embeddedFrontendDir); err != nil {
		_ = os.RemoveAll(tmpDir)
		return fmt.Errorf("替换前端目录失败: %w", err)
	}
	if err := os.WriteFile(embeddedFrontendHash, []byte(hash), 0o644); err != nil {
		return fmt.Errorf("写入前端哈希失败: %w", err)
	}
	return nil
}

func hashEmbeddedFS(src fs.FS) (string, error) {
	digest := sha256.New()
	err := fs.WalkDir(src, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if _, err := io.WriteString(digest, name); err != nil {
			return err
		}
		if _, err := io.WriteString(digest, "\x00"); err != nil {
			return err
		}
		file, err := src.Open(name)
		if err != nil {
			return err
		}
		if _, err := io.Copy(digest, file); err != nil {
			_ = file.Close()
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}
		_, err = io.WriteString(digest, "\x00")
		return err
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}
