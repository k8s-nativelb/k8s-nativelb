FROM nginx

RUN apt-get update && apt-get upgrade -y

COPY default.conf.template /etc/nginx/conf.d/default.conf.template
COPY server /server
COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]