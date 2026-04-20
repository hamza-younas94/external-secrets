/*
Copyright © The ESO Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gcp

import (
	"testing"

	frameworkv2 "github.com/external-secrets/external-secrets-e2e/framework/v2"
	gcpsmv2alpha1 "github.com/external-secrets/external-secrets/apis/provider/gcp/v2alpha1"
)

func TestNewSecretManagerV2StaticConfig(t *testing.T) {
	t.Parallel()

	cfg := newSecretManagerV2StaticConfig("workload-ns", "gcp-config", gcpAccessConfig{
		Credentials: "service-account-json",
		ProjectID:   "project-1",
	})

	if cfg.APIVersion != gcpsmv2alpha1.GroupVersion.String() {
		t.Fatalf("unexpected apiVersion: %q", cfg.APIVersion)
	}
	if cfg.Kind != gcpsmv2alpha1.SecretManagerKind {
		t.Fatalf("unexpected kind: %q", cfg.Kind)
	}
	if cfg.Namespace != "workload-ns" || cfg.Name != "gcp-config" {
		t.Fatalf("unexpected object metadata: %s/%s", cfg.Namespace, cfg.Name)
	}
	if cfg.Spec.ProjectID != "project-1" {
		t.Fatalf("unexpected project ID: %q", cfg.Spec.ProjectID)
	}
	if cfg.Spec.Auth.SecretRef == nil {
		t.Fatal("expected static auth secretRef")
	}
	if got := cfg.Spec.Auth.SecretRef.SecretAccessKey.Name; got != staticCredentialsSecretName {
		t.Fatalf("unexpected secret ref name: %q", got)
	}
	if got := cfg.Spec.Auth.SecretRef.SecretAccessKey.Key; got != serviceAccountKey {
		t.Fatalf("unexpected secret ref key: %q", got)
	}
}

func TestProviderAddressInNamespace(t *testing.T) {
	t.Parallel()

	got := frameworkv2.ProviderAddressInNamespace("gcp", "gcp-provider-system")
	if got != "provider-gcp.gcp-provider-system.svc:8080" {
		t.Fatalf("unexpected address: %s", got)
	}
}
