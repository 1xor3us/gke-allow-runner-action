package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	container "google.golang.org/api/container/v1"
	"google.golang.org/api/option"
)

// ──────────────────────────────────────────────
// Logging utils
// ──────────────────────────────────────────────

var logger *slog.Logger

func init() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey || a.Key == slog.LevelKey || a.Key == slog.MessageKey {
				return slog.Attr{}
			}
			return a
		},
	})
	logger = slog.New(handler)
}

// ───────────────────────────────
// Couleurs
// ───────────────────────────────
var (
	colorInfo    = "\033[36m"      // cyan
	colorWarn    = "\033[33m"      // jaune
	colorError   = "\033[31m"      // rouge
	colorSuccess = "\033[32m"      // vert
	colorArgs    = "\033[95m"      // magenta clair
	colorCIDR    = "\033[38;5;39m" // bleu clair profond
	colorReset   = "\033[0m"
)

// ===== Couleurs (Gradient) par cluster (HSV -> ANSI 256) =====
var clusterColorMap sync.Map

// table de hash -> teinte jusqu'a [0..360)
func hashHue(s string) float64 {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
		h ^= h >> 13
		h += h << 5
		h ^= h >> 7
	}
	return float64(h % 360)
}

func hsvToRgb(h, s, v float64) (float64, float64, float64) {
	// h: 0..360, s,v: 0..1
	if s == 0 {
		return v, v, v
	}
	hh := h / 60.0
	i := math.Floor(hh)
	ff := hh - i
	p := v * (1 - s)
	q := v * (1 - (s * ff))
	t := v * (1 - (s * (1 - ff)))

	switch int(i) % 6 {
	case 0:
		return v, t, p
	case 1:
		return q, v, p
	case 2:
		return p, v, t
	case 3:
		return p, q, v
	case 4:
		return t, p, v
	default:
		return v, p, q
	}
}

func rgbToAnsi256(r, g, b float64) int {
	// map 0..1 -> 0..5 (216 cube de couleur)
	toCube := func(x float64) int {
		i := int(math.Round(x * 5))
		if i < 0 {
			i = 0
		}
		if i > 5 {
			i = 5
		}
		return i
	}
	R, G, B := toCube(r), toCube(g), toCube(b)
	return 16 + (36 * R) + (6 * G) + B
}

func ansi256(code int) string { return fmt.Sprintf("\033[38;5;%dm", code) }

// une Couleur défini par cluster (pioché dand la teinte généré)
func getClusterColor(cluster string) string {
	if cluster == "" {
		return colorArgs
	}

	if v, ok := clusterColorMap.Load(cluster); ok {
		return v.(string)
	}

	h := hashHue(cluster)
	clusterColorMap.Range(func(_, val any) bool {
		//si la couleur existe déja on l'extraie
		return true
	})

	s, v := 0.65, 0.95
	r, g, b := hsvToRgb(h, s, v)
	code := rgbToAnsi256(r, g, b)
	col := ansi256(code)
	clusterColorMap.Store(cluster, col)
	return col
}

// ───────────────────────────────
// Coloration contextuelle des actions
// ───────────────────────────────
func highlightAction(msg string) string {
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "removing"):
		return colorWrap(colorWarn, msg)
	case strings.Contains(lower, "adding"):
		return colorWrap(colorSuccess, msg)
	case strings.Contains(lower, "update"): // suffit pour update & updating
		return colorWrap(colorInfo, msg)
	case strings.Contains(lower, "waiting"):
		return colorWrap("\033[38;5;208m", msg)
	case strings.Contains(lower, "new authorized networks"):
		return colorWrap("\033[38;5;117m", msg)
	case strings.Contains(lower, "operation completed"), strings.Contains(lower, "updated successfully"):
		return colorWrap(colorSuccess, msg)
	case strings.Contains(lower, "processed successfully"):
		return colorWrap(colorSuccess, msg)
	default:
		return msg
	}
}

func colorWrap(color, msg string) string {
	return fmt.Sprintf("%s%s%s", color, msg, colorReset)
}

// ───────────────────────────────
// Colorisation des arguments
// ───────────────────────────────
func colorizeArgs(s string) string {
	replacements := []struct {
		re   *regexp.Regexp
		desc string
	}{
		{regexp.MustCompile(`(/\d{1,2})\b`), "cidr"},
		{regexp.MustCompile(`(?:^|[^0-9/])(\d{1,3}(?:\.\d{1,3}){3})(?:[^0-9/]|$)`), "ip"},
		{regexp.MustCompile(`(?i)(?:europe|us|asia)-[a-z0-9-]+`), "region"},
		{regexp.MustCompile(`(?:^|[^A-Za-z0-9-./\s])(\d+)(?:[^A-Za-z0-9-.\s]|$)`), "number"},
		{regexp.MustCompile(`([\(\)])`), "paren"},
	}

	ansiRe := regexp.MustCompile(`\033\[[0-9;]*m`)

	for _, r := range replacements {
		s = r.re.ReplaceAllStringFunc(s, func(m string) string {
			if ansiRe.MatchString(m) || strings.Contains(m, "\033[") || strings.Contains(m, ";") {
				return m
			}
			sub := r.re.FindStringSubmatch(m)
			if len(sub) > 1 {
				switch r.desc {
				case "cidr":
					return strings.Replace(m, sub[1], colorCIDR+sub[1]+colorReset, 1)
				case "paren":
					return strings.Replace(m, sub[1], colorReset+sub[1]+colorReset, 1)
				default:
					return strings.Replace(m, sub[1], colorArgs+sub[1]+colorReset, 1)
				}
			}
			if r.desc == "cidr" {
				return colorCIDR + m + colorReset
			}
			return colorArgs + m + colorReset
		})
	}
	return s
}

