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
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/external-secrets/external-secrets-e2e/framework"
	frameworkv2 "github.com/external-secrets/external-secrets-e2e/framework/v2"
	esv1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1"
	esmeta "github.com/external-secrets/external-secrets/apis/meta/v1"
	gcpsmv2alpha1 "github.com/external-secrets/external-secrets/apis/provider/gcp/v2alpha1"
)

const defaultV2WaitTimeout = 60 * time.Second

type ProviderV2 struct {
	access    gcpAccessConfig
	backend   *GcpProvider
	framework *framework.Framework
}

func NewProviderV2(f *framework.Framework) *ProviderV2 {
	access := newGCPAccessConfigFromEnv()
	configureGCPRemoteRefKey(f)

	prov := &ProviderV2{
		access: access,
		backend: &GcpProvider{
			ServiceAccountName:      access.ServiceAccountName,
			ServiceAccountNamespace: "default",
			framework:               f,
			credentials:             access.Credentials,
			projectID:               access.ProjectID,
			clusterLocation:         access.ClusterLocation,
			clusterName:             access.ClusterName,
			access:                  access,
		},
		framework: f,
	}

	BeforeEach(func() {
		if !framework.IsV2ProviderMode() {
			return
		}
		skipIfGCPStaticEnvMissing(access)
	})

	return prov
}

func configureGCPRemoteRefKey(f *framework.Framework) {
	f.MakeRemoteRefKey = func(base string) string {
		if f.Namespace == nil {
			return base
		}
		suffix := f.Namespace.Name
		if len(suffix) > 8 {
			suffix = suffix[len(suffix)-8:]
		}
		if suffix == "" {
			return base
		}
		return fmt.Sprintf("%s-%s", base, suffix)
	}
}

func (p *ProviderV2) CreateSecret(key string, val framework.SecretEntry) {
	p.backend.CreateSecret(key, val)
}

func (p *ProviderV2) UpdateSecret(key string, val framework.SecretEntry) {
	p.backend.UpdateSecret(key, val)
}

func (p *ProviderV2) DeleteSecret(key string) {
	p.backend.DeleteSecret(key)
}

func useV2StaticAuth(prov *ProviderV2) func(*framework.TestCase) {
	return func(tc *framework.TestCase) {
		tc.Prepare = prov.prepareNamespacedProviderWithStaticAuthAtAddress(frameworkv2.ProviderAddress("gcp"))
	}
}

func (p *ProviderV2) prepareNamespacedProviderWithStaticAuthAtAddress(address string) func(*framework.TestCase, framework.SecretStoreProvider) {
	return func(_ *framework.TestCase, _ framework.SecretStoreProvider) {
		createSecretManagerV2StaticConfig(p.framework, p.framework.Namespace.Name, p.framework.Namespace.Name, p.access)
		frameworkv2.CreateProviderConnection(
			p.framework,
			p.framework.Namespace.Name,
			p.framework.Namespace.Name,
			address,
			gcpsmv2alpha1.GroupVersion.String(),
			gcpsmv2alpha1.SecretManagerKind,
			p.framework.Namespace.Name,
			p.framework.Namespace.Name,
		)
		frameworkv2.WaitForProviderConnectionReady(p.framework, p.framework.Namespace.Name, p.framework.Namespace.Name, defaultV2WaitTimeout)
	}
}

func newSecretManagerV2StaticConfig(namespace, name string, access gcpAccessConfig) *gcpsmv2alpha1.SecretManager {
	return &gcpsmv2alpha1.SecretManager{
		TypeMeta: metav1.TypeMeta{
			APIVersion: gcpsmv2alpha1.GroupVersion.String(),
			Kind:       gcpsmv2alpha1.SecretManagerKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: gcpsmv2alpha1.SecretManagerSpec{
			ProjectID: access.ProjectID,
			Auth: esv1.GCPSMAuth{
				SecretRef: &esv1.GCPSMAuthSecretRef{
					SecretAccessKey: esmeta.SecretKeySelector{
						Name: staticCredentialsSecretName,
						Key:  serviceAccountKey,
					},
				},
			},
		},
	}
}

func createSecretManagerV2StaticConfig(f *framework.Framework, namespace, name string, access gcpAccessConfig) *gcpsmv2alpha1.SecretManager {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      staticCredentialsSecretName,
			Namespace: namespace,
		},
		StringData: map[string]string{
			serviceAccountKey: access.Credentials,
		},
	}
	Expect(f.CreateObjectWithRetry(secret)).To(Succeed())

	cfg := newSecretManagerV2StaticConfig(namespace, name, access)
	Expect(f.CreateObjectWithRetry(cfg)).To(Succeed())
	return cfg
}
