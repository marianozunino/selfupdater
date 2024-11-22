package selfupdater

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v66/github"
	"github.com/minio/selfupdate"
)

type Updater struct {
	CurrentVersion string
	Owner          string
	Repo           string
	BinaryName     string
	client         *github.Client
	latestRelease  *github.RepositoryRelease
	assets         []Asset
}

func NewUpdater(owner, repo, binaryName, currentVersion string, options ...Option) *Updater {
	u := &Updater{
		Owner:          owner,
		Repo:           repo,
		BinaryName:     binaryName,
		CurrentVersion: currentVersion,
		client:         github.NewClient(nil),
	}

	for _, option := range options {
		option(u)
	}

	return u
}

func (u *Updater) Update() {
	log.Println("Updater starting...")

	if err := u.getLatestRelease(); err != nil {
		log.Fatalf("Get latest release failed: %v", err)
	}

	if u.isUpToDate() {
		log.Println("Current version is up to date.")
		return
	}

	if err := u.getAssets(); err != nil {
		log.Fatalf("Get assets failed: %v", err)
	}

	asset, err := u.selectAsset()
	if err != nil {
		log.Fatalf("Select asset failed: %v", err)
	}

	if err := u.downloadAndApply(asset); err != nil {
		log.Fatalf("Download and apply failed: %v", err)
	}

	log.Println("Update successful!")
}

func (u *Updater) getLatestRelease() error {
	log.Println("Fetching the latest release information...")
	ctx := context.Background()
	release, _, err := u.client.Repositories.GetLatestRelease(ctx, u.Owner, u.Repo)
	if err != nil {
		return fmt.Errorf("failed to get latest release: %v", err)
	}
	u.latestRelease = release
	log.Printf("Latest release found: %s", *u.latestRelease.TagName)
	return nil
}

func (u *Updater) isUpToDate() bool {
	log.Println("Checking if the current version is up to date...")
	current := strings.TrimPrefix(u.CurrentVersion, "v")
	if u.latestRelease == nil {
		return false
	}
	latest := strings.TrimPrefix(*u.latestRelease.TagName, "v")
	log.Printf("Current version: %s, Latest version: %s", current, latest)
	return current == latest
}

func (u *Updater) downloadAndApply(asset Asset) error {
	log.Printf("Starting download for asset: %s", asset.Name)

	tempDir, err := os.MkdirTemp("", fmt.Sprintf("%s-update", u.BinaryName))
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	log.Printf("Temporary directory created: %s", tempDir)

	tarPath := filepath.Join(tempDir, fmt.Sprintf("%s.tar.gz", u.BinaryName))

	ctx := context.Background()
	reader, _, err := u.client.Repositories.DownloadReleaseAsset(ctx, u.Owner, u.Repo, asset.ID, http.DefaultClient)
	if err != nil {
		return fmt.Errorf("failed to download asset: %v", err)
	}
	defer reader.Close()

	outFile, err := os.Create(tarPath)
	if err != nil {
		return fmt.Errorf("failed to create tarball file: %v", err)
	}
	defer outFile.Close()

	log.Printf("Saving asset to temporary tarball: %s", tarPath)
	if _, err := io.Copy(outFile, reader); err != nil {
		return fmt.Errorf("failed to save tarball: %v", err)
	}

	log.Println("Verifying checksum...")
	if err := u.verifyChecksum(tarPath, asset.Checksum); err != nil {
		return fmt.Errorf("checksum validation failed: %v", err)
	}
	log.Println("Checksum verified successfully.")

	log.Println("Extracting binary from tarball...")
	extractedBinaryPath, err := u.extractBinary(tarPath, tempDir)
	if err != nil {
		return fmt.Errorf("failed to extract binary: %v", err)
	}
	log.Printf("Binary extracted to: %s", extractedBinaryPath)

	log.Println("Applying update...")
	if err := u.applyUpdate(extractedBinaryPath); err != nil {
		return fmt.Errorf("failed to apply update: %v", err)
	}

	log.Println("Update applied successfully.")
	return nil
}

func (u *Updater) applyUpdate(binaryPath string) error {
	log.Printf("Applying update from binary: %s", binaryPath)

	binary, err := os.Open(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to open extracted binary: %v", err)
	}
	defer binary.Close()

	err = selfupdate.Apply(binary, selfupdate.Options{})
	if err != nil {
		log.Printf("Update failed: %v. Attempting rollback...", err)
		if rollbackErr := selfupdate.RollbackError(err); rollbackErr != nil {
			return fmt.Errorf("rollback failed: %v", rollbackErr)
		}
		return fmt.Errorf("update failed and rollback was successful: %v", err)
	}

	return nil
}
