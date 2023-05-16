#!/bin/bash
set -u

function check_container_exists() {
    [[ -n "$(docker ps -a -q -f name="\b${container_name}\b")" ]]
}

if [[ "${EUID}" -ne "0" ]]; then
  echo "Please run as root" >&2
  exit 1
fi

container_name='mongo-d292524a8c5494'
container_vol_name="${container_name}"

if [[ "$#" -eq "0" ]]; then

    if ! check_container_exists; then

        # The container does not exist

        docker volume create "${container_vol_name}"

        docker run -d \
            --name "${container_name}" \
            -p "${MONGO_PORT}:${MONGO_PORT}" \
            -e "MONGO_INITDB_ROOT_USERNAME=${MONGO_INITDB_ROOT_USERNAME}" \
            -e "MONGO_INITDB_ROOT_PASSWORD=${MONGO_INITDB_ROOT_PASSWORD}" \
            -v "${container_vol_name}:/data/db" \
            "mongo:${MONGO_VERSION}" && {
                echo -e "\nContainer ${container_name} created."
            } || {
                echo -e "\nFailed to create the mongo conatiner!" 1>&2
            }

    else

        # The container already exists

        docker start "${container_name}" && echo "Container ${container_name} started."

    fi

elif [[ "$1" == "stop" ]]; then

    docker stop "${container_name}" && {
            echo -e "\nContainer ${container_name} stopped."
        } || {
            echo "can not stop container ${container_name} or it does not exist."
        }

elif [[ "$1" == "remove" ]]; then

    if ! check_container_exists; then
        echo "Container ${container_name} does not exist" 1>%2
        exit 1
    fi

    answer=''
    while [[ "${answer}" != 'yes' && "${answer}" != 'no' ]]; do
        read -p "Are you sure to remove container ${container_name}? (please type 'yes' or 'no'): " answer
        echo "${answer}"
    done
    if [[ "${answer}" == 'no' ]]; then
        echo "Aborted!"
        exit 0
    fi

    docker stop "${container_name}" && docker rm "${container_name}" && {
            echo -e "\nContainer ${container_name} removed."
        } || {
            echo -e "\ncan not stop or remove container ${container_name}, it may not existing." 1>&2
        }

fi
