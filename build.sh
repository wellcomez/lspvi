git submodule init && git submodule update --recursive
build_os() {
  pushd pkg/lspr
  bash build.sh -w
  popd
  go build -o lspvi-$GOOS
}
# export GOARCH=amd64
export CGO_ENABLED=0 
export GOOS=darwin
build_os
export GOOS=linux
build_os

