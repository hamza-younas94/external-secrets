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

package e2e

import (
	"strings"
	"testing"
)

func TestNewProviderV2ESOIncludesGCPProvider(t *testing.T) {
	t.Parallel()

	eso := newProviderV2ESO()

	for _, variable := range eso.HelmChart.Vars {
		if !strings.HasSuffix(variable.Key, ".name") {
			continue
		}
		if variable.Value == "gcp" {
			return
		}
	}

	t.Fatal("expected v2 provider suite ESO installation to include the gcp provider")
}
