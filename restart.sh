#!/bin/bash
./remove_containers.sh
kubectl delete -f .
./build.sh
kubectl create -f .