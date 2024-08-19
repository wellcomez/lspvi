#!/usr/bin/env bash
dockname=lspviu22
if [[ $1 == "create" ]]; then
    sudo docker build -t $dockname   .
    exit
fi
if [[ $1 == "push" ]]; then
    docker tag $dockname lailainux/$dockname:lastest
    sudo docker push lailainux/$dockname:lastest
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
    --workdir "$(realpath ..)" -it $dockname
