#!/bin/bash
set -euo pipefail
IFS=$'\n'

function finish {
    popd
}

# set cmd to $1 or an empty value if not set
cmd=${1:-}

if [ "$cmd" == "start" ]
then
    pushd ../..
    trap finish EXIT

    # build juggler for linux-amd64
    GOOS=linux GOARCH=amd64 make

    # create the redis droplet
    doctl compute droplet create \
        juggler-redis \
        --image redis \
        --region ${JUGGLER_DO_REGION} \
        --size ${JUGGLER_DO_SIZE} \
        --ssh-keys ${JUGGLER_DO_SSHKEY} \
        --wait

    # create the client, callee and server droplet
    doctl compute droplet create \
        juggler-server juggler-callee juggler-load \
        --image ubuntu-14-04-x64 \
        --region ${JUGGLER_DO_REGION} \
        --size ${JUGGLER_DO_SIZE} \
        --ssh-keys ${JUGGLER_DO_SSHKEY} \
        --wait

    # start redis on the expected port and with the right config
    dropletname=juggler-redis
    getip='doctl compute droplet list --format PublicIPv4 --no-header ${dropletname} | head -n 1'
    redisip=$(eval ${getip})
    echo "redis IP: " ${redisip}
    ssh-keygen -R ${redisip}
    ssh -n -f -oStrictHostKeyChecking=no root@${redisip} "sh -c 'pkill redis-server; echo 511 > /proc/sys/net/core/somaxconn; nohup redis-server --port 7000 --maxclients 100000 > /dev/null 2>&1 &'"

    # copy the server to juggler-server
    dropletname=juggler-server
    serverip=$(eval ${getip})
    echo "server IP: " ${serverip}
    ssh-keygen -R ${serverip}
    scp -C -oStrictHostKeyChecking=no juggler-server root@${serverip}:~
    ssh -n -f root@${serverip} "sh -c 'nohup ~/juggler-server -L -redis=${redisip}:7000'"

    # copy the callee to juggler-callee
    dropletname=juggler-callee
    calleeip=$(eval ${getip})
    echo "callee IP: " ${calleeip}
    ssh-keygen -R ${calleeip}
    scp -C -oStrictHostKeyChecking=no juggler-callee root@${calleeip}:~
    ssh -n -f root@${calleeip} "sh -c 'nohup ~/juggler-callee -redis=${redisip}:7000'"

    # copy the load tool to juggler-load
    dropletname=juggler-load
    loadip=$(eval ${getip})
    echo "load IP: " ${loadip}
    ssh-keygen -R ${loadip}
    scp -C -oStrictHostKeyChecking=no juggler-load root@${loadip}:~

    exit 0
fi

if [ "$cmd" == "stop" ]
then
    ids=$(doctl compute droplet list juggler-* --no-header --format ID)
    for id in ${ids}; do
        doctl compute droplet delete ${id}
    done
    exit 0
fi

echo "Usage: $0 [start|stop]"
echo
echo "start       -- Launch droplets and run load test."
echo "               WARNING: will charge money!"
echo "stop        -- Destroy droplets."
echo "               WARNING: destroys droplets by name!"
echo
