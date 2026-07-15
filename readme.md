# qapp

a mini app framework, aim to integrate serveral popular go libs and start a app quick

## Limitation

* DONOT set qapp.Context in multiple functions in one init stage now.

## Howto

### override config

By default qapp reads `app.yml` as the main config file. If an `app.override.yml`
file exists, it is merged on top of the main config at startup, and values for the
same keys override those in the main config. If the override file does not exist it
is simply ignored, so behavior stays the same as before.

The override file name is derived from the main config file name by inserting
`.override` before the extension, for example:

| Main config (`-f`) | Derived override file |
| --- | --- |
| `app.yml` | `app.override.yml` |
| `config/prod.yaml` | `config/prod.override.yaml` |
| `app` | `app.override` |

You can also specify the override file name explicitly via `WithConfigOverrideFile`:

```go
qapp.New("myapp",
    qapp.WithConfigOverrideFile("/etc/myapp/local.yml"),
)
```

Notes:
- A missing override file is not an error.
- Precedence: command line flags / env vars > override config > main config.
- Hot reload (`WithConfigChanged`) still watches only the main config file; when the
  main config changes and is re-read, the override file is not automatically re-merged.

### add version information

``` shell
go build -ldflags "-X 'github.com/kkkbird/qapp.Version=1.0.0' -X 'github.com/kkkbird/qapp.BuildTime=`date`' -X 'github.com/kkkbird/qapp.GitHash=`git rev-parse HEAD`' -X 'github.com/kkkbird/qapp.GoVersion=`go version`'" .
```
