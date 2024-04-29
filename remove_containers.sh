#!/bin/bash
eval $(minikube docker-env)  
docker stop mapreduce-mapper
docker rm mapreduce-mapper
docker stop mapreduce-reducer
docker rm mapreduce-reducer
docker stop mapreduce-master
docker rm mapreduce-master

