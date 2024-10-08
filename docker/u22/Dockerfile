# FROM ubuntu:20.04

FROM ubuntu:22.04 
# FROM buildpack-deps:bullseye
RUN apt update
RUN apt-get install -y gpg
RUN apt-get install -y ca-certificates
COPY  sourcelist  /etc/apt/sources.list
RUN apt update
RUN groupadd --gid 1000 node \
    && useradd --uid 1000 --gid node --shell /bin/bash --create-home node
RUN apt-get install -y curl tar
RUN apt-get install -y xz-utils
RUN apt-get install -y wget 

RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
    g++ \
    gcc \
    libc6-dev \
    make \
    pkg-config \
    clangd \
    ; 
RUN apt-get install -y python3 pip
RUN apt-get install -y zsh 
RUN apt-get install -y git curl wget
RUN apt-get install -y tmux
RUN apt-get install -y openjdk-17-jdk

ENV NODE_VERSION 18.20.4
ENV ARCH=x64
RUN export GNUPGHOME="$(mktemp -d)" \
    && set -ex \
    && for key in \
    4ED778F539E3634C779C87C6D7062848A1AB005C \
    141F07595B7B3FFE74309A937405533BE57C7D57 \
    74F12602B6F1C4E913FAA37AD3A89613643B6201 \
    DD792F5973C6DE52C432CBDAC77ABFA00DDBF2B7 \
    61FC681DFB92A079F1685E77973F295594EC4689 \
    8FCCA13FEF1D0C2E91008E09770F7A9A5AE15600 \
    C4F0DFFF4E8C1A8236409D08E73BC641CC11F4C8 \
    890C08DB8579162FEE0DF9DB8BEAB4DFCF555EF4 \
    C82FA3AE1CBEDC6BE46B9360C43CEC45C17AB93C \
    108F52B48DB57BB0CC439B2997B01419BD92F80A \
    A363A499291CBBC940DD62E41F10027AF002F8B0 \
    CC68F5A3106FF448322E48ED27F5E38D5B0A215F \
    ; do \
    gpg --batch --keyserver hkps://keys.openpgp.org --recv-keys "$key" || \
    gpg --batch --keyserver keyserver.ubuntu.com --recv-keys "$key" ; \
    done \
    && curl -fsSLO --compressed "https://nodejs.org/dist/v$NODE_VERSION/node-v$NODE_VERSION-linux-$ARCH.tar.xz" \
    && curl -fsSLO --compressed "https://nodejs.org/dist/v$NODE_VERSION/SHASUMS256.txt.asc" \
    && gpg --batch --decrypt --output SHASUMS256.txt SHASUMS256.txt.asc \
    && gpgconf --kill all \
    && rm -rf "$GNUPGHOME" \
    && grep " node-v$NODE_VERSION-linux-$ARCH.tar.xz\$" SHASUMS256.txt | sha256sum -c - \
    && tar -xJf "node-v$NODE_VERSION-linux-$ARCH.tar.xz" -C /usr/local --strip-components=1 --no-same-owner \
    && rm "node-v$NODE_VERSION-linux-$ARCH.tar.xz" SHASUMS256.txt.asc SHASUMS256.txt \
    && ln -s /usr/local/bin/node /usr/local/bin/nodejs \
    # smoke tests
    && node --version \
    && npm --version

ENV YARN_VERSION 1.22.19

