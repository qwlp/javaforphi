package checker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/javaforphi/javaforphi/internal/catalog"
)

func resolveDependencies(ctx context.Context, dependencies []catalog.Dependency) ([]string, error) {
	if len(dependencies) == 0 {
		return nil, nil
	}
	cacheRoot := os.Getenv("PHI_CACHE")
	if cacheRoot == "" {
		cacheRoot = os.Getenv("JAVAFORPHI_CACHE") // Legacy name used by v0.1.0.
	}
	if cacheRoot == "" {
		userCache, err := os.UserCacheDir()
		if err != nil {
			return nil, fmt.Errorf("find user cache: %w", err)
		}
		cacheRoot = filepath.Join(userCache, "phi", "dependencies")
	}
	if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
		return nil, fmt.Errorf("create dependency cache: %w", err)
	}

	paths := make([]string, 0, len(dependencies))
	for _, dependency := range dependencies {
		if dependency.Name == "" || strings.ContainsAny(dependency.Name, `/\\`) {
			return nil, fmt.Errorf("invalid dependency filename %q", dependency.Name)
		}
		target := filepath.Join(cacheRoot, dependency.Name)
		if matchesSHA256(target, dependency.SHA256) {
			paths = append(paths, target)
			continue
		}
		if err := downloadVerified(ctx, dependency, target); err != nil {
			return nil, err
		}
		paths = append(paths, target)
	}
	return paths, nil
}

func downloadVerified(ctx context.Context, dependency catalog.Dependency, target string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, dependency.URL, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 60 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("download %s: %w", dependency.Name, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: server returned %s", dependency.Name, response.Status)
	}

	temporary, err := os.CreateTemp(filepath.Dir(target), ".dependency-*")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(temporary, hash), io.LimitReader(response.Body, 32<<20)); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(actual, dependency.SHA256) {
		return fmt.Errorf("download %s: SHA-256 mismatch (got %s)", dependency.Name, actual)
	}
	if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(temporaryName, target); err != nil {
		return err
	}
	return nil
}

func matchesSHA256(filename, expected string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false
	}
	return strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), expected)
}
