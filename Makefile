# Include enviroment variables
include .env

# binuction
build:
	@CGO_ENABLED=0 go build -o bin/zero -ldflags="-extldflags=-static" .
run: build
	@./bin/zero
stripe: 
	@stripe listen --forward-to $(STRIPE_DOMAIN)/api/$(API_VERSION)/stripe/webhook

# Cleanup
clean:
	rm bin/*
	
# Goose
up:
	@GOOSE_DRIVER=$(DATABASE_DRIVER) GOOSE_DBSTRING=db/$(DATABASE_NAME) goose -dir=db/migrations up
down:
	@GOOSE_DRIVER=$(DATABASE_DRIVER) GOOSE_DBSTRING=db/$(DATABASE_NAME) goose -dir=db/migrations down
status:
	@GOOSE_DRIVER=$(DATABASE_DRIVER) GOOSE_DBSTRING=db/$(DATABASE_NAME) goose -dir=db/migrations status

release:
	@env CGO_DISABLED=0 GOOS="windows" GOARCH="amd64" go build -o bin/zero_windows_amd64.exe -ldflags="-s -w -extldflags=-static" .
	@env CGO_DISABLED=0 GOOS="darwin" GOARCH="amd64" go build -o bin/zero_macos_amd64 -ldflags="-s -w -extldflags=-static" .
	@env CGO_DISABLED=0 GOOS="linux" GOARCH="amd64" go build -o bin/zero_linux_amd64 -ldflags="-s -w -extldflags=-static" .
	@env CGO_DISABLED=0 GOOS="freebsd" GOARCH="amd64" go build -o bin/zero_freebsd_amd64 -ldflags="-s -w -extldflags=-static" .
	@env CGO_DISABLED=0 GOOS="openbsd" GOARCH="amd64" go build -o bin/zero_openbsd_amd64 -ldflags="-s -w -extldflags=-static" .