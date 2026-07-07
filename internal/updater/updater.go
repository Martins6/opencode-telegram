package updater

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	RepoOwner   = "martins6"
	RepoName    = "acolyte"
	BinaryName  = "acolyte"
	APIBase     = "https://api.github.com"
	checksumsFn = "checksums.txt"

	defaultAPITimeout = 30 * time.Second
	defaultDownloadTO = 2 * time.Minute
)

var Version = "dev"

var (
	ErrUnsupportedPlatform = errors.New("unsupported platform: no release asset for this OS/ARCH")
	ErrAssetNotFound       = errors.New("release asset not found for platform")
	ErrChecksumMissing     = errors.New("checksum entry missing for asset")
	ErrChecksumMismatch    = errors.New("checksum verification failed")
	ErrDevBuild            = errors.New("dev build; update check skipped")
	ErrReleaseNotFound     = errors.New("release not found")
)

type Options struct {
	TargetVersion   string
	AssetOverride   string
	ExePathOverride string
	SkipChecksum    bool
	Restart         bool
	RestartArgs     []string
	SignMac         bool
	APIBase         string
	HTTPClient      *http.Client
	APITimeout      time.Duration
	DownloadTimeout time.Duration
}

type Release struct {
	Tag                string
	AssetURL           string
	BrowserDownloadURL string
	ArchiveName        string
	ChecksumsURL       string
	Digest             string
}

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Name    string        `json:"name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	URL                string `json:"url"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Digest             string `json:"digest"`
}

func Current() string { return Version }

func OS() string {
	switch runtime.GOOS {
	case "linux":
		return "linux"
	case "darwin":
		return "darwin"
	default:
		return runtime.GOOS
	}
}

func Arch() string {
	switch runtime.GOARCH {
	case "amd64", "x86_64":
		return "amd64"
	case "arm64", "aarch64":
		return "arm64"
	default:
		return runtime.GOARCH
	}
}

func AssetName() (string, error) {
	osName, archName := OS(), Arch()
	if (osName != "linux" && osName != "darwin") || (archName != "amd64" && archName != "arm64") {
		return "", fmt.Errorf("%w: %s/%s", ErrUnsupportedPlatform, runtime.GOOS, runtime.GOARCH)
	}
	return fmt.Sprintf("%s_%s_%s.tar.gz", BinaryName, osName, archName), nil
}

func (o Options) httpClient() *http.Client {
	if o.HTTPClient != nil {
		return o.HTTPClient
	}
	timeout := o.APITimeout
	if timeout == 0 {
		timeout = defaultAPITimeout
	}
	return &http.Client{Timeout: timeout}
}

func (o Options) apiBase() string {
	if o.APIBase != "" {
		return o.APIBase
	}
	return APIBase
}

func (o Options) downloadTimeout() time.Duration {
	if o.DownloadTimeout != 0 {
		return o.DownloadTimeout
	}
	return defaultDownloadTO
}

func Latest(ctx context.Context, opts Options) (*Release, error) {
	name, err := AssetName()
	if err != nil {
		return nil, err
	}

	rel, err := fetchRelease(ctx, opts, opts.TargetVersion)
	if err != nil {
		return nil, err
	}

	var (
		matched      *githubAsset
		checksumsURL string
		available    []string
	)
	for i := range rel.Assets {
		asset := &rel.Assets[i]
		available = append(available, asset.Name)
		if asset.Name == checksumsFn && checksumsURL == "" {
			checksumsURL = asset.BrowserDownloadURL
		}
		if matched == nil && asset.Name == name {
			a := *asset
			matched = &a
		}
	}
	if matched == nil {
		return nil, fmt.Errorf("%w: expected %s in release %s; available assets: %s",
			ErrAssetNotFound, name, rel.TagName, strings.Join(available, ", "))
	}

	return &Release{
		Tag:                rel.TagName,
		AssetURL:           matched.URL,
		BrowserDownloadURL: matched.BrowserDownloadURL,
		ArchiveName:        matched.Name,
		ChecksumsURL:       checksumsURL,
		Digest:             matched.Digest,
	}, nil
}

func fetchRelease(ctx context.Context, opts Options, version string) (*githubRelease, error) {
	var url string
	if version == "" {
		url = fmt.Sprintf("%s/repos/%s/%s/releases/latest", opts.apiBase(), RepoOwner, RepoName)
	} else {
		url = fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", opts.apiBase(), RepoOwner, RepoName, version)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", BinaryName, Current()))

	resp, err := opts.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("github api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrReleaseNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("github api returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}
	if rel.TagName == "" {
		return nil, ErrReleaseNotFound
	}
	return &rel, nil
}

func IsNewer(latest, current string) (bool, error) {
	if current == "" || current == "dev" {
		return false, nil
	}
	if latest == current {
		return false, nil
	}
	return CompareSemver(stripV(latest), stripV(current)) > 0, nil
}

func IsOutdated(ctx context.Context, opts Options) (string, bool, error) {
	if Current() == "dev" {
		return "", false, ErrDevBuild
	}
	rel, err := Latest(ctx, opts)
	if err != nil {
		return "", false, err
	}
	yes, err := IsNewer(rel.Tag, Current())
	if err != nil {
		return rel.Tag, false, err
	}
	return rel.Tag, yes, nil
}

