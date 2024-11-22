# SelfUpdater

## Motivation
Eliminate copy-paste self-update logic across personal CLI projects. Designed for GoReleaser-based release workflows.

## Features

- Automatic version checking against GitHub releases
- Platform-specific asset selection
- Checksum verification
- Safe binary update with rollback support

## Installation

```bash
go get github.com/marianozunino/selfupdater
```

## Usage

```go
package main

import selfupdater "github.com/marianozunino/selfupdater/pkg/selfupdater"

func main() {
    // Create an updater instance
    updater := selfupdater.NewUpdater(
        "your-github-username",   // GitHub owner
        "your-repo-name",         // Repository name
        "your-binary-name",       // Binary name (matching GoReleaser config)
        "current-version"         // Current version of your application
    )

    // Perform update check and apply if needed
    updater.Update()
}
```

### Optional Authentication

If you're using a private repository or need higher API rate limits:

```go
updater := selfupdater.NewUpdater(
    "owner",
    "repo",
    "binary",
    "version",
    selfupdater.WithToken("your-github-token")
)
```

## License

MIT License
