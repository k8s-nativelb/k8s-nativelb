#!/bin/bash

echo 1 > /proc/sys/net/ipv4/ip_nonlocal_bind
echo 1 > /var/proc/sys/net/ipv4/ip_nonlocal_bind

./nativelb-agent