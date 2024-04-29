#!/bin/bash
eval $(minikube docker-env)  
docker build -t mapreduce-mapper ./mapper && docker run --name mapreduce-mapper -p 8081:8081 -d mapreduce-mapper
docker build -t mapreduce-reducer ./reducer && docker run --name mapreduce-reducer -d -p 8082:8082 mapreduce-reducer
docker build -t mapreduce-master ./master && docker run --name mapreduce-master -d -p 8080:8080 mapreduce-master
