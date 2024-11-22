package selfupdater

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func (u *Updater) getAssets() error {
	log.Println("Fetching assets for the latest release...")
	ctx := context.Background()
	var checksumReader io.ReadCloser

	assets := []Asset{}
	for _, asset := range u.latestRelease.Assets {
		name := *asset.Name
		id := *asset.ID

		log.Printf("Processing asset: %s", name)
		if strings.Contains(name, "checksums") {
			var err error
			checksumReader, _, err = u.client.Repositories.DownloadReleaseAsset(ctx, u.Owner, u.Repo, id, http.DefaultClient)
			if err != nil {
				return fmt.Errorf("failed to download checksum asset: %v", err)
			}
			defer checksumReader.Close()
		}

		assets = append(assets, Asset{
			ID:   id,
			Name: name,
		})
	}

	checksumContent, err := io.ReadAll(checksumReader)
	if err != nil {
		return fmt.Errorf("failed to read checksum content: %v", err)
	}

	log.Println("Parsing checksum file...")
	lines := strings.Split(string(checksumContent), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}

		checksum := strings.TrimSpace(parts[0])
		assetName := strings.TrimSpace(parts[1])

		for i, a := range assets {
			if a.Name == assetName {
				assets[i].Checksum = checksum
			}
		}
	}

	u.assets = assets
	log.Println("Assets and checksums fetched successfully.")
	return nil
}

func (u *Updater) selectAsset() (Asset, error) {
	platform := getPlatform()
	name := fmt.Sprintf("%s_%s.tar.gz", u.BinaryName, platform)

	for _, asset := range u.assets {
		if asset.Name == name {
			return asset, nil
		}
	}

	return Asset{}, fmt.Errorf("no asset found for platform %s", platform)
}
