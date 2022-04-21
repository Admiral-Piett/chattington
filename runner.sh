#!/bin/sh

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
IMAGE_TAG="chat"
ENV_FILE="${DIR}/app.env"

read_variable() {
    VAR=$(grep $1 $2 | xargs)
    IFS="=" read -ra VAR <<< "$VAR"
    echo ${VAR[1]}
}

post_process(){
  echo "stopping ${IMAGE_TAG}"
  docker stop $IMAGE_TAG
}
trap post_process SIGINT

PORT=$(read_variable PORT "${ENV_FILE}")

docker build --no-cache -t $IMAGE_TAG .
docker run --rm -d --name="${IMAGE_TAG}" -v "${DIR}/log:/app/log" -p="${PORT}":"${PORT}" --env-file="${ENV_FILE}" "${IMAGE_TAG}"

tail -F "${DIR}/log/chat.log"
