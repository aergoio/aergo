#!/usr/bin/env bash

# This script can be used to build the Docker images manually (outside of CI)

set -e

MAIN_TAG=$1
SECOND_TAG=$2
THIRD_TAG=$3

if [[ -z "$MAIN_TAG" ]]
then
    echo "Usage:"
    echo "  build-local.sh tag [second-tag] [third-tag]"
    echo "Example:"
    echo "  build-local.sh 0.12.0-rc"
    echo "  build-local.sh 0.12.0 0.12 latest"
    exit 1
fi


echo "Preparing local folder for build"

cd ..
git submodule update --init --recursive
make clean || true
rm -rf build
go clean --cache


if [[ -z "$THIRD_TAG" ]]
then
    if [[ -z "$SECOND_TAG" ]]
    then
        declare -a tags=("$MAIN_TAG")
    else
        declare -a tags=("$MAIN_TAG" "$SECOND_TAG")
    fi
else
    declare -a tags=("$MAIN_TAG" "$SECOND_TAG" "$THIRD_TAG")
fi

echo "Building Docker images for ${tags[*]} using local folder"
sleep 1

BUILDER_TAG="aergo/local-builder"
echo "Building ${BUILDER_TAG}"

docker build --no-cache --file Docker/Dockerfile.local -t ${BUILDER_TAG} .
cd -
docker create --name extract ${BUILDER_TAG}
docker cp extract:/go/aergo/bin/ .
docker cp extract:/go/aergo/cmd/brick/arglog.toml bin/brick-arglog.toml
docker cp extract:/go/aergo/libtool/lib/ .
docker rm -f extract

declare -a names=("node" "tools" "polaris")
for name in "${names[@]}"
do
    tagsExpanded=()
    for tag in "${tags[@]}"; do
        tagsExpanded+=("-t aergo/$name:$tag")
    done
    echo "[aergo/$name:${tags[*]}]"
    DOCKERFILE="Dockerfile.$name"
    echo docker build -q ${tagsExpanded[@]} --file $DOCKERFILE .
    imageid=`docker build -q ${tagsExpanded[@]} --file $DOCKERFILE .`
    docker images --format "Done: \t{{.Repository}}:{{.Tag}} \t{{.ID}} ({{.Size}})" | grep "${imageid:7:12}"
done

rm -rf bin lib

echo -e "\nREPOSITORY          TAG                 IMAGE ID            CREATED             SIZE"
for name in "${names[@]}"
do
    for tag in "${tags[@]}"
    do
        docker images aergo/$name:$tag | tail -1
    done
done

echo -e "\nYou can now push these to Docker Hub."
echo "For example:"

declare -a names=("node" "tools" "polaris")
for name in "${names[@]}"
do
    for tag in "${tags[@]}"
    do
        echo "  docker push aergo/$name:$tag"
    done
done
