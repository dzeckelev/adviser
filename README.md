# Adviser

### Preparation

Install Golang 1.11, Docker and Docker-Compose.

### Build

```bash
cd $GOPATH/src/github.com/dzeckelev
git clone https://github.com/dzeckelev/adviser.git
cd adviser
make build
```

### Config

- `Addr` - Local http server address.
- `CacheSize` - cache size.
- `LogLevel` - Log level, must be `debug`, `info`, `warn`, `error`, `panic`, `dpanic` or `fatal`
- `RequestTimeout` - Request timeout in milliseconds.
- `TargetAddr` - Address of the target service.

### Run

```bash
make run
```

or as a demon:

```bash
make demon
```

### Stop

```bash
make stop
```

### Sample

Request:

```bash
curl -X GET -H "Content-Type: application/json" "http://localhost:8081/v2/places.json?term=%D0%9C%D0%BE%D1%81%D0%BA%D0%B2%D0%B0&locale=ru&types%5B%5D=city&types%5B%5D=airport"
```

Response:

```bash
[{"slug":"MOW","subtitle":"Россия","title":"Москва"},{"slug":"DME","subtitle":"Россия","title":"Домодедово"},{"slug":"SVO","subtitle":"Россия","title":"Шереметьево"},{"slug":"VKO","subtitle":"Россия","title":"Внуково"},{"slug":"ZIA","subtitle":"Россия","title":"Жуковский"}]
```
### Environment

```bash
Linux host 4.15.0-45-generic
Docker version 18.06.1-ce, build e68fc7a
docker-compose version 1.23.2, build 1110ad01
```