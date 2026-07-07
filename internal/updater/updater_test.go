package updater

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOSArchMapping(t *testing.T) {
	if OS() == "" || Arch() == "" {
		t.Fatal("OS()/Arch() returned empty")
	}
	name, err := AssetName()
	if err != nil {
		t.Skipf("AssetName() not supported on this platform: %v", err)
	}
	if !strings.HasPrefix(name, BinaryName+"_") {
		t.Fatalf("unexpected asset name: %s", name)
	}
	if !strings.HasSuffix(name, ".tar.gz") {
		t.Fatalf("expected .tar.gz suffix, got %s", name)
	}
}

func TestCompareSemver(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.2.3", "1.2.3", 0},
		{"1.2.4", "1.2.3", 1},
		{"1.2.3", "1.2.4", -1},
		{"1.10.0", "1.9.9", 1},
		{"2.0.0", "1.99.99", 1},
		{"1.0.0", "1.0.0-rc1", 1},
		{"1.0.0-rc2", "1.0.0-rc1", 1},
		{"1.0.0-rc1", "1.0.0-rc2", -1},
	}
	for _, c := range cases {
		if got := CompareSemver(c.a, c.b); got != c.want {
			t.Errorf("CompareSemver(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		name    string
		latest  string
		current string
		want    bool
	}{
		{"dev short-circuits", "v9.9.9", "dev", false},
		{"empty short-circuits", "v9.9.9", "", false},
		{"equal", "v1.2.3", "v1.2.3", false},
		{"newer", "v1.2.4", "v1.2.3", true},
		{"older", "v1.2.2", "v1.2.3", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsNewer(tt.latest, tt.current)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("IsNewer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
			}
		})
	}
}

func TestParseChecksums(t *testing.T) {
	data := []byte(`abc123  acolyte_linux_amd64.tar.gz
DEF456 *acolyte_darwin_arm64.tar.gz

ignored line with no hash
`)
	got := parseChecksums(data)
	if got["acolyte_linux_amd64.tar.gz"] != "abc123" {
		t.Errorf("expected abc123, got %s", got["acolyte_linux_amd64.tar.gz"])
	}
	if !strings.EqualFold(got["acolyte_darwin_arm64.tar.gz"], "def456") {
		t.Errorf("expected DEF456 (case-insensitive), got %s", got["acolyte_darwin_arm64.tar.gz"])
	}
}

func TestExtractBinaryFromRealTarball(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	body := []byte("#!/bin/sh\necho hello\n")
	hdr := &tar.Header{Name: BinaryName, Mode: 0o755, Size: int64(len(body))}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}

	tmp := t.TempDir()
	src := filepath.Join(tmp, "fake.tar.gz")
	if err := os.WriteFile(src, buf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := extractBinary(src, tmp)
	if err != nil {
		t.Fatalf("extractBinary: %v", err)
	}
	data, err := os.ReadFile(got)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, body) {
		t.Fatalf("extracted bytes mismatch: %q vs %q", data, body)
	}
}

func TestLatestWithStubHTTP(t *testing.T) {
	asset, err := AssetName()
	if err != nil {
		t.Skipf("AssetName unsupported: %v", err)
	}
	body := buildGitHubReleaseJSON("v1.2.3", asset)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()

	rel, err := Latest(context.Background(), Options{APIBase: srv.URL})
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if rel.Tag != "v1.2.3" {
		t.Errorf("Tag = %q, want v1.2.3", rel.Tag)
	}
	want, _ := AssetName()
	if rel.ArchiveName != want {
		t.Errorf("ArchiveName = %q, want %q", rel.ArchiveName, want)
	}
	if rel.BrowserDownloadURL == "" {
		t.Errorf("BrowserDownloadURL empty")
	}
}

func TestLatestReportsAvailableAssetsWhenMissing(t *testing.T) {
	body := buildGitHubReleaseJSON("v1.2.3", "totally_unrelated_asset.zip")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()

	_, err := Latest(context.Background(), Options{APIBase: srv.URL})
	if err == nil {
		t.Fatal("expected ErrAssetNotFound, got nil")
	}
	if !errors.Is(err, ErrAssetNotFound) {
		t.Fatalf("expected ErrAssetNotFound in chain, got %v", err)
	}
	if !strings.Contains(err.Error(), "totally_unrelated_asset.zip") {
		t.Errorf("error should list available assets, got: %v", err)
	}
}

func TestIsOutdatedDevShortCircuit(t *testing.T) {
	orig := Version
	Version = "dev"
	defer func() { Version = orig }()
	_, outdated, err := IsOutdated(context.Background(), Options{})
	if !errors.Is(err, ErrDevBuild) {
		t.Fatalf("expected ErrDevBuild, got %v", err)
	}
	if outdated {
		t.Fatal("dev should never be reported as outdated")
	}
}