// ───────────────────────────────
// Logger générique avec couleur
// ───────────────────────────────
func logGeneric(level string, color string, msg string, err error, clusterCtx string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	coloredArgs := colorizeArgs(formatted)
	coloredMsg := highlightAction(coloredArgs)

	prefix := ""
	if clusterCtx != "" {
		prefixColor := getClusterColor(clusterCtx) // <- gradient color here
		prefix = fmt.Sprintf("[%s%s%s] ", prefixColor, clusterCtx, colorReset)
	}

	fmt.Printf("%s%s:%s %s%s", color, strings.ToUpper(level), colorReset, prefix, coloredMsg)
	if err != nil {
		fmt.Printf(" %serror=%v%s", colorArgs, err, colorReset)
	}
	fmt.Println()
}

// ───────────────────────────────
// Extraction automatique du contexte cluster
// ───────────────────────────────
func extractLogArgs(args ...any) (cluster string, msg string, msgArgs []any) {
	if len(args) == 0 {
		return "", "", nil
	}

	if s, ok := args[0].(string); ok && !regexp.MustCompile(`^(cluster|[a-z]+-[a-z0-9-]+)$`).MatchString(s) {
		return "", s, args[1:]
	}

	if len(args) >= 2 {
		if s, ok := args[0].(string); ok {
			return s, args[1].(string), args[2:]
		}
	}
	return "", "", nil
}

// ───────────────────────────────
// Fonctions wrapper simplifiées
// ───────────────────────────────
func logInfo(args ...any) {
	cluster, msg, msgArgs := extractLogArgs(args...)
	logGeneric("ℹ️ INFO", colorInfo, msg, nil, cluster, msgArgs...)
}

func logWarn(args ...any) {
	cluster, msg, msgArgs := extractLogArgs(args...)
	logGeneric("⚠️ WARN", colorWarn, msg, nil, cluster, msgArgs...)
}

func logError(args ...any) {
	cluster, msg, msgArgs := extractLogArgs(args...)
	var err error
	if len(msgArgs) > 0 {
		if e, ok := msgArgs[0].(error); ok {
			err = e
			msgArgs = msgArgs[1:]
		}
	}
	logGeneric("❌ ERROR", colorError, msg, err, cluster, msgArgs...)
}

func logSuccess(args ...any) {
	cluster, msg, msgArgs := extractLogArgs(args...)
	logGeneric("✅ SUCCESS", colorSuccess, msg, nil, cluster, msgArgs...)
}

func fatal(args ...any) {
	cluster, msg, msgArgs := extractLogArgs(args...)
	var err error
	if len(msgArgs) > 0 {
		if e, ok := msgArgs[0].(error); ok {
			err = e
			msgArgs = msgArgs[1:]
		}
	}
	logGeneric("❌ FATAL", colorError, msg, err, cluster, msgArgs...)
	os.Exit(1)
}

// ──────────────────────────────────────────────
// GKE Utils
// ──────────────────────────────────────────────

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

func waitOperation(ctx context.Context, svc *container.Service, project, region, opName, clusterName string) {
	logInfo(clusterName, "Waiting for operation to complete")
	for {
		op, err := svc.Projects.Locations.Operations.Get(
			fmt.Sprintf("projects/%s/locations/%s/operations/%s", project, region, opName),
		).Do()
		if err != nil {
			fatal(clusterName, "Failed to fetch operation status", err)
		}

		if op.Status == "DONE" {
			if op.Error != nil {
				logError(clusterName, "Operation failed", fmt.Errorf("%s", op.Error.Message))
				for _, d := range op.Error.Details {
					logger.Warn("Detail", "info", string(d))
				}
				os.Exit(1)
			}
			logSuccess(clusterName, "Operation completed")
			return
		}
		time.Sleep(2 * time.Second)
	}
}