func Apply(ctx context.Context, rel *Release, opts Options) error {
	if rel == nil {
		return errors.New("release is nil")
	}

	exePath, err := resolveExePath(opts.ExePathOverride)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "acolyte-update-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, rel.ArchiveName)
	if err := downloadFile(ctx, opts, rel.BrowserDownloadURL, archivePath); err != nil {
		return fmt.Errorf("download archive: %w", err)
	}

	if !opts.SkipChecksum {
		if err := verifyChecksum(ctx, opts, rel, archivePath); err != nil {
			return err
		}
	}

	newBinary, err := extractBinary(archivePath, tmpDir)
	if err != nil {
		return fmt.Errorf("extract binary: %w", err)
	}

	if err := os.Chmod(newBinary, 0o755); err != nil {
		return fmt.Errorf("chmod new binary: %w", err)
	}

	if err := replaceExe(exePath, newBinary); err != nil {
		return err
	}

	if OS() == "darwin" && opts.SignMac {
		if _, err := exec.LookPath("codesign"); err == nil {
			sign := exec.Command("codesign", "--force", "--sign", "-", "--deep", exePath)
			sign.Stdout = os.Stderr
			sign.Stderr = os.Stderr
			_ = sign.Run()
		}
	}

	if opts.Restart {
		cmd := exec.Command(exePath, opts.RestartArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("start new binary: %w", err)
		}
		os.Exit(0)
	}
	return nil
}

func resolveExePath(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	p, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate executable: %w", err)
	}
	return p, nil
}

func downloadFile(ctx context.Context, opts Options, url, dst string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", BinaryName, Current()))

	client := opts.httpClient()
	if client == nil {
		client = &http.Client{Timeout: opts.downloadTimeout()}
	} else if opts.downloadTimeout() > 0 {
		client.Timeout = opts.downloadTimeout()
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download %s: status %d", url, resp.StatusCode)
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	return nil
}

func parseChecksums(data []byte) map[string]string {
	out := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		hash := parts[0]
		name := strings.TrimPrefix(parts[len(parts)-1], "*")
		if strings.HasPrefix(hash, "*") {
			hash = strings.TrimPrefix(hash, "*")
		}
		out[name] = hash
	}
	return out
}

func verifyChecksum(ctx context.Context, opts Options, rel *Release, archivePath string) error {
	if rel.ChecksumsURL == "" {
		return fmt.Errorf("%w: no checksums.txt asset attached to release %s", ErrChecksumMissing, rel.Tag)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rel.ChecksumsURL, nil)
	if err != nil {
		return err
	}
	client := opts.httpClient()
	if client == nil {
		client = &http.Client{Timeout: opts.downloadTimeout()}
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download checksums: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download checksums: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	checksums := parseChecksums(body)
	want, ok := checksums[rel.ArchiveName]
	if !ok {
		return fmt.Errorf("%w: %s not in checksums.txt", ErrChecksumMissing, rel.ArchiveName)
	}
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, want) {
		return fmt.Errorf("%w: want=%s got=%s", ErrChecksumMismatch, want, got)
	}
	return nil
}

func extractBinary(archivePath, destDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	var found string
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
		base := filepath.Base(hdr.Name)
		if hdr.Typeflag != tar.TypeReg || base != BinaryName {
			continue
		}
		dst := filepath.Join(destDir, BinaryName+".new")
		out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return "", err
		}
		out.Close()
		found = dst
	}
	if found == "" {
		return "", fmt.Errorf("no %q entry found inside archive", BinaryName)
	}
	return found, nil
}

func replaceExe(currentPath, newPath string) error {
	if _, err := os.Stat(currentPath); err != nil {
		return fmt.Errorf("locate current executable %q: %w", currentPath, err)
	}

	oldPath := currentPath + ".old"
	if _, err := os.Stat(oldPath); err == nil {
		_ = os.Remove(oldPath)
	}
	if err := os.Rename(currentPath, oldPath); err != nil && !os.IsNotExist(err) {
		if !isPermission(err) {
			fmt.Fprintf(os.Stderr, "warning: could not move old binary aside: %v\n", err)
		}
	}

	if err := os.Rename(newPath, currentPath); err != nil {
		_ = os.Rename(oldPath, currentPath)
		return fmt.Errorf("replace executable %q: %w (hint: run install.sh to a writable prefix or use sudo)", currentPath, err)
	}
	return nil
}

func isPermission(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "permission denied") || strings.Contains(msg, "read-only file system")
}

func stripV(v string) string { return strings.TrimPrefix(v, "v") }

func CompareSemver(a, b string) int {
	if a == b {
		return 0
	}
	ap, apr := splitPre(a)
	bp, bpr := splitPre(b)
	if n := compareParts(ap, bp); n != 0 {
		return n
	}
	return comparePre(apr, bpr)
}

func splitPre(v string) ([]int, string) {
	idx := strings.IndexAny(v, "-+")
	if idx == -1 {
		return parseParts(v), ""
	}
	return parseParts(v[:idx]), v[idx+1:]
}

func parseParts(v string) []int {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			out = append(out, 0)
			continue
		}
		out = append(out, n)
	}
	return out
}

func compareParts(a, b []int) int {
	n := len(a)
	if len(b) > n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		var ai, bi int
		if i < len(a) {
			ai = a[i]
		}
		if i < len(b) {
			bi = b[i]
		}
		if ai != bi {
			if ai < bi {
				return -1
			}
			return 1
		}
	}
	return 0
}

func comparePre(a, b string) int {
	if a == b {
		return 0
	}
	if a == "" {
		return 1
	}
	if b == "" {
		return -1
	}
	if a < b {
		return -1
	}
	return 1
}
