package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Platform string

const (
	Android     Platform = "android"
	IOS         Platform = "ios"
	ReactNative Platform = "react-native"
	Flutter     Platform = "flutter"
	Unknown     Platform = "unknown"
)

type Report struct {
	Root         string     `json:"root"`
	Platforms    []Platform `json:"platforms"`
	ApplicationID string    `json:"applicationId,omitempty"`
	BundleID     string     `json:"bundleId,omitempty"`
	PackageName  string     `json:"packageName,omitempty"`
	HasExpo      bool       `json:"hasExpo,omitempty"`
	Hints        []string   `json:"hints"`
}

func Inspect(root string) (Report, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return Report{}, err
	}
	r := Report{Root: abs, Platforms: []Platform{}}
	if exists(abs, "pubspec.yaml") {
		r.Platforms = append(r.Platforms, Flutter)
	}
	if pkgJSON(abs) {
		r.Platforms = append(r.Platforms, ReactNative)
		r.HasExpo = exists(abs, "app.json") || exists(abs, "app.config.ts") || exists(abs, "app.config.js")
	}
	if hasAndroid(abs) {
		r.Platforms = append(r.Platforms, Android)
		r.ApplicationID = findApplicationID(abs)
		r.PackageName = r.ApplicationID
	}
	if hasIOS(abs) {
		r.Platforms = append(r.Platforms, IOS)
		r.BundleID = findBundleID(abs)
	}
	if len(r.Platforms) == 0 {
		r.Platforms = []Platform{Unknown}
		r.Hints = append(r.Hints, "No Android, iOS, React Native, or Flutter project markers found.")
	}
	if r.ApplicationID == "" && contains(r.Platforms, Android) {
		r.Hints = append(r.Hints, "Could not read applicationId. Pass packageName when binding Android.")
	}
	if r.BundleID == "" && contains(r.Platforms, IOS) {
		r.Hints = append(r.Hints, "Could not read CFBundleIdentifier. Pass bundleId, teamId, and appleId when binding iOS.")
	}
	return r, nil
}

func exists(root, name string) bool {
	_, err := os.Stat(filepath.Join(root, name))
	return err == nil
}

func pkgJSON(root string) bool {
	b, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return false
	}
	var m map[string]any
	if json.Unmarshal(b, &m) != nil {
		return false
	}
	deps := map[string]any{}
	if d, ok := m["dependencies"].(map[string]any); ok {
		deps = d
	}
	_, rn := deps["react-native"]
	_, expo := deps["expo"]
	return rn || expo
}

func hasAndroid(root string) bool {
	return exists(root, "settings.gradle") ||
		exists(root, "android") ||
		exists(root, "app/build.gradle") ||
		exists(root, "app/build.gradle.kts")
}

func hasIOS(root string) bool {
	if exists(root, "ios") {
		return true
	}
	matches, _ := filepath.Glob(filepath.Join(root, "*.xcodeproj"))
	return len(matches) > 0
}

func findApplicationID(root string) string {
	candidates := []string{
		filepath.Join(root, "app", "build.gradle"),
		filepath.Join(root, "app", "build.gradle.kts"),
		filepath.Join(root, "android", "app", "build.gradle"),
		filepath.Join(root, "android", "app", "build.gradle.kts"),
	}
	for _, p := range candidates {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "applicationId") {
				line = strings.TrimPrefix(line, "applicationId")
				line = strings.Trim(line, " =\"'")
				if line != "" {
					return line
				}
			}
		}
	}
	return ""
}

func findBundleID(root string) string {
	plist := filepath.Join(root, "ios")
	var found string
	_ = filepath.WalkDir(plist, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.EqualFold(d.Name(), "Info.plist") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if id := plistString(string(b), "CFBundleIdentifier"); id != "" && !strings.Contains(id, "$(") {
			found = id
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func plistString(xml, key string) string {
	needle := "<key>" + key + "</key>"
	i := strings.Index(xml, needle)
	if i < 0 {
		return ""
	}
	rest := xml[i+len(needle):]
	open := strings.Index(rest, "<string>")
	close := strings.Index(rest, "</string>")
	if open < 0 || close < 0 || close <= open {
		return ""
	}
	return strings.TrimSpace(rest[open+len("<string>") : close])
}

func contains(ps []Platform, p Platform) bool {
	for _, x := range ps {
		if x == p {
			return true
		}
	}
	return false
}
