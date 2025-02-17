package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"
	"time"
)

func runSync(ctx context.Context, src, dst string) {
	if !mu.TryLock() {
		logger.Info("Sync already in progress")
		return
	}
	go func() {
		defer mu.Unlock()
		if cnt, skp, err := syncFiles(ctx, src, dst); err != nil {
			logger.Error("Sync failed: %v", err)
			time.Sleep(time.Minute)
		} else {
			logger.Info("Sync complete: %v files", cnt)
			logger.Info("Files skipped: %v files", skp)
		}
	}()
}

func syncFiles(ctx context.Context, src, dst string) (int, int, error) {
	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return 0, 0, err
	}
	count, skipped := 0, 0
	err := filepath.Walk(src, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(dst, rel)
		if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
			return err
		}
		if _, err := os.Stat(dest); err == nil {
			srcHash, err := fileHash(path)
			if err != nil {
				return err
			}
			dstHash, err := fileHash(dest)
			if err != nil {
				return err
			}
			if bytes.Equal(srcHash, dstHash) {
				skipped++
				return nil
			}
		}
		if err := copyFile(path, dest); err != nil {
			return err
		}
		count++
		return nil
	})
	return count, skipped, err
}

func fileHash(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
