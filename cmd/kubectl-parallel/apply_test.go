package main

import (
	"os"
	"testing"
)

func TestCollectManifestsGroupsResourcesByLabel(t *testing.T) {
	dir := t.TempDir()

	f1, err := os.CreateTemp(dir, "m1-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer f1.Close()
	if _, err := f1.WriteString(
		"apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: g1a\n  labels:\n    parallel/group: g1\n---\n" +
			"apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: d1\n"); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}

	f2, err := os.CreateTemp(dir, "m2-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer f2.Close()
	if _, err := f2.WriteString(
		"apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: g1b\n  labels:\n    parallel/group: g1\n---\n" +
			"apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: d2\n---\n" +
			"apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: g2a\n  labels:\n    parallel/group: g2\n"); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}

	manifests, err := collectManifests([]string{f1.Name(), f2.Name()}, defaultLabel)
	if err != nil {
		t.Fatalf("collectManifests returned error: %v", err)
	}

	if got := len(manifests["g1"]); got != 2 {
		t.Fatalf("expected 2 g1 resources, got %d", got)
	}
	if got := len(manifests["g2"]); got != 1 {
		t.Fatalf("expected 1 g2 resource, got %d", got)
	}
	if got := len(manifests["default"]); got != 2 {
		t.Fatalf("expected 2 default resources, got %d", got)
	}
}
