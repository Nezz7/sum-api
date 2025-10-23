## Kubernetes Demo

# Create a kind cluster
kind create cluster --config ./k8s/0-kind-config.yaml


# Kube-config 
kubectl config view

cat ~/.kube/config

kubectl cluster-info

kubectl get namespace

kubectl get pods -n kube-system

kubectl get pods -n kube-system -v=7

kubectl get pods -n kube-system -o wide

# Get the list of nodes
kubectl get nodes

# Get the list of pods
kubectl get pods --all-namespaces

# Set the current namespace to kube-system
kubectl config set-context --current --namespace kube-system

# Exploring the Control plane
``` bash
# Get the control plane container ID and access it
export id=$(docker ps | grep control-plane | cut -d " " -f1)
docker exec -it $id sh

# Check kubelet service status
systemctl status kubelet

# View recent kubelet logs
journalctl -u kubelet -n 100

# List all running processes
ps aux 

top 

# List all containers managed by CRI
crictl ps -a 

# View container logs
crictl logs 
```

