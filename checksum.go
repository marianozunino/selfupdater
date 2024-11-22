package selfupdater

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (u *Updater) verifyChecksum(filename, expectedChecksum string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to calculate SHA-256: %v", err)
	}

	calculatedChecksum := hex.EncodeToString(hash.Sum(nil))
	if calculatedChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, calculatedChecksum)
	}

	return nil
}

func (u *Updater) extractBinary(tarPath, targetDir string) (string, error) {
	file, err := os.Open(tarPath)
	if err != nil {
		return "", fmt.Errorf("failed to open tar.gz file: %v", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("error reading tar header: %v", err)
		}

		if header.Typeflag == tar.TypeReg && filepath.Base(header.Name) == u.BinaryName {
			extractedPath := filepath.Join(targetDir, u.BinaryName)
			outFile, err := os.OpenFile(extractedPath, os.O_CREATE|os.O_WRONLY, 0o755)
			if err != nil {
				return "", fmt.Errorf("failed to create binary file: %v", err)
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tr); err != nil {
				return "", fmt.Errorf("failed to extract binary: %v", err)
			}

			return extractedPath, nil
		}
	}

	return "", fmt.Errorf("binary %s not found in archive", u.BinaryName)
}
