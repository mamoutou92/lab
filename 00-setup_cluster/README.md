# ⚙️ Deploying the kubernetes cluster (K3S + Calico)

This guide Create a minimal K3s cluster with **Calico CNI**, then applies all manifests in this folder in order to create the monitoring/logging stack for `nimp2p` experiments. The monitoring/logging stack is mainly based on **Prometheus**, **Loki**, **Promtail** and **Grafana**.

##  1)  Install K3s (with flannel disabled)
1. Run following commands on the **Master Node**: 
```bash 
export INSTALL_K3S_VERSION=v1.33.4+k3s1 && curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="--flannel-backend=none --disable-network-policy --cluster-cidr=10.42.0.0/16" sh -s - --token mamoutou@2025
```
2. Get the **K3S token**:  
```bash 
sudo cat /var/lib/rancher/k3s/server/node-token
```
3. Run this on each **Worker Node**: <br />
``` export INSTALL_K3S_VERSION=v1.33.4+k3s1 && curl -sfL https://get.k3s.io | K3S_URL=https://51.91.101.28:6443 K3S_TOKEN=<TOKEN> sh -```

##  2) Install Calico
From the master node run:
```bash
sudo kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.30.3/manifests/operator-crds.yaml
sudo kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.30.3/manifests/tigera-operator.yaml
sudo kubectl create -f CustomResources/custom-resources.yaml 
```
  - Wait for all calico pods to be ready:
```bash 
sudo kubectl get pods --all-namespaces
```