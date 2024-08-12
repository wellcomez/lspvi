git submodule init  && git submodule update --recursive
pushd pkg/lspr
bash build.sh -w
popd
go build