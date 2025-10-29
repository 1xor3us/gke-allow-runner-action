package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	container "google.golang.org/api/container/v1"
	"google.golang.org/api/option"
)

func fetchPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return strings.TrimSpace(string(body)), nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func waitOperation(ctx context.Context, svc *container.Service, project, region, opName string) {
	fmt.Print("Waiting for operation to complete")
	for {
		op, err := svc.Projects.Locations.Operations.Get(
			fmt.Sprintf("projects/%s/locations/%s/operations/%s", project, region, opName),
		).Do()
		if err != nil {
			fmt.Println("\nFailed to fetch operation status:", err)
			os.Exit(1)
		}

		if op.Status == "DONE" {
			fmt.Println("\nOperation completed.")
			return
		}

		fmt.Print(".")
		time.Sleep(2 * time.Second)
	}
}


func main() {
	ctx := context.Background()

	// ──────────────────────────────────────────────
	// Inputs GitHub Action (env vars)
	// ──────────────────────────────────────────────
	// Si credentials_json est passé via INPUT_CREDENTIALS_JSON, on l'écrit dans /tmp/sa.json
	if raw := os.Getenv("INPUT_CREDENTIALS_JSON"); raw != "" {
		tmpFile, err := os.CreateTemp("/tmp", "gcp-creds-*.json")
		if err != nil {
			fmt.Println("❌ Failed to create temp credentials file:", err)
			os.Exit(1)
		}
		defer tmpFile.Close()

		if _, err := tmpFile.Write([]byte(raw)); err != nil {
			fmt.Println("❌ Failed to write credentials file:", err)
			os.Exit(1)
		}

		// On force la variable d'environnement pour le SDK GCP
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tmpFile.Name())
	}


	action := getenv("INPUT_ACTION", "allow")
	clusterName := getenv("INPUT_CLUSTER_NAME", "")
	region := getenv("INPUT_REGION", "")
	project := getenv("INPUT_PROJECT_ID", "")
	creds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	if clusterName == "" || region == "" || project == "" {
		fmt.Println("Missing required environment variables (cluster/region/project).")
		os.Exit(1)
	}

	// ──────────────────────────────────────────────
	// Authentification GCP via JSON
	// ──────────────────────────────────────────────
	opts := []option.ClientOption{}
	if creds != "" {
		fmt.Println("Using service account:", creds)
		opts = append(opts, option.WithCredentialsFile(creds))
	} else {
		fmt.Println("No GOOGLE_APPLICATION_CREDENTIALS provided — using default ADC context.")
	}

	svc, err := container.NewService(ctx, opts...)
	if err != nil {
		fmt.Println("Failed to create GKE service:", err)
		os.Exit(1)
	}

	// ──────────────────────────────────────────────
	// Étape 1 : Récupération IP publique du runner
	// ──────────────────────────────────────────────
	fmt.Println("📡 Fetching public IP...")
	ip, err := fetchPublicIP()
	if err != nil {
		fmt.Println("Failed to get public IP:", err)
		os.Exit(1)
	}
	fmt.Println("🌐 Runner IP detected:", ip)

	// ──────────────────────────────────────────────
	// Étape 2 : Lecture du cluster existant
	// ──────────────────────────────────────────────
	fmt.Println("Fetching current authorized networks...")
	clusterPath := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, region, clusterName)

	cluster, err := svc.Projects.Locations.Clusters.Get(clusterPath).Do()
	if err != nil {
		fmt.Println("Failed to get cluster info:", err)
		os.Exit(1)
	}

	currentCfg := cluster.MasterAuthorizedNetworksConfig
	var currentCIDRs []string
	if currentCfg != nil {
		for _, cidr := range currentCfg.CidrBlocks {
			currentCIDRs = append(currentCIDRs, cidr.CidrBlock)
		}
	}

	if len(currentCIDRs) == 0 {
		fmt.Println("ℹNo authorized IPs found — initializing with runner IP.")
	}

	// ──────────────────────────────────────────────
	// Étape 3 : Ajout ou suppression de l’IP
	// ──────────────────────────────────────────────
	newCIDRs := map[string]bool{}
	for _, c := range currentCIDRs {
		newCIDRs[c] = true
	}

	runnerCIDR := ip + "/32"
	if action == "cleanup" {
		fmt.Println("Removing runner IP...")
		delete(newCIDRs, runnerCIDR)
	} else {
		fmt.Println("Adding runner IP...")
		newCIDRs[runnerCIDR] = true
	}

	// ──────────────────────────────────────────────
	// Étape 4 : Construction de la nouvelle config
	// ──────────────────────────────────────────────
	finalCfg := &container.MasterAuthorizedNetworksConfig{
		Enabled:    true,
		CidrBlocks: []*container.CidrBlock{},
	}

	for cidr := range newCIDRs {
		finalCfg.CidrBlocks = append(finalCfg.CidrBlocks, &container.CidrBlock{CidrBlock: cidr})
	}
	slices.SortFunc(finalCfg.CidrBlocks, func(a, b *container.CidrBlock) int {
		return strings.Compare(a.CidrBlock, b.CidrBlock)
	})

	data, _ := json.MarshalIndent(finalCfg, "", "  ")
	fmt.Println("New authorized networks:\n", string(data))

	// ──────────────────────────────────────────────
	// Étape 5 : Mise à jour du cluster
	// ──────────────────────────────────────────────
	fmt.Println("🚀 Updating cluster configuration...")
	updateReq := &container.UpdateClusterRequest{
		Update: &container.ClusterUpdate{
			DesiredMasterAuthorizedNetworksConfig: finalCfg,
		},
	}

	op, err := svc.Projects.Locations.Clusters.Update(clusterPath, updateReq).Do()
	if err != nil {
		fmt.Println("Failed to update cluster:", err)
		os.Exit(1)
	}

	fmt.Printf("Update request submitted. Operation: %s\n", op.Name)
	waitOperation(ctx, svc, project, region, op.Name)
	fmt.Println("Operation completed successfully for action:", action)
}
