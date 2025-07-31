module github.com/Quok-it/dingo

go 1.24.4

require (
	github.com/hashicorp/go-multierror v1.1.1
	github.com/pkg/sftp v1.13.9
	golang.org/x/crypto v0.39.0
)

require (
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/kr/fs v0.1.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
)

// Replace directive for local development and CI
replace github.com/Quok-it/dingo => ./
