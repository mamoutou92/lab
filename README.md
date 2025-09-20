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

## Deploying the kubernetes cluster
