# qapp

a mini app framework, aim to integrate serveral popular go libs and start a app quick

## Limitation

* DONOT set qapp.Context in multiple functions in one init stage now.

## Howto

### add version information

``` shell
go build -ldflags "-X 'github.com/kkkbird/qapp.Version=1.0.0' -X 'github.com/kkkbird/qapp.BuildTime=`date`' -X 'github.com/kkkbird/qapp.GitHash=`git rev-parse HEAD`' -X 'github.com/kkkbird/qapp.GoVersion=`go version`'" .
```
