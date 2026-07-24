/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package serviceaccount

import (
	"testing"

	. "github.com/onsi/gomega" //revive:disable:dot-imports
)

func TestKubeAPIAccessVolume(t *testing.T) {
	g := NewWithT(t)

	vol := KubeAPIAccessVolume()

	g.Expect(vol.Name).To(Equal(KubeAPIAccessVolumeName))
	g.Expect(vol.Projected).NotTo(BeNil())
	g.Expect(*vol.Projected.DefaultMode).To(Equal(int32(0444)))

	sources := vol.Projected.Sources
	g.Expect(sources).To(HaveLen(3))

	g.Expect(sources[0].ServiceAccountToken).NotTo(BeNil())
	g.Expect(sources[0].ServiceAccountToken.Path).To(Equal("token"))
	g.Expect(*sources[0].ServiceAccountToken.ExpirationSeconds).To(Equal(DefaultTokenExpirationSeconds))

	g.Expect(sources[1].ConfigMap).NotTo(BeNil())
	g.Expect(sources[1].ConfigMap.Name).To(Equal("kube-root-ca.crt"))
	g.Expect(sources[1].ConfigMap.Items).To(HaveLen(1))
	g.Expect(sources[1].ConfigMap.Items[0].Key).To(Equal("ca.crt"))

	g.Expect(sources[2].DownwardAPI).NotTo(BeNil())
	g.Expect(sources[2].DownwardAPI.Items).To(HaveLen(1))
	g.Expect(sources[2].DownwardAPI.Items[0].Path).To(Equal("namespace"))
	g.Expect(sources[2].DownwardAPI.Items[0].FieldRef.FieldPath).To(Equal("metadata.namespace"))
}

func TestKubeAPIAccessVolumeCustomExpiration(t *testing.T) {
	g := NewWithT(t)

	var customExpiration int64 = 7200
	vol := KubeAPIAccessVolume(customExpiration)

	g.Expect(*vol.Projected.Sources[0].ServiceAccountToken.ExpirationSeconds).To(Equal(customExpiration))
}

func TestKubeAPIAccessVolumeMount(t *testing.T) {
	g := NewWithT(t)

	mount := KubeAPIAccessVolumeMount()

	g.Expect(mount.Name).To(Equal(KubeAPIAccessVolumeName))
	g.Expect(mount.MountPath).To(Equal(KubeAPIAccessMountPath))
	g.Expect(mount.ReadOnly).To(BeTrue())
}
