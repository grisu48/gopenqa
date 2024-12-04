default: gopenqa
gopenqa: cmd/gopenqa/*.go *.go
	go build -o gopenqa cmd/gopenqa/*.go
