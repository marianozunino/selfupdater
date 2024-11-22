module github.com/marianozunino/selfupdater

go 1.23.0

retract [v0.0.1, v0.0.2]

require (
	github.com/google/go-github/v66 v66.0.0
	github.com/minio/selfupdate v0.6.0
)

require (
	aead.dev/minisign v0.2.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	golang.org/x/crypto v0.0.0-20211209193657-4570a0811e8b // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
)
