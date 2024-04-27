#!/bin/sh

set -e

SOURCE="${PWD}/source"
RELEASE="${PWD}/release"
PACKED="${PWD}/packed"

# convert to NFD to enhance compatibility with tools
NAME=$(iconv -t UTF-8-MAC <<< "${PWD##*/}")

function gob {
    pushd "$SOURCE"
    gobuild.sh "$NAME" "$1" "$2" '' '' "$3" '' "$RELEASE/$OS"
    popd
}

function pack_if_requested {
    if [[ "$PACK" != "" ]]; then
        if [[ "$OS" == "mac" ]]; then
            PACKED_NAME="$NAME-$OS-$ARCHS-$PACK.dmg"
            rm -f "$PACKED/$PACKED_NAME"
            create-dmg \
                --window-pos 200 120 \
                --window-size 640 540 \
                --icon "$NAME.app" 170 110 \
                --app-drop-link 470 110 \
                --icon-size 160 \
                --background "$RELEASE/$OS/$NAME.app/Contents/Resources/installer-mac.jpeg" \
                "$PACKED/$PACKED_NAME" "$RELEASE/$OS"
        else
            PACKED_NAME="$NAME-$OS-$ARCHS-$PACK.zip"
            rm -f "$PACKED/$PACKED_NAME"
            pushd "$RELEASE/$OS"
            zip -9r "$PACKED/$PACKED_NAME" *
            popd
        fi
    fi
}

function list_package {
    find "$RELEASE/$OS" -type f -ls
    echo "<-- did list $OS package."
}

export CGO_ENABLED=1

echo "Building $NAME..."

pushd "$SOURCE"
    find . -name .DS_Store -ls -delete
    . prepare.sh
popd

OS='mac'
export COMPRESS=
OSARCH="$(uname -m)"
if [[ "$OSARCH" == "arm64" ]]; then
    # unsupported GOOS/GOARCH pair darwin/386
    # gob 'darwin'   '386' '-i386'
    gob 'darwin' 'amd64' '-amd64'
    gob 'darwin' 'arm64' '-arm64'
    BIN="$RELEASE/$OS/$NAME"
    echo "Creating universal binary"
    lipo -create -output "$BIN" "$BIN-amd64" "$BIN-arm64"
    rm -v "$BIN-amd64"
    rm -v "$BIN-arm64"
    ARCHS="amd64+arm64"
else
    gob 'darwin' 'amd64' ''
    ARCHS="amd64"
fi
echo "--> will create $OS package..."
mkdir -p "$RELEASE/$OS/$NAME.app/Contents"
rsync -av --delete-after "$SOURCE/packaging/$OS/" "$RELEASE/$OS/$NAME.app/Contents/"
mkdir -p "$RELEASE/$OS/$NAME.app/Contents/MacOS/config"
rsync -av --delete-after "$SOURCE/config/$OS/" "$RELEASE/$OS/$NAME.app/Contents/MacOS/config"
mv -f "$RELEASE/$OS/$NAME" "$RELEASE/$OS/$NAME.app/Contents/MacOS"
echo "<-- did create $OS package."
pack_if_requested
list_package
echo pkill -x $NAME || true
pkill -x $NAME || true
echo "<-- did stop $OS app -->"
if [[ "$OSARCH" == "arm64" ]]; then
    echo "<-- did skip launch $OS app -->"
else
    sleep 2
    open -a "$RELEASE/$OS/$NAME.app"
    echo "<-- did launch $OS app -->"
fi

OS='win'
export COMPRESS='upx'
export LDFLAGS="-H=windowsgui"
gob 'windows' 'amd64' '.exe'
ARCHS="amd64"
echo "--> will create $OS package..."
rsync -av --delete-after "$SOURCE/config/$OS/" "$RELEASE/$OS/config"
echo "<-- did create $OS package."
pack_if_requested
find "$RELEASE/$OS" -type f -ls
echo "<-- did list $OS package."

echo "Done."
