#!/bin/sh

read_variable() {
    VAR=$(grep $1 $2 | xargs)
    IFS="=" read -ra VAR <<< "$VAR"
    echo ${VAR[1]}
}

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
IMAGE_TAG="chat"
ENV_FILE="${DIR}/app.env"
PORT=$(read_variable PORT "${ENV_FILE}")

docker build --no-cache -t $IMAGE_TAG .
docker run --rm --name="${IMAGE_TAG}" -v "${DIR}/log:/app/log" -p="${PORT}":"${PORT}" --env-file="${ENV_FILE}" "${IMAGE_TAG}"
