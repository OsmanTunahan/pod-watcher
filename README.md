# Pod Watcher ![version](https://img.shields.io/badge/version-1.1-blue) ![rating](https://img.shields.io/badge/rating-★★★★☆-brightgreen)


Pod Watcher is a simple program to monitor Kubernetes pods and log their events such as creation, deletion, and updates.

## Features

- Logs pod creation, deletion, and updates.
- Option to print detailed pod object information.
- Filters pods based on namespace and label selector.

## Requirements

- Go 1.22 or later
- Kubernetes cluster
- kubeconfig file for authentication (if running outside a cluster)

## Installation

1. Clone the repository:

    ```bash
    git clone https://github.com/OsmanTunahan/pod-watcher.git
    cd pod-watcher
    ```

2. Build the project:

    ```bash
    go build -o pod-watcher
    ```

## Usage

### Command-line Options

- `--kubeconfig` (string): Absolute path to the kubeconfig file (optional if running inside a cluster).
- `--namespace` (string): Namespace to watch (default: all namespaces).
- `--details` (bool): Print detailed pod object information (default: false).
- `--selector` (string): Label selector to filter pods (default: "foo=bar,baz=quux").

### Running the Program

1. Ensure you have access to your Kubernetes cluster either by running inside the cluster or by providing a valid `kubeconfig` file.
2. Run the program:

    ```bash
    ./pod-watcher --kubeconfig=~/.kube/config --namespace=default --details=true --selector="app=testk8sapp"
    ```

### Example

To watch all pods in the `default` namespace with detailed logging enabled:

```bash
./pod-watcher --namespace=default --details=true
```
