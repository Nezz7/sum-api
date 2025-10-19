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

# Set the current namespace to kub
kubectl config set-context --current --namespace kube-system

# Exploring the Control plane
export id=$(docker ps | grep control-plane | cut -d " " -f1)
docker exec -it $id sh
journalctl -u kubelet -n 100
systemctl status kubelet
top 
ps aux 
crictl ps -a 
crictl logs 

