rename_binary() {
  outpname=$(basename "$1" .go)
  if [[ -n $compiler ]]; then
    mv "$outpname" "$outpname"_"$compiler"
    exit
  fi
  if [[ -f $outpname ]]; then
    if [[ -n $GOARM ]]; then
      mv "$outpname" "$outpname"_"$GOOS"_"$GOARCH"_"$GOARM"
    else
      mv "$outpname" "$outpname"_"$GOOS"_"$GOARCH"
    fi
  else
    mv "$outpname".exe "$outpname"_"$GOOS"_"$GOARCH".exe
  fi
}
if [[ -z $all ]]; then
  git submodule init && git submodule update --recursive
fi
build_os() {
  pushd pkg/lspr || exit
  bash build.sh -w
  popd || exit
  go build -o lspvi-$GOOS
  ls .
}
build_win_x64() {
  CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o lspvi-window-x64
}
# export GOARCH=amd64
export CGO_ENABLED=0
if [[ -n $mac ]]; then
  export GOOS=darwin
  build_os
fi
if [[ -n $win ]]; then
  build_win_x64
fi
if [[ -n $linux ]]; then
  export GOOS=linux
  build_os
fi

build_win_x64
