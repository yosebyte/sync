package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"
)

func runSync(ctx context.Context, src, dst string) {
	if !mu.TryLock() {
		logger.Info("Sync in progress, skipping...")
		return
	}
	go func() {
		defer mu.Unlock()
		if cnt, err := syncFiles(ctx, src, dst); err != nil {
			logger.Error("Sync failed: %v", err)
			time.Sleep(time.Minute)
		} else {
			logger.Info("Sync complete: %v files", cnt)
		}
	}()
}

func syncFiles(ctx context.Context, src, dst string) (int, error) {
	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return 0, err
	}
	count := 0
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(dst, rel)
		if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
			return err
		}
		if err := copyFile(path, dest); err != nil {
			return err
		}
		count++
		return nil
	})
	return count, err
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
