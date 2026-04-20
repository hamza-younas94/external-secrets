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

package addon

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

func TestNeedsCRDPreinstallForV2ProvidersWhenHelmWouldCreateCRDs(t *testing.T) {
	t.Setenv("VERSION", "test-version")

	eso := NewESO(WithCRDs(), WithV2FakeProvider())
	if !needsCRDPreinstall(eso.HelmChart) {
		t.Fatal("expected v2 provider install with installCRDs=true to require CRD preinstall")
	}
}

func TestNeedsCRDPreinstallDisabledWhenHelmCRDsDisabled(t *testing.T) {
	t.Setenv("VERSION", "test-version")

	eso := NewESO(WithV2FakeProvider())
	if needsCRDPreinstall(eso.HelmChart) {
		t.Fatal("did not expect CRD preinstall when installCRDs=false")
	}
}

func TestNeedsCRDPreinstallForDefaultV2RuntimeCRDs(t *testing.T) {
	eso := NewESO(WithCRDs())
	if !needsCRDPreinstall(eso.HelmChart) {
		t.Fatal("expected default chart v2 runtime CRDs to require CRD preinstall")
	}
}

func TestCleanupProviderTLSSecretIfNeededDeletesStaleProviderSecret(t *testing.T) {
	t.Setenv("VERSION", "test-version")

	eso := NewESO(WithV2Namespace(), WithV2GCPProvider())
	eso.config = &Config{
		KubeClientSet: kubefake.NewSimpleClientset(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      providerTLSSecretName,
				Namespace: v2HelmNamespace,
			},
		}),
	}

	if err := eso.cleanupProviderTLSSecretIfNeeded(context.Background()); err != nil {
		t.Fatalf("cleanupProviderTLSSecretIfNeeded() error = %v", err)
	}

	_, err := eso.config.KubeClientSet.CoreV1().Secrets(v2HelmNamespace).Get(context.Background(), providerTLSSecretName, metav1.GetOptions{})
	if !apierrors.IsNotFound(err) {
		t.Fatalf("expected provider TLS secret to be deleted, got err = %v", err)
	}
}

func TestCleanupProviderTLSSecretIfNeededSkipsNonProviderInstall(t *testing.T) {
	eso := NewESO()
	eso.config = &Config{
		KubeClientSet: kubefake.NewSimpleClientset(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      providerTLSSecretName,
				Namespace: "default",
			},
		}),
	}

	if err := eso.cleanupProviderTLSSecretIfNeeded(context.Background()); err != nil {
		t.Fatalf("cleanupProviderTLSSecretIfNeeded() error = %v", err)
	}

	if _, err := eso.config.KubeClientSet.CoreV1().Secrets("default").Get(context.Background(), providerTLSSecretName, metav1.GetOptions{}); err != nil {
		t.Fatalf("expected non-provider install to leave secret in place, got err = %v", err)
	}
}
