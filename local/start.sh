#!/bin/bash

kind create cluster --name path-localnet --config ./local/kind-config.yaml
kubectl config use-context kind-path-localnet
kubectl create namespace path
kubectl config set-context --current --namespace=path
kubectl create secret generic path-config --from-file=./local/path/.config.yaml -n path
tilt up