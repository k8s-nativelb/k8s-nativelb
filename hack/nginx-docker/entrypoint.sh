#!/bin/bash

if [ "$1" == "nginx" ]; then
    envsubst < /etc/nginx/conf.d/default.conf.template > /etc/nginx/conf.d/default.conf && exec nginx -g 'daemon off;'
elif [ "$1" == "server" ]; then
    /server "$2"
else
    echo "no command found"
    exit 1
fi
