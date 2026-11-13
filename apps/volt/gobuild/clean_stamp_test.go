package gobuild

import (
	"archive/tar"
	"compress/gzip"
	"debug/buildinfo"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunKeepsVCSStampClean pins the dirty-stamp fix: with dist/ inside an
// unignoring git checkout, an archive written there between platform builds
// used to make every later binary report vcs.modified=true. Each binary in a
// multi-platform build must stamp the commit as clean.
func TestRunKeepsVCSStampClean(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := scaffoldCLI(t)
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.test",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.test",
			"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	git("add", ".")
	git("commit", "-q", "-m", "init")

	cfg := hostCfg(dir)
	cfg.Platforms = []string{"linux/amd64", "linux/arm64", "darwin/arm64"}
	dist := filepath.Join(dir, "dist") // inside the repo, and not ignored
	res, err := Run(Options{Dir: dir, Version: "v0.1.0", DistDir: dist}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Assets) != 3 {
		t.Fatalf("assets: %v", res.Assets)
	}
	for _, asset := range res.Assets {
		bin := extractBinary(t, filepath.Join(dist, asset), cfg.Binary)
		info, err := buildinfo.ReadFile(bin)
		if err != nil {
			t.Fatalf("%s: %v", asset, err)
		}
		got := map[string]string{}
		for _, s := range info.Settings {
			got[s.Key] = s.Value
		}
		if got["vcs.revision"] == "" {
			t.Fatalf("%s: no vcs stamp at all: %v", asset, got)
		}
		if got["vcs.modified"] != "false" {
			t.Errorf("%s: vcs.modified=%q, want false (a build output dirtied the tree)", asset, got["vcs.modified"])
		}
	}
	if _, err := os.Stat(res.Checksums); err != nil || filepath.Dir(res.Checksums) != dist {
		t.Errorf("checksums %q not in dist: %v", res.Checksums, err)
	}
}

// extractBinary pulls the named binary out of a .tar.gz asset into a temp file.
func extractBinary(t *testing.T, archivePath, binary string) string {
	t.Helper()
	f, err := os.Open(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }() // read-only
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if filepath.Base(h.Name) != binary || strings.HasSuffix(h.Name, "/") {
			continue
		}
		out := filepath.Join(t.TempDir(), binary)
		w, err := os.Create(out)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.Copy(w, tr); err != nil {
			t.Fatal(err)
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		return out
	}
	t.Fatalf("%s: no %s inside", archivePath, binary)
	return ""
}
