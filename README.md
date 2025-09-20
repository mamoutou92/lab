# NimP2P Lab
This repository presents a Kubernetes-based monitoring and logging stack for large-scale experimentation with NimP2P nodes.

The solution is **multi-tenant**: multiple experiments can run concurrently without interfering with each other. Each experiment’s metrics and logs are isolated in Grafana dashboards, thanks to a **unique experiment label** applied to every pod created by the experiment’s StatefulSet.
## Architecture and Design Choices
At a high level, the proposed solution consists of a lightweight **Golang-based CLI tool**(```nimp2p-lab```) that interacts with a **K3S cluster** (preconfigured with monitoring/logging components), allowing users to: 
  - Create new experiments
  - List active experiments and their status
  - Scale experiments (up or down)
  - Delete experiments cleanly list, scale, and delete experiments.
  The picture below illustrates the overall stack, highlighting the main components and their corresponding kubernetes objects.
### Experiment Unit (Statefulset + ClusterIP + Custom Label)
An **experiment unit** is a self-contained deployment that represents a NimP2P network.  
Each unit consists of a **StatefulSet** (to manage the pods) and a **dedicated headless service** (to handle peer discovery).  

Key aspects:  
- **Unique identity**: Every experiment has its own StatefulSet and headless ClusterIP service. Their names are unique, and the experiment name is propagated as a **custom label** to all pods.  
- **Metrics isolation**: The custom labels are later used in Prometheus, Grafana and Promtail configuration to separate metrics per experiment, enabling a true multi-tenant setup.  
- **Peer isolation**: A dedicated headless service ensures that nodes only discover peers within the same experiment, preventing cross-experiment traffic.  
- **Resource control**: CPU, memory, and bandwidth limits are applied to each pod. This allows experiments to emulate constrained environments (e.g., Raspberry Pi, smartphones, or low-bandwidth links) while keeping cluster resources balanced.  
#### Example: Creating an Experiment
The command below uses the `nimp2p-lab` tool to create an experiment named **purple** with 5 peers.  
Each peer is restricted to 5% of a CPU core, 16 MiB of RAM, and 16 Mbps uplink/downlink bandwidth.  
Peers connect to 5 discovered nodes and send messages of 1440 B every 2000 ms.
```bash
$ nimp2p-lab create --name purple --peers 5 --msg-size 1440 --msg-rate 2000 --cpu 0.05 --ram 16 --downlink-bw 16 --uplink-bw 16
[INFO] headless service 'nimp2p-exp-purple' created
[INFO] experiment 'purple' created
```
```bash
$ nimp2p-lab list
EXPERIMENT  FULLNAME           SCALE  RUNNING  STARTED AT            AGE
blue        nimp2p-exp-blue    4      4        2025-09-20T11:19:09Z  38m24s
purple      nimp2p-exp-purple  5      5        2025-09-20T11:57:09Z  24s
```
---
### Metrics Collectors and Exporters (Deployments + DaemonSets + Filters)

The monitoring stack collects **both host-level and pod-level metrics** to provide visibility into the health of the cluster and the performance of NimP2P experiments.

#### Host-Level Metrics
Host-level metrics are essential for debugging node issues and tracking aggregate resource consumption.  
These include:
- CPU and memory usage  
- Bandwidth utilization  
- Packet loss and byte errors  
To collect these, I deployed **Node Exporter** as a **DaemonSet**, ensuring that metrics of every node in the cluster can be scrapped by Prometheus.

#### Pod-Level Metrics
Pod-level metrics focus on evaluating NimP2P nodes directly. For these, I've selected:  
- Uplink and downlink data rate between peers  
- Packet drops and byte errors  
- CPU, memory, and bandwidth consumption per pod  
To export the above metrics, I deployed  **cAdvisor** as a **DaemonSet**.

#### RTT Metrics (Work in Progress)
Round-trip time (RTT) between peers is a critical metric for evaluating **network health and latency**, especially in a **GossipSub network**:  
- Messages may be forwarded across multiple peers, each hop adding extra latency.  
- Measuring RTT between peers helps detect when messages are relayed rather than delivered directly.  
- Nodes can be behind restrictive NATs, in such cases, direct peer-to-peer connections may fail due to unsupported or unsuccessful hole punching, so, traffic may be routed through a **relay node**, increasing RTT.  

Currently, RTT metrics are **not yet collected** due to time constraints.  
The plan is to implement a **custom Golang exporter** that runs as a **sidecar container** in each StatefulSet pod.

This exporter will:  
- Discover peers via the experiment’s headless service  
- Periodically probe them (e.g., via ICMP or lightweight protocols)  
- Expose a /metrics endpoint for Prometheus scrapping.  
- The **Blackbox Exporter** is insufficient for this scenario, since it cannot measure pod-to-pod RTT.

#### Prometheus Setup
Metrics are scraped by **Prometheus**, deployed on the master node as a standard Kubernetes **Deployment**.

- The configuration is stored in a **ConfigMap** (`00-setup_cluster/ConfigMaps/lab-prometheus-config.yaml`).  
- Three jobs are currently defined:  
  - Node Exporter (host-level metrics)  
  - cAdvisor (pod-level metrics)  
  - RTT Exporter (as a placeholder)  
- Filters are applied in the Prometheus config to **only scrape the `dst-lab` namespace**, avoiding unnecessary metrics collection and reducing bandwidth overhead.
---

### Log Collectors and Exporters (Deployments + DaemonSets + Filters)
- Logs are pushed to **Loki**, deployed on the master node as a standard Kubernetes **Deployment**. The Loki configuration is stored in a **ConfigMap** (`00-setup_cluster/ConfigMaps/lab-loki-config.yaml`).
- Logs are pushed to Loki by **Promtail**, deployed as a **DaemonSet**, with the host’s pod log directory (`/var/log/pods`) mounted to expose container-level logs. The Promtail configuration is stored in a **ConfigMap** (`00-setup_cluster/ConfigMaps/lab-promtail-config.yaml`). The config includes multiple relabelling (in order to be able to show logs per experiment in grafana) and uses a regex to consider pods with label `prefix=nimp2p-exp.*`
## Deploying the kubernetes cluster
