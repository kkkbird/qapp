# bshark

a mini app framework, aim to integrate serveral popular go libs and start a app quick

## Limitation

* DONOT set bshark.Context in multiple functions in one init stage now.

## Howto

### add version

go build -ldflags "-X github.com/kkkbird/bshark.Version=1.0.0 -X 'github.com/kkkbird/bshark.BuildTime=`date`' -X 'github.com/kkkbird/bshark.GitHash=`git rev-parse HEAD`' -X'github.com/kkkbird/bshark.GoVersion=`go version`'" .
