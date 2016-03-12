.PHONY: build deps test clean

build:
	go build ./...

deps:
	go get github.com/cihub/seelog
	go get github.com/facebookgo/inject
	go get github.com/mjibson/esc
	go get golang.org/x/net/context
	go get golang.org/x/oauth2
	go get golang.org/x/oauth2/google
	go get golang.org/x/oauth2/jwt
	go get google.golang.org/api/cloudmonitoring/v2beta2
	go get google.golang.org/api/compute/v1
	go get google.golang.org/api/replicapoolupdater/v1beta1
	go get google.golang.org/api/sqladmin/v1beta4
	go get google.golang.org/cloud
	go get google.golang.org/cloud/datastore

test:
	go test ./...
