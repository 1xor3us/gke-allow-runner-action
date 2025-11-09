# 🔐 GKE Allow Runner Action

[![GitHub Marketplace](https://img.shields.io/badge/GitHub%20Marketplace-GKE%20Allow%20Runner-blue?logo=github)](https://github.com/marketplace/actions/gke-allow-runner-action)
[![Container](https://img.shields.io/badge/Image-ghcr.io%2F1xor3us%2Fgke--allow--runner--action-blue)](https://github.com/1xor3us/gke-allow-runner-action/pkgs/container/gke-allow-runner-action)
[![Feature: Parallel Multi-Cluster](https://img.shields.io/badge/Feature-Multi--Cluster%20Parallel-yellowgreen?logo=googlecloud)]()
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](./LICENSE)
[![Ko-fi](https://img.shields.io/badge/Ko--fi-☕-blue?logo=kofi&logoColor=white)](https://ko-fi.com/1xor3us)
[![Buy Me a Coffee](https://img.shields.io/badge/Buy%20Me%20a%20Coffee-💛-yellow?logo=buymeacoffee&logoColor=black)](https://www.buymeacoffee.com/1xor3us)
---

**GKE Allow Runner** automatically adds or removes the current GitHub Actions runner IP  to your Google Kubernetes Engine (GKE) master authorized networks.

> Automatically authorize your GitHub Action runner IP on GKE clusters — and remove it safely when your workflow ends.

> ⚡ Lightweight, statically compiled and distroless — designed for secure CI/CD environments.

> **Version:** v1.7.0 • **Signed via OIDC** • [Verify on Sigstore](https://search.sigstore.dev/?hash=sha256:6a0524531a3a860123d0ec6e9690d4fab7e10267d32c96181825a502329852f7)
---

## 🚀 Example Usage

```yaml
name: Manage GKE Runner IPs
on:
  workflow_dispatch:

jobs:
  update-gke:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Allow GitHub runner IPs on multiple GKE clusters
        uses: 1xor3us/gke-allow-runner-action@v1.7.0
        with:
          clusters: |
            europe-west1/cluster-test
            us-central1/cluster-prod
            asia-southeast1/cluster-staging
          project_id: my-gcp-project
          action: allow
          credentials_json: ${{ secrets.GCP_CREDENTIALS }}

      # 🧹 Always cleanup after
      - name: Cleanup GitHub runner IPs
        if: always()
        uses: 1xor3us/gke-allow-runner-action@v1.7.0
        with:
          clusters: |
            europe-west1/cluster-test
            us-central1/cluster-prod
            asia-southeast1/cluster-staging
          project_id: my-gcp-project
          action: cleanup
          credentials_json: ${{ secrets.GCP_CREDENTIALS }}

```
> 🧠 Now supports multiple clusters processed in parallel with automatic retries and rich structured logging.

---

## 🧩 Inputs

| Name               | Description                                                | Required             | Example                                                          |
| ------------------ | ---------------------------------------------------------- | -------------------- | ----------------------------------------------------------------  |
| `clusters`         | Multi-line list of clusters (`region/cluster-name` format) | ✅                    | `"europe-west1/cluster-test`<br>`us-central1/cluster-prod"`        |
| `project_id`       | GCP Project ID                                             | ✅                    | `my-gcp-project`                                                 |
| `action`           | Action to perform (`allow` or `cleanup`)                   | ❌ (default: `allow`) | `cleanup`                                                        |
| `credentials_json` | GCP Service Account JSON (via GitHub Secret)               | ✅                    | `${{ secrets.GCP_CREDENTIALS }}`                                 |


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

## 🐳 Why use a Docker-based Action instead of inline code?

Some GitHub Actions run inline scripts directly from the `action.yml`.  
However, **this Action intentionally uses a Docker image**, for several key reasons:

---

### 🧱 1. Reproducibility and Isolation
Running inside a container ensures a **predictable, controlled environment** — regardless of the GitHub runner OS or updates.  
All dependencies (Go binary, SDKs, system libraries) are frozen at build time.  

✅ No dependency drift  
✅ Identical behavior across all runners  
✅ Fully isolated from the host system

---

### 🧩 2. Security by Design (Distroless)
The image is built on **Google’s Distroless base**, meaning:
- No shell (`bash`, `sh`, etc.)  
- No package manager (`apt`, `apk`, etc.)  
- No root user or exec capabilities  

This drastically reduces the attack surface compared to inline scripts or normal container bases (e.g., Alpine or Ubuntu).

---

### 🏷 3. Signed and Verifiable     
Every published image is **cryptographically signed with [Sigstore Cosign](https://github.com/sigstore/cosign)** using GitHub OIDC.
> Each image is verifiable (see [Reproducible Build Verification](#-reproducible-build-verification) section.)

Users can independently verify that:
- The image was built **from this repository’s source code**  
- It has not been tampered with since publication  

This ensures full transparency and trust for anyone using the Action in production.

---

### ⚙️ 4. Faster Execution and Zero Setup
Because the compiled Go binary and GKE API client are embedded, there’s:
- No need to install `gcloud` or any SDK at runtime  
- No dependency installation step  
- Almost instant startup time  

This makes the Action ideal for ephemeral GitHub runners.

---

### 🧠 5. Optional Local Testing
Unlike pure inline Actions, Docker-based Actions can be **run and tested locally** using `docker run`,  
which is useful for debugging or validating behavior before using them in production workflows.

---

> 💡 In short: the Docker-based approach provides **security, determinism, and transparency** —  
> without sacrificing portability or speed.

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

## 🧪 Reproducible Build Verification

You can verify that the **executable inside the published image** is **bit-for-bit identical** to the one you can build yourself from this repository.  
This ensures that the code you see here is exactly what runs in the GitHub Action.

---

### 1️⃣ Clone and Build Locally
```bash
git clone https://github.com/1xor3us/gke-allow-runner-action.git
cd gke-allow-runner-action
docker build -t gke-allow-runner-action .
```

### 2️⃣ Pull the official image
Before comparison, download the signed public image from GHCR:

```bash
docker pull ghcr.io/1xor3us/gke-allow-runner-action:v1.7.0
```

---

### 3️⃣ Extract the Binary from Both Images

#### On 🪟 Windows (PowerShell)

for local build :
```bash
docker export $(docker create gke-allow-runner-action) | tar -xO main > main_local
```

for remote build :
```bash
docker export $(docker create ghcr.io/1xor3us/gke-allow-runner-action:v1.7.0) | tar -xO main > main_remote
```

#### 🐧 On Linux/macOS
for local build :
```bash
docker export $(docker create gke-allow-runner-action) | tar -xO --wildcards ./main > main_local
```
for remote build :
```bash
docker export $(docker create ghcr.io/1xor3us/gke-allow-runner-action:v1.7.0) | tar -xO --wildcards ./main > main_remote
```

---

### 4️⃣ Compare Their SHA-256 Hashes
#### 🪟 On Windows (PowerShell)
```bash
Get-FileHash main_local -Algorithm SHA256
Get-FileHash main_remote -Algorithm SHA256
```

#### 🐧 On Linux/macOS
```bash
sha256sum main_local main_remote
```

✅ If **both hashes are identical** →    
the executable inside the published GHCR image is **exactly the same** as the one you can build locally from this source code.

---

## 🔍 Verify Image Signature (CLI)

You can quickly verify the authenticity of the published image without rebuilding it:

```bash
cosign verify ghcr.io/1xor3us/gke-allow-runner-action:v1.7.0 \
  --certificate-identity-regexp "https://github.com/1xor3us/gke-allow-runner-action/.*" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com"

```

✅ If verification succeeds, it means:
- The signature was issued via GitHub OIDC (keyless   
- The image digest matches the transparency log on Sigstore/Rekor     
- The /main binary you built locally is identical to the signed one   

---

## 🧾 Verify SLSA Provenance (SLSA v1)

Every published image also includes a SLSA v1 provenance attestation, generated automatically by the [SLSA GIthub Generator](https://github.com/slsa-framework/slsa-github-generator) and signed with GitHub OIDC via [Sigstore Cosign V4](https://github.com/sigstore/cosign)

You can verify this provenance directly from GHCR using:
```bash
cosign verify-attestation \
  --type slsaprovenance1 \
  --certificate-identity-regexp ".*" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  ghcr.io/1xor3us/gke-allow-runner-action:v1.7.0
```

This confirms that the image:
- **was built from this repository** and **signed by GitHub Actions** (OIDC identity),  
- includes a **SLSA v1 provenance predicate**,  
- and is **logged in the Sigstore transparency log** for immutable traceability.  
  💡 To export the provenance for inspection:
  ```bash
  cosign verify-attestation --type slsaprovenance1 ghcr.io/1xor3us/gke-allow-runner-action:v1.7.0 \
  --certificate-identity-regexp ".*" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  --output-file provenance.json
  ```

---

## 🧩 What is SLSA Provenance?

SLSA (Supply-chain Levels for Software Artifacts) defines how to trace and verify how software was built.
The provenance includes:

- The exact commit and workflow that built the image  
- All inputs and dependencies used during the build 
- The builder identity (`GitHub Actions OIDC`)  
- Build timestamps and metadata 

Together with Sigstore signatures, this makes the action fully transparent and tamper-proof from source to artifact.

---

## 🧠 How This Proves Integrity

When you use a GitHub Action that runs inside a Docker image, you’re essentially trusting that image to do exactly what it says.    

This verification process lets anyone independently confirm that:

1️⃣ The published image on GHCR contains **exactly the same binary** built from this open-source repository.  
2️⃣ That binary has **not been altered, injected, or rebuilt** by anyone else.    
3️⃣ The image was **signed automatically by GitHub’s OIDC identity**, not manually by a private key.  
4️⃣ The signature is publicly verifiable in the [Sigstore/Rekor transparency log](https://search.sigstore.dev/).

### 🧩 In short

> Anyone can rebuild this project, extract the binary, compare its hash, and verify the signature —      
**proving that the GitHub Action is transparent, reproducible, and tamper-proof**.

This approach follows the same supply-chain security principles used by:    
- Google’s Distroless & SLSA frameworks   
- Sigstore / Cosign keyless signing   
- Chainguard Trusted Builds   

💬 `This ensures that the “code you see” is truly the “code you run”.`

---

## 🧰 Local Usage Example

You can also run the same container locally to test it.  
⚠️ You will need a **GCP Service Account JSON key** manually downloaded from your Google Cloud project.


### ⚙️ Configuration Parameters

Before running the scripts, edit the following variables according to your environment:

| Variable      | Description                           | Example                                                     |
| ------------- | ------------------------------------- | ----------------------------------------------------------- |
| `PROJECT_ID`  | Your GCP project ID                   | `my-gcp-project`                                            |
| `CLUSTERS`    | Multi-line cluster list (region/name) | `"europe-west1/cluster-test`<br>`us-central1/cluster-prod"` |
| `SA_KEY_PATH` | Path to the Service Account JSON key  | `./gcp-sa.json`                                             |
| `IMAGE_NAME`  | Docker image name                     | `gke-allow-runner-action:latest`                            |
| `ACTION`      | Operation (`allow` or `cleanup`)      | `allow`                                                     |


### 🪟 Windows (PowerShell)

1️⃣ Edit the `run_local.ps1` file to match your configuration.  
2️⃣ Run it in PowerShell:

```bash
.\run_local.ps1
```

### 🐧 Linux / macOS

1️⃣ Edit the `run_local.sh` file to match your configuration.   
2️⃣ Make it executable and run it:

```bash
chmod +x run_local.sh
./run_local.sh
```

### 🧩 Notes

- The Service Account JSON key must have the correct IAM roles (see [Required Permissions](#-required-permissions)).
- Both scripts behave exactly like the GitHub Action workflow.    
- You can test both modes:    
    - `allow` → adds your current public IP to GKE  
    - `cleanup` → removes it afterward

### 💡 Tip:
These local scripts are mainly for debugging or verifying functionality outside of GitHub Actions.  
For CI/CD workflows, prefer using the Action directly with:
```yaml
uses: 1xor3us/gke-allow-runner-action@v1.7.0
```

---

## 🧑‍💻 Author

- [![GitHub Profile](https://img.shields.io/badge/GitHub%20Profile-1xor3us-blue?logo=github)](https://github.com/1xor3us)
- [![GHCR Packages](https://img.shields.io/badge/GHCR%20Packages-gke--allow--runner--action-blue?logo=github)](https://github.com/1xor3us)

---

## ⚖️ License

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](./LICENSE)

---

## 🧭 Roadmap

- ✅ Add and remove runner IP dynamically  
- ✅ Full GKE API-only implementation (no gcloud dependency)  
- ✅ Retry mechanism when concurrent operations occur  
- ✅ Support multiple clusters in parallel  
- ✅ Advanced error handling and structured logging  
- ✅ Automatic tag bump & release workflow integration  

## 🌍 Planned Multi-Provider Support (Next Step)

| Provider | Status | Notes |
|-----------|---------|-------|
| **GKE (Google Cloud)** | ✅ Implemented | Uses native GKE API |
| **EKS (Amazon Web Services)** | 🚧 Planned | Will use AWS Go SDK v2 |
| **AKS (Microsoft Azure)** | 🚧 Planned | Will use Azure SDK for Go |

---

## 🧩 Latest Release — v1.7.0

- Parallel multi-cluster updates (3 concurrent workers)
- Automatic retry with exponential backoff
- Structured colorized logging
- PowerShell local test support
- Enhanced error handling and graceful failure mode

Stay tuned for updates on [GitHub Releases](https://github.com/1xor3us/gke-allow-runner-action/releases).

