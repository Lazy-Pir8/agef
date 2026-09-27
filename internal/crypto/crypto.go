package crypto

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"filippo.io/age"
)

// ProgressFunc reports (completedFiles, totalFiles, currentRelativePath).
type ProgressFunc func(completed int64, total int64, currentPath string)

// EncryptFolder recursively encrypts all files in srcDir into destDir.
func EncryptFolder(srcDir, destDir string, recipient age.Recipient, progress ProgressFunc) (int64, error) {
	absSrc, err := filepath.Abs(srcDir)
	if err != nil {
		return 0, fmt.Errorf("invalid source path: %w", err)
	}

	srcInfo, err := os.Stat(absSrc)
	if err != nil || !srcInfo.IsDir() {
		return 0, fmt.Errorf("source is not a valid directory: %s", absSrc)
	}

	if destDir == "" {
		parent := filepath.Dir(absSrc)
		base := filepath.Base(absSrc)
		destDir = filepath.Join(parent, base+"_encrypted")
	}

	absDest, err := filepath.Abs(destDir)
	if err != nil {
		return 0, fmt.Errorf("invalid destination path: %w", err)
	}

	if absSrc == absDest {
		return 0, fmt.Errorf("destination cannot be identical to source directory")
	}

	if strings.HasPrefix(absDest, absSrc+string(filepath.Separator)) {
		return 0, fmt.Errorf("destination cannot be placed inside the source directory")
	}

	// 1. Scan files and replicate directories
	var filesToProcess []string
	err = filepath.WalkDir(absSrc, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(absSrc, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(absDest, rel)

		if d.IsDir() {
			info, err := d.Info()
			perm := os.FileMode(0755)
			if err == nil {
				perm = info.Mode().Perm()
			}
			return os.MkdirAll(targetPath, perm)
		}

		filesToProcess = append(filesToProcess, path)
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("error scanning directory: %w", err)
	}

	total := int64(len(filesToProcess))
	if total == 0 {
		return 0, nil
	}

	// 2. Parallel encryption pool
	workerCount := runtime.NumCPU() * 2
	if workerCount < 4 {
		workerCount = 4
	}

	jobs := make(chan string, workerCount)
	errChan := make(chan error, workerCount)
	var completed int64
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				rel, _ := filepath.Rel(absSrc, path)
				targetPath := filepath.Join(absDest, rel+".age")

				if err := encryptSingleFile(path, targetPath, recipient); err != nil {
					select {
					case errChan <- fmt.Errorf("failed to encrypt %s: %w", rel, err):
					default:
					}
					return
				}

				done := atomic.AddInt64(&completed, 1)
				if progress != nil {
					progress(done, total, rel)
				}
			}
		}()
	}

	// Send jobs
	go func() {
		for _, f := range filesToProcess {
			jobs <- f
		}
		close(jobs)
	}()

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return completed, <-errChan
	}

	return completed, nil
}

// DecryptFolder recursively decrypts all files in srcDir into destDir.
func DecryptFolder(srcDir, destDir string, identity age.Identity, progress ProgressFunc) (int64, error) {
	absSrc, err := filepath.Abs(srcDir)
	if err != nil {
		return 0, fmt.Errorf("invalid source path: %w", err)
	}

	srcInfo, err := os.Stat(absSrc)
	if err != nil || !srcInfo.IsDir() {
		return 0, fmt.Errorf("source is not a valid directory: %s", absSrc)
	}

	if destDir == "" {
		parent := filepath.Dir(absSrc)
		base := filepath.Base(absSrc)
		if strings.HasSuffix(base, "_encrypted") {
			destDir = filepath.Join(parent, strings.TrimSuffix(base, "_encrypted")+"_decrypted")
		} else {
			destDir = filepath.Join(parent, base+"_decrypted")
		}
	}

	absDest, err := filepath.Abs(destDir)
	if err != nil {
		return 0, fmt.Errorf("invalid destination path: %w", err)
	}

	if absSrc == absDest {
		return 0, fmt.Errorf("destination cannot be identical to source directory")
	}

	if strings.HasPrefix(absDest, absSrc+string(filepath.Separator)) {
		return 0, fmt.Errorf("destination cannot be placed inside the source directory")
	}

	// 1. Scan files and replicate directories
	var filesToProcess []string
	err = filepath.WalkDir(absSrc, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(absSrc, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(absDest, rel)

		if d.IsDir() {
			info, err := d.Info()
			perm := os.FileMode(0755)
			if err == nil {
				perm = info.Mode().Perm()
			}
			return os.MkdirAll(targetPath, perm)
		}

		filesToProcess = append(filesToProcess, path)
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("error scanning directory: %w", err)
	}

	total := int64(len(filesToProcess))
	if total == 0 {
		return 0, nil
	}

	// 2. Parallel decryption pool
	workerCount := runtime.NumCPU() * 2
	if workerCount < 4 {
		workerCount = 4
	}

	jobs := make(chan string, workerCount)
	errChan := make(chan error, workerCount)
	var completed int64
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				rel, _ := filepath.Rel(absSrc, path)
				var outRel string
				if strings.HasSuffix(rel, ".age") {
					outRel = strings.TrimSuffix(rel, ".age")
				} else {
					outRel = rel
				}

				targetPath := filepath.Join(absDest, outRel)

				if err := decryptSingleFile(path, targetPath, identity); err != nil {
					select {
					case errChan <- fmt.Errorf("failed to decrypt %s: %w", rel, err):
					default:
					}
					return
				}

				done := atomic.AddInt64(&completed, 1)
				if progress != nil {
					progress(done, total, outRel)
				}
			}
		}()
	}

	// Send jobs
	go func() {
		for _, f := range filesToProcess {
			jobs <- f
		}
		close(jobs)
	}()

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return completed, <-errChan
	}

	return completed, nil
}

func encryptSingleFile(srcPath, destPath string, recipient age.Recipient) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcStat, err := srcFile.Stat()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcStat.Mode().Perm())
	if err != nil {
		return err
	}
	defer destFile.Close()

	bufDest := bufio.NewWriterSize(destFile, 64*1024)
	w, err := age.Encrypt(bufDest, recipient)
	if err != nil {
		return err
	}

	bufSrc := bufio.NewReaderSize(srcFile, 64*1024)
	if _, err := io.Copy(w, bufSrc); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	if err := bufDest.Flush(); err != nil {
		return err
	}

	// Preserve timestamps
	_ = os.Chtimes(destPath, srcStat.ModTime(), srcStat.ModTime())
	return nil
}

func decryptSingleFile(srcPath, destPath string, identity age.Identity) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcStat, err := srcFile.Stat()
	if err != nil {
		return err
	}

	bufSrc := bufio.NewReaderSize(srcFile, 64*1024)
	r, err := age.Decrypt(bufSrc, identity)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcStat.Mode().Perm())
	if err != nil {
		return err
	}
	defer destFile.Close()

	bufDest := bufio.NewWriterSize(destFile, 64*1024)
	if _, err := io.Copy(bufDest, r); err != nil {
		return err
	}

	if err := bufDest.Flush(); err != nil {
		return err
	}

	// Preserve timestamps
	_ = os.Chtimes(destPath, srcStat.ModTime(), srcStat.ModTime())
	return nil
}
