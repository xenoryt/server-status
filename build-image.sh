TAG="${1:-latest}"

set -x

CGO_ENABLED=0 go build -a -installsuffix cgo .
docker build . -t xenoryt/stream-status:$TAG
