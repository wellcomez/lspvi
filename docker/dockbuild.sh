#!/usr/bin/env bash
if [[ $1 == "create" ]]; then
    sudo docker build -t  lspvi.
    exit
fi

uid=$(id -u)
gid=$(id -g)
# if [[ ! -d flutter ]]; then
#     flutter_tar=flutter_linux_3.13.6-stable.tar.xz
#     if [[ ! -f $flutter_tar ]]; then
#         wget https://storage.googleapis.com/flutter_infra_release/releases/stable/linux/
#     fi
#     tar xf $flutter_tar
# fi
sudo docker run --user "$uid:$gid" \
    -v "$(pwd)"/home:/home/z/\
    -v /home/z/dev:/home/z/dev \
    --workdir "$(realpath ..)" -it lspvi 
