.PHONY: install deps 

install:
	sh scripts/build-asset.sh
	go install ./...

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
	go get google.golang.org/api/pubsub/v1
	go get google.golang.org/api/replicapoolupdater/v1beta1
	go get google.golang.org/api/sqladmin/v1beta4
	go get google.golang.org/api/storage/v1
	go get google.golang.org/cloud
	go get google.golang.org/cloud/datastore
	# below is for test
	go get github.com/stretchr/testify/suite
	go get github.com/patrickmn/go-cache
	go get github.com/satori/go.uuid
	sh scripts/build-asset.sh
