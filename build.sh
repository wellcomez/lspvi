git submodule init && git submodule update --recursive
build_os() {
  pushd pkg/lspr || exit
  bash build.sh -w
  popd || exit
  go build -o lspvi-$GOOS
  ls .
}
# export GOARCH=amd64
export CGO_ENABLED=0 
export GOOS=darwin
build_os
export GOOS=linux
build_os