func retryClusterUpdate(ctx context.Context, svc *container.Service, clusterPath, project, region string, updateReq *container.UpdateClusterRequest, clusterName string) (*container.Operation, error) {
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		op, err := svc.Projects.Locations.Clusters.Update(clusterPath, updateReq).Do()
		if err == nil {
			if i > 0 {
				logSuccess(clusterName, "Retry succeeded after %d attempt(s)", i+1)
			}
			return op, nil
		}

		msg := err.Error()
		if strings.Contains(msg, "operation in progress") ||
			strings.Contains(msg, "another operation") ||
			strings.Contains(msg, "CLUSTER_ALREADY_HAS_OPERATION") ||
			strings.Contains(msg, "failedPrecondition") ||
			strings.Contains(msg, "409") {

			wait := time.Duration(20*(i+1)) * time.Second
			logWarn(clusterName, "Concurrent operation detected, retrying in %v...", wait)
			time.Sleep(wait)
			continue
		}
		return nil, err
	}
	return nil, fmt.Errorf("max retries reached while waiting for concurrent operations to finish")
}

// ──────────────────────────────────────────────
// Main logic
// ──────────────────────────────────────────────

func main() {
	ctx := context.Background()

	if raw := os.Getenv("INPUT_CREDENTIALS_JSON"); raw != "" {
		tmpFile, err := os.CreateTemp("/tmp", "gcp-creds-*.json")
		if err != nil {
			fatal("Failed to create temp credentials file", err)
		}
		defer tmpFile.Close()

		if _, err := tmpFile.Write([]byte(raw)); err != nil {
			fatal("Failed to write credentials file", err)
		}

		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tmpFile.Name())
	}

	clustersRaw := getenv("INPUT_CLUSTERS", "")
	if clustersRaw == "" {
		fatal("Missing INPUT_CLUSTERS", nil)
	}

	lines := strings.Split(strings.TrimSpace(clustersRaw), "\n")
	var clusters []struct {
		Region string
		Name   string
	}
	for _, line := range lines {
		parts := strings.Split(strings.TrimSpace(line), "/")
		if len(parts) != 2 {
			fatal("Invalid cluster entry: %s (expected format: region/cluster-name)", nil, line)
		}
		clusters = append(clusters, struct {
			Region string
			Name   string
		}{Region: parts[0], Name: parts[1]})
	}

	project := getenv("INPUT_PROJECT_ID", "")
	creds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	action := getenv("INPUT_ACTION", "allow")

	opts := []option.ClientOption{}
	if creds != "" {
		logInfo("Using service account at %s", creds)
		opts = append(opts, option.WithCredentialsFile(creds))
	} else {
		logWarn("No GOOGLE_APPLICATION_CREDENTIALS provided — using default ADC context.")
	}

	svc, err := container.NewService(ctx, opts...)
	if err != nil {
		fatal("Failed to create GKE service", err)
	}

	logInfo("📡 Fetching public IP...")
	ip, err := fetchPublicIP()
	if err != nil {
		fatal("Failed to get public IP", err)
	}
	logInfo("🌐 Runner IP detected: %s", ip)
	logInfo("⚙️ Starting multi-cluster update process (%d clusters)", len(clusters))

	var wg sync.WaitGroup
	maxParallel := 3
	sem := make(chan struct{}, maxParallel)

	for _, c := range clusters {
		wg.Add(1)
		go func(region, clusterName string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			logInfo(clusterName, "Processing cluster (%s)", region)
			clusterPath := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, region, clusterName)
			cluster, err := svc.Projects.Locations.Clusters.Get(clusterPath).Do()
			if err != nil {
				logError(clusterName, "Failed to get cluster info", err)
				return
			}

			currentCfg := cluster.MasterAuthorizedNetworksConfig
			var currentCIDRs []string
			if currentCfg != nil {
				for _, cidr := range currentCfg.CidrBlocks {
					currentCIDRs = append(currentCIDRs, cidr.CidrBlock)
				}
			}

			newCIDRs := map[string]bool{}
			for _, c := range currentCIDRs {
				newCIDRs[c] = true
			}

			runnerCIDR := ip + "/32"
			if action == "cleanup" {
				logInfo(clusterName, "🧹 Removing runner IP")
				delete(newCIDRs, runnerCIDR)
			} else {
				logInfo(clusterName, "➕ Adding runner IP")
				newCIDRs[runnerCIDR] = true
			}

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
			logInfo(clusterName, "New authorized networks:")
			fmt.Println()
			indented := strings.ReplaceAll(string(data), "\n", "\n  ")
			fmt.Println("  " + indented)

			logInfo(clusterName, "🚀 Updating cluster configuration")
			updateReq := &container.UpdateClusterRequest{
				Update: &container.ClusterUpdate{
					DesiredMasterAuthorizedNetworksConfig: finalCfg,
				},
			}

			op, err := retryClusterUpdate(ctx, svc, clusterPath, project, region, updateReq, clusterName)
			if err != nil {
				logError(clusterName, "Failed to update cluster", err)
				return
			}

			logInfo(clusterName, "Update request submitted")
			waitOperation(ctx, svc, project, region, op.Name, clusterName)
			logSuccess(clusterName, "Cluster Updated Successfully")

		}(c.Region, c.Name)
	}

	wg.Wait()
	logSuccess("All Clusters Processed Successfully (%d total)", len(clusters))
}
