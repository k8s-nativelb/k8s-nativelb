#!/bin/bash

kubectl get pod --all-namespaces -o wide

kubectl get events --all-namespaces

kubectl get nodes -o wide
