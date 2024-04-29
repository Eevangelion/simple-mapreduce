#!/bin/bash
kubectl port-forward -n mapreduce service/mappers 8081:8081 & \
kubectl port-forward -n mapreduce service/reducers 8082:8082 & \
kubectl port-forward -n mapreduce service/mapreduce-master 8080:8080