func TestIsOutdatedWithStubHTTP(t *testing.T) {
	orig := Version
	defer func() { Version = orig }()

	asset, err := AssetName()
	if err != nil {
		t.Skipf("AssetName unsupported: %v", err)
	}
	body := buildGitHubReleaseJSON("v9.9.9", asset)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()

	Version = "v9.9.8"
	_, outdated, err := IsOutdated(context.Background(), Options{APIBase: srv.URL})
	if err != nil {
		t.Fatalf("IsOutdated: %v", err)
	}
	if !outdated {
		t.Fatal("expected outdated=true")
	}

	Version = "v9.9.9"
	_, outdated, err = IsOutdated(context.Background(), Options{APIBase: srv.URL})
	if err != nil {
		t.Fatalf("IsOutdated: %v", err)
	}
	if outdated {
		t.Fatal("expected outdated=false when equal")
	}

	Version = "v10.0.0"
	_, outdated, err = IsOutdated(context.Background(), Options{APIBase: srv.URL})
	if err != nil {
		t.Fatalf("IsOutdated: %v", err)
	}
	if outdated {
		t.Fatal("expected outdated=false when current is newer")
	}
}

func TestApplyReplacesExecutableAndVerifiesChecksum(t *testing.T) {
	body := []byte("#!/bin/sh\necho NEW\n")
	var archiveBuf bytes.Buffer
	gz := gzip.NewWriter(&archiveBuf)
	tw := tar.NewWriter(gz)
	hdr := &tar.Header{Name: BinaryName, Mode: 0o755, Size: int64(len(body))}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatal(err)
	}
	_ = tw.Close()
	_ = gz.Close()

	archiveHash := sha256.Sum256(archiveBuf.Bytes())
	archiveHex := hex.EncodeToString(archiveHash[:])
	assetName, err := AssetName()
	if err != nil {
		t.Skipf("AssetName unsupported: %v", err)
	}

	checksums := fmt.Sprintf("%s  %s\n", archiveHex, assetName)

	exeDir := t.TempDir()
	exePath := filepath.Join(exeDir, BinaryName)
	if err := os.WriteFile(exePath, []byte("#!/bin/sh\necho OLD\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	releasePath := ""
	checksumsPath := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/archive":
			w.Header().Set("Content-Type", "application/octet-stream")
			_, _ = w.Write(archiveBuf.Bytes())
		case "/checksums.txt":
			w.Header().Set("Content-Type", "text/plain")
			_, _ = io.WriteString(w, checksums)
		case "/releases":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, buildGitHubReleaseJSON("v9.9.9", assetName))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	releasePath = srv.URL + "/archive"
	checksumsPath = srv.URL + "/checksums.txt"

	rel := &Release{
		Tag:                "v9.9.9",
		ArchiveName:        assetName,
		BrowserDownloadURL: releasePath,
		ChecksumsURL:       checksumsPath,
	}

	if err := Apply(context.Background(), rel, Options{
		ExePathOverride: exePath,
		Restart:         false,
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, body) {
		t.Fatalf("expected binary replacement, got %q", data)
	}
}

func TestApplyRejectsChecksumMismatch(t *testing.T) {
	assetName, err := AssetName()
	if err != nil {
		t.Skipf("AssetName unsupported: %v", err)
	}

	checksums := fmt.Sprintf("%s  %s\n", strings.Repeat("0", 64), assetName)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/archive":
			_, _ = w.Write([]byte("not a real tarball"))
		case "/checksums.txt":
			_, _ = io.WriteString(w, checksums)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	rel := &Release{
		Tag:                "v9.9.9",
		ArchiveName:        assetName,
		BrowserDownloadURL: srv.URL + "/archive",
		ChecksumsURL:       srv.URL + "/checksums.txt",
	}
	tmpExe := filepath.Join(t.TempDir(), BinaryName)
	_ = os.WriteFile(tmpExe, []byte("placeholder"), 0o755)

	applyErr := Apply(context.Background(), rel, Options{ExePathOverride: tmpExe, Restart: false})
	if applyErr == nil || !strings.Contains(applyErr.Error(), "checksum") {
		t.Fatalf("expected checksum mismatch error, got %v", applyErr)
	}
}

func buildGitHubReleaseJSON(tag, assetName string) string {
	type asset struct {
		Name               string `json:"name"`
		URL                string `json:"url"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Digest             string `json:"digest"`
	}
	type release struct {
		TagName string  `json:"tag_name"`
		Name    string  `json:"name"`
		Assets  []asset `json:"assets"`
	}
	r := release{
		TagName: tag,
		Name:    tag,
		Assets: []asset{
			{
				Name:               assetName,
				URL:                "https://api.github.com/test/asset",
				BrowserDownloadURL: "https://github.com/test/asset",
				Digest:             "sha256:0000000000000000000000000000000000000000000000000000000000000000",
			},
			{
				Name:               "checksums.txt",
				URL:                "https://api.github.com/test/checksums",
				BrowserDownloadURL: "https://github.com/test/checksums",
				Digest:             "",
			},
		},
	}
	b, _ := json.Marshal(r)
	return string(b)
}
