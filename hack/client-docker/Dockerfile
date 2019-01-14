FROM centos

RUN yum install -y telnet wget curl && yum update -y
COPY client /client

ENTRYPOINT ["tail", "-f", "/dev/null"]