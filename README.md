# gogoo 

[![Join the chat at https://gitter.im/browny/gogoo](https://badges.gitter.im/browny/gogoo.svg)](https://gitter.im/browny/gogoo?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

[![GoDoc](https://godoc.org/github.com/browny/gogoo?status.svg)](http://godoc.org/github.com/browny/gogoo)
[![Travis Build Status](https://travis-ci.org/browny/gogoo.svg?branch=master)](https://travis-ci.org/browny/gogoo)
[![Go Report Card](https://goreportcard.com/badge/github.com/browny/gogoo)](https://goreportcard.com/report/github.com/browny/gogoo)

**gogoo** encapsulates [google cloud api](https://godoc.org/google.golang.org/api) for more specific operation logic. Below are
the including components

- [compute engine](https://godoc.org/google.golang.org/api/compute/v1) - v1
- [datastore](https://godoc.org/google.golang.org/cloud/datastore)
- [cloud monitoring](https://godoc.org/google.golang.org/api/cloudmonitoring/v2beta2) - v2beta2
- [cloudsql](https://godoc.org/google.golang.org/api/sqladmin/v1beta4) - v1beta4
- [replicapoolupdater](https://godoc.org/google.golang.org/api/replicapoolupdater/v1beta1) -v1beta1
- [pubsub](https://godoc.org/google.golang.org/api/pubsub/v1) - v1
- [storage](https://godoc.org/google.golang.org/api/storage/v1) - v1


## Install

```bash
go get github.com/browny/gogoo
```

## Develop

- Clone this project to your `$GOPATH/src`

```sh
cd $GOPATH/src
git clone git@github.com:browny/gogoo.git github.com/browny/gogoo
```

- You should setup one google cloud project, and create a [service account](https://developers.google.com/identity/protocols/OAuth2ServiceAccount)
- Enable the relating API you want to test
- Create a `./gogoo/config/config.json` file to containes below information

```json
{                                                                                                                         
  "service_account": "ooxx@developer.gserviceaccount.com",
  "project_id": "your_project_name"
}
```
- Put the key of service account in `./gogoo/config/key.pem` 

## Reference
- [Converting the service account credential to other formats](https://cloud.google.com/storage/docs/authentication#converting-the-private-key) (`.p12` to `.pem`)


## License

gogoo is MIT License.
