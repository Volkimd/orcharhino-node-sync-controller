# orcharhino Node Sync Controller

A Kubernetes controller built with **Kubebuilder** that automatically synchronizes Kubernetes Nodes with the **orcharhino** host management system.

## Overview

This controller monitors the lifecycle of Nodes within a Kubernetes cluster. When a new Node is detected, the controller registers it as a host in orcharhino. When a Node is removed, it ensures the host is cleaned up.

### Key Features

* **Automated Registration:** Detects new Nodes and sends a `POST` request to the orcharhino API.
* **State Tracking:** Uses the label `orcharhino.de/synced=true` to prevent redundant API calls.
* **Mock Environment:** Includes a lightweight Go-based API mock for local development and testing without requiring a live orcharhino instance.

---

## Getting Started

### Prerequisites

* Go (v1.22+)
* Docker (for local cluster testing)
* `kubectl` and `kind` (or access to a KKP User-cluster)
* `make`

### Installation

1. **Clone the repository:**
```bash
git clone https://github.com/volkimd/orcharhino-node-sync-controller.git
cd orcharhino-node-sync-controller

```


2. **Install CRDs:**
```bash
make install

```



---

## Development & Testing

You can test the controller locally using the provided Mock API.

### 1. Start the Mock API

In a separate terminal, run the mock server:

```bash
make mock

```

The mock will start listening at `http://localhost:8080`.

### 2. Run the Controller

Start the controller locally and point it to your cluster (via `KUBECONFIG`) and the mock server:

```bash
make run-with-mock

```

### 3. Verify Synchronization

* **Check Labels:** Verify that the controller has labeled the nodes:
```bash
kubectl get nodes --show-labels | grep orcharhino

```


* **Check Mock Logs:** The mock server terminal should show:
`✅ ADD: Node 'node-name' registered.`

---

## Configuration

The controller is configured via environment variables:

| Variable | Description | Default |
| --- | --- | --- |
| `ORCHARHINO_URL` | Base URL of the orcharhino API | `http://localhost:8080` |
| `ORCHARHINO_USER` | API Username | `admin` |
| `ORCHARHINO_PASS` | API Password | `password` |

---

## Project Structure

* `cmd/main.go`: Entry point for the manager.
* `internal/controller/`: Contains the `NodeReconciler` logic.
* `test/mock_api.go`: A simple HTTP server simulating the orcharhino REST API.
* `config/`: Kubernetes manifests (CRDs, RBAC, etc.).

---

## Troubleshooting

**Connection Refused:**
If you see `dial tcp [::1]:8080: connect: connection refused`, ensure the Mock API is running in another terminal. On some systems (macOS), you may need to explicitly use `127.0.0.1` instead of `localhost`.

**Nodes not syncing:**
If a Node already has the label `orcharhino.de/synced=true`, the controller will skip it. To force a re-sync, remove the label:

```bash
kubectl label node <node-name> orcharhino.de/synced-

```

---

**Soll ich noch einen speziellen Abschnitt für die KKP-Besonderheiten (wie das Auslesen der IP-Adressen) hinzufügen?**