#!/bin/sh

set -e

# to cross-compile:
# https://blog.filippo.io/easy-windows-and-linux-cross-compilers-for-macos/
# brew install FiloSottile/musl-cross/musl-cross
# brew install mingw-w64

# to use some windows libs:
# export GOOS=windows; go get "github.com/lxn/win"
# export GOOS=windows; go get "gopkg.in/Knetic/govaluate.v3"
# export GOOS=windows; go get -u -v "golang.org/x/crypto/..." (with the ...)

BUILD="$(uname -n)／$(uname -r)／$(date +'%Y-%m-%d／%T')"

OSARCH="$(uname -m)"

#LDFLAGS="$LDFLAGS -linkmode=internal"

#GCFLAGS="$GCFLAGS -l"

#LOOPVAR=S
if [[ $LOOPVAR == "S" ]]; then
    GCFLAGS="$GCFLAGS -d=loopvar=2"
    export GOEXPERIMENT=loopvar
fi

echo "name      = [$1]"
echo "goos      = [$2]"
echo "goarch    = [$3]"
echo "os        = [$4]"
echo "arch      = [$5]"
echo "extension = [$6]"
echo "version   = [$7]"
echo "output    = [$8]"

export GOOS="$2"
export GOARCH="$3"
echo "--> will build $GOOS-$GOARCH"
FILE="/tmp/$1$4$5$7$6"

if [[ "$GOOS" == "windows" ]]; then
    export CC="x86_64-w64-mingw32-gcc"
elif [[ "$GOARCH" == "amd64" && "$OSARCH" == "arm64" ]]; then
    export CC="clang -arch x86_64 -mmacosx-version-min=10.14"
fi

#CGO_ENABLED=$_CGO_ENABLED \
#GOEXPERIMENT="$_GOEXPERIMENT" \

# move the .syso out of the way for non-windows builds
SYSO=resource.syso
[ ! -f $SYSO ] || mv -vf $SYSO ../$SYSO
if [[ "$GOOS" == "windows" ]]; then
    [ ! -f ../$SYSO ] || cp -v ../$SYSO $SYSO
fi
# a .syso file will cause an 'ld: warning'
go build \
    -ldflags "-s -w $LDFLAGS -X main.build=$BUILD" \
    -gcflags=all="$GCFLAGS" \
    -o "$FILE"

if [[ "$COMPRESS" == "upx" ]]; then
    upx --lzma "$FILE" || echo "Could not run UPX, skipping"
else
    echo "Not compressing"
fi

mkdir -p "$8"
mv -f "$FILE" "$8/"
echo "<-- did build $GOOS-$GOARCH"
