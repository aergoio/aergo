# This script can be used to build the Docker images manually (outside of CI)
set -e

GIT_TAG=$1
MAIN_TAG=$2
SECOND_TAG=$3
THIRD_TAG=$4
if [[ -z "$MAIN_TAG" || -z "$GIT_TAG" ]]
then
    echo "Usage:"
    echo "  build.sh git-tag-or-hash tag [second-tag] [third-tag]"
    echo "Example:"
    echo "  build.sh release/0.12 0.12.0-rc"
    echo "  build.sh release/0.12 0.12.0 0.12 latest"
    exit 1
fi

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

echo "Building Docker images for ${tags[*]} using git tag $GIT_TAG"
sleep 1

declare -a names=("node" "tools" "polaris")
for name in "${names[@]}"
do

    tagsExpanded=()
    for tag in "${tags[@]}"; do
        tagsExpanded+=("-t aergo/$name:$tag")
    done
    echo "[aergo/$name:${tags[*]}]"
    DOCKERFILE="../Dockerfile.$name"
    [[ $name == "node" ]] && DOCKERFILE="../Dockerfile"
    echo docker build -q --build-arg GIT_TAG=$GIT_TAG ${tagsExpanded[@]} ../ --file $DOCKERFILE
    imageid=`docker build -q --build-arg GIT_TAG=$GIT_TAG ${tagsExpanded[@]} ../ --file $DOCKERFILE`
    docker images --format "Done: \t{{.Repository}}:{{.Tag}} \t{{.ID}} ({{.Size}})" | grep "${imageid:7:12}"
done

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