RUN set -ex \
    # use pre-existing gpg directory, see https://github.com/nodejs/docker-node/pull/1895#issuecomment-1550389150
    && export GNUPGHOME="$(mktemp -d)" \
    && for key in \
    6A010C5166006599AA17F08146C2130DFD2497F5 \
    ; do \
    gpg --batch --keyserver hkps://keys.openpgp.org --recv-keys "$key" || \
    gpg --batch --keyserver keyserver.ubuntu.com --recv-keys "$key" ; \
    done \
    && curl -fsSLO --compressed "https://yarnpkg.com/downloads/$YARN_VERSION/yarn-v$YARN_VERSION.tar.gz" \
    && curl -fsSLO --compressed "https://yarnpkg.com/downloads/$YARN_VERSION/yarn-v$YARN_VERSION.tar.gz.asc" \
    && gpg --batch --verify yarn-v$YARN_VERSION.tar.gz.asc yarn-v$YARN_VERSION.tar.gz \
    && gpgconf --kill all \
    && rm -rf "$GNUPGHOME" \
    && mkdir -p /opt \
    && tar -xzf yarn-v$YARN_VERSION.tar.gz -C /opt/ \
    && ln -s /opt/yarn-v$YARN_VERSION/bin/yarn /usr/local/bin/yarn \
    && ln -s /opt/yarn-v$YARN_VERSION/bin/yarnpkg /usr/local/bin/yarnpkg \
    && rm yarn-v$YARN_VERSION.tar.gz.asc yarn-v$YARN_VERSION.tar.gz \
    # smoke test
    && yarn --version \
    && rm -rf /tmp/*



ENV PATH /usr/local/go/bin:$PATH

ENV GOLANG_VERSION 1.21.3

RUN set -eux; \
    now="$(date '+%s')"; \
    arch="$(dpkg --print-architecture)"; arch="${arch##*-}"; \
    url='https://dl.google.com/go/go1.21.3.linux-amd64.tar.gz'; \
    # sha256='999805bed7d9039ec3da1a53bfbcafc13e367da52aa823cb60b68ba22d44c616'; \
    \
    wget -O go.tgz.asc "$url.asc"; \
    wget -O go.tgz "$url" --progress=dot:giga; \
    # echo "$sha256 *go.tgz" | sha256sum -c -; \
    \
    # https://github.com/golang/go/issues/14739#issuecomment-324767697
    GNUPGHOME="$(mktemp -d)"; export GNUPGHOME; \
    # https://www.google.com/linuxrepositories/
    gpg --batch --keyserver keyserver.ubuntu.com --recv-keys 'EB4C 1BFD 4F04 2F6D DDCC  EC91 7721 F63B D38B 4796'; \
    # let's also fetch the specific subkey of that key explicitly that we expect "go.tgz.asc" to be signed by, just to make sure we definitely have it
    gpg --batch --keyserver keyserver.ubuntu.com --recv-keys '2F52 8D36 D67B 69ED F998  D857 78BD 6547 3CB3 BD13'; \
    gpg --batch --verify go.tgz.asc go.tgz; \
    gpgconf --kill all; \
    rm -rf "$GNUPGHOME" go.tgz.asc; \
    \
    tar -C /usr/local -xzf go.tgz; \
    rm go.tgz; \
    \
    # save the timestamp from the tarball so we can restore it for reproducibility, if necessary (see below)
    SOURCE_DATE_EPOCH="$(stat -c '%Y' /usr/local/go)"; \
    export SOURCE_DATE_EPOCH; \
    touchy="$(date -d "@$SOURCE_DATE_EPOCH" '+%Y%m%d%H%M.%S')"; \
    # for logging validation/edification
    date --date "@$SOURCE_DATE_EPOCH" --rfc-2822; \
    # sanity check (detected value should be older than our wall clock)
    [ "$SOURCE_DATE_EPOCH" -lt "$now" ]; \
    \
    if [ "$arch" = 'armhf' ]; then \
    [ -s /usr/local/go/go.env ]; \
    before="$(go env GOARM)"; [ "$before" != '7' ]; \
    { \
    echo; \
    echo '# https://github.com/docker-library/golang/issues/494'; \
    echo 'GOARM=7'; \
    } >> /usr/local/go/go.env; \
    after="$(go env GOARM)"; [ "$after" = '7' ]; \
    # (re-)clamp timestamp for reproducibility (allows "COPY --link" to be more clever/useful)
    touch -t "$touchy" /usr/local/go/go.env /usr/local/go; \
    fi; \
    \
    # ideally at this point, we would just "COPY --link ... /usr/local/go/ /usr/local/go/" but BuildKit insists on creating the parent directories (perhaps related to https://github.com/opencontainers/image-spec/pull/970), and does so with unreproducible timestamps, so we instead create a whole new "directory tree" that we can "COPY --link" to accomplish what we want
    mkdir /target /target/usr /target/usr/local; \
    mv -vT /usr/local/go /target/usr/local/go; \
    ln -svfT /target/usr/local/go /usr/local/go; \
    touch -t "$touchy" /target/usr/local /target/usr /target; \
    \
    # smoke test
    go version; \
    # make sure our reproducibile timestamp is probably still correct (best-effort inline reproducibility test)
    epoch="$(stat -c '%Y' /target/usr/local/go)"; \
    [ "$SOURCE_DATE_EPOCH" = "$epoch" ]; \
    find /target -newer /target/usr/local/go -exec sh -c 'ls -ld "$@" && exit "$#"' -- '{}' +

# FROM buildpack-deps:bookworm-scm

# install cgo-related dependencies



RUN npm install -g typescript-language-server typescript
# RUN apt-get install -y python-lsp-server

WORKDIR /home/node
USER node
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"


ENV GOLANG_VERSION 1.21.3

# don't auto-upgrade the gotoolchain
# https://github.com/docker-library/golang/issues/472
ENV GOTOOLCHAIN=local

ENV GOPATH /home/node/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
# (see notes above about "COPY --link")
# COPY --from=build --link /target/ /
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 1777 "$GOPATH"
WORKDIR $GOPATH
RUN go install golang.org/x/tools/gopls@latest
RUN echo '' >>~/.zshrc
RUN echo 'export PROMPT="lspvi %(?:%{%}%1{➜%} :%{%}%1{➜%} ) %{%}%c%{%} $(git_prompt_info)"' >>~/.zshrc

USER root
COPY plantuml-1.2024.6.jar /usr/local/bin/
RUN ln -s /home/node/go/bin/gopls /usr/local/bin/gopls
COPY lspvi  /usr/local/bin/
RUN apt-get install -y git_prompt_info

USER node
ENTRYPOINT ["/bin/zsh"]
