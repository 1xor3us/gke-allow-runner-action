# 🔐 GKE Allow Runner Action

[![GitHub Marketplace](https://img.shields.io/badge/GitHub%20Marketplace-GKE%20Allow%20Runner-blue?logo=github)](https://github.com/marketplace/actions/gke-allow-runner)
[![Container](https://img.shields.io/badge/Image-ghcr.io%2F1xor3us%2Fgke--allow--runner--action-blue)](https://github.com/1xor3us/gke-allow-runner-action/pkgs/container/gke-allow-runner-action)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](./LICENSE)
[![Build](https://img.shields.io/github/actions/workflow/status/1xor3us/gke-allow-runner-action/release.yml?label=Build)](https://github.com/1xor3us/gke-allow-runner-action/actions)
[![Version](https://img.shields.io/github/v/release/1xor3us/gke-allow-runner-action?logo=github)](https://github.com/1xor3us/gke-allow-runner-action/releases)

---

**GKE Allow Runner** automatically adds or removes the current GitHub Actions runner IP  to your Google Kubernetes Engine (GKE) master authorized networks.

> Automatically authorize your GitHub Action runner IP on GKE clusters — and remove it safely when your workflow ends.

> ⚡ Lightweight, statically compiled and distroless — designed for secure CI/CD environments.

---

## 🚀 Example Usage

```yaml
name: Manage GKE Runner IP
on:
  workflow_dispatch:

jobs:
  update-gke:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Allow GitHub runner IP on GKE
        uses: 1xor3us/gke-allow-runner-action@v1.0.0
        with:
          cluster_name: cluster01
          region: europe-west1
          project_id: reliable-plasma-476112-q3
          action: allow
          credentials_json: ${{ secrets.GCP_CREDENTIALS }}
     
     # do what you have to do ...
     # ALWAYS CLEAN UP IN THE END OF THE WORKFLOWS

      - name: Cleanup GitHub runner IP
        if: always()
        uses: 1xor3us/gke-allow-runner-action@v1.0.0
        with:
          cluster_name: cluster01
          region: europe-west1
          project_id: reliable-plasma-476112-q3
          action: cleanup
          credentials_json: ${{ secrets.GCP_CREDENTIALS }}
```

---

## 🧩 Inputs

| Name               | Description                                  | Required             | Example                          |
| ------------------ | -------------------------------------------- | -------------------- | -------------------------------- |
| `cluster_name`     | Target GKE cluster name                      | ✅                    | `cluster-test`                   |
| `region`           | GCP region                                   | ✅                    | `europe-west1`                   |
| `project_id`       | GCP project ID                               | ✅                    | `my-gcp-project`                 |
| `action`           | Action to perform (`allow` or `cleanup`)     | ❌ (default: `allow`) | `cleanup`                        |
| `credentials_json` | GCP Service Account JSON (via GitHub Secret) | ✅                    | `${{ secrets.GCP_CREDENTIALS }}` |

---

## 🔑 Required Permissions

### The provided Service Account must have the following IAM roles:

- `roles/container.admin`

- `roles/container.clusterViewer`

- `roles/container.clusterAdmin`

### Example of assigning permissions:

```bash
gcloud projects add-iam-policy-binding <PROJECT_ID> \
  --member="serviceAccount:<SERVICE_ACCOUNT_EMAIL>" \
  --role="roles/container.admin"
```

---

## 🧰 Technical details

| Property       | Description                                           |
| -------------- | ----------------------------------------------------- |
| **Language**   | Go (statically compiled)                              |
| **Base Image** | Distroless Debian 12                                  |
| **SDK/API**    | Google GKE API (`google.golang.org/api/container/v1`) |
| **Image Size** | ~65 MB                                                |
| **Runtime**    | Instant execution (no gcloud CLI needed)              |
| **Security**   | No shell, no package manager — minimal attack surface |

---

### 💡 Why this Action?

Traditional GitHub runners have dynamic public IPs.

This action automates the process of allowing that IP in your GKE cluster’s Master Authorized Networks,
and then removes it once the job is complete — ensuring **zero stale IP exposure**.

✅ Fully automated

✅ Secure & stateless

✅ Instant API-based updates

✅ Works with ephemeral GitHub runners

---

## 🔒 Security Recommendations

- Use a dedicated Service Account with **minimal GKE privileges**.

- Keep your `GCP_CREDENTIALS` secret stored securely in GitHub Secrets.

- Always clean up the runner IP (action: cleanup) at the end of your job.

---

## 🧑‍💻 Author

- [![GitHub Profile](https://img.shields.io/badge/GitHub%20Profile-1xor3us-blue?logo=github)](https://github.com/1xor3us)
- [![GHCR Packages](https://img.shields.io/badge/GHCR%20Packages-gke--allow--runner--action-blue?logo=github)](https://github.com/1xor3us)

---

## ⚖️ License

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](./LICENSE)

---

## 🧭 Roadmap

- [x] Add and remove runner IP dynamically  
- [x] Full GKE API-only implementation (no gcloud dependency)  
- [ ] Retry mechanism when concurrent operations occur  
- [ ] Support multiple clusters in parallel  
- [ ] Advanced error handling and structured logging  
- [ ] Automatic tag bump & release workflow integration  

## 🌍 Planned Multi-Provider Support

| Provider | Status | Notes |
|-----------|---------|-------|
| **GKE (Google Cloud)** | ✅ Implemented | Uses native GKE API |
| **EKS (Amazon Web Services)** | 🚧 Planned | Will use AWS Go SDK v2 |
| **AKS (Microsoft Azure)** | 🚧 Planned | Will use Azure SDK for Go |

Stay tuned for updates on [GitHub Releases](https://github.com/1xor3us/gke-allow-runner-action/releases).

