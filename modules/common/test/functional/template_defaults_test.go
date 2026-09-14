/*
Copyright 2026 Red Hat

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
package functional

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openstack-k8s-operators/lib-common/modules/common/secret"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Template defaults", func() {
	var namespace string

	// createDefaultsConfigMap - create a ConfigMap holding template default values
	createDefaultsConfigMap := func(name string, data map[string]string) {
		Expect(cClient.Create(ctx, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Data: data,
		})).To(Succeed())
	}

	// sslTemplate - a Template that renders nothing but the common ssl.conf.
	// TemplateTypeNone skips the on-disk operator templates, which the test
	// suite does not ship.
	sslTemplate := func(name string, configOptions map[string]any) util.Template {
		return util.Template{
			Name:            name,
			Namespace:       namespace,
			Type:            util.TemplateTypeNone,
			CommonTemplates: []string{"ssl.conf"},
			ConfigOptions:   configOptions,
		}
	}

	BeforeEach(func() {
		namespace = uuid.New().String()
		th.CreateNamespace(namespace)
	})

	DescribeTable("ApplyTemplateDefaults",
		func(configMaps map[string]map[string]string, configMapNames []string, configOptions, expected map[string]any) {
			for name, data := range configMaps {
				createDefaultsConfigMap(name, data)
			}

			got, err := util.ApplyTemplateDefaults(
				ctx, h, []util.Template{sslTemplate("t", configOptions)}, configMapNames)

			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].ConfigOptions).To(Equal(expected))
		},
		Entry("leaves ConfigOptions alone when the ConfigMap does not exist",
			nil,
			[]string{util.TLSProfileConfigMap},
			map[string]any{"SSLProtocol": "from-caller"},
			map[string]any{"SSLProtocol": "from-caller"}),
		Entry("merges the ConfigMap data in when the caller passed no ConfigOptions",
			map[string]map[string]string{
				util.TLSProfileConfigMap: {"SSLCipherSuite": "ECDHE-RSA-AES256-GCM-SHA384"},
			},
			[]string{util.TLSProfileConfigMap},
			nil,
			map[string]any{"SSLCipherSuite": "ECDHE-RSA-AES256-GCM-SHA384"}),
		Entry("keeps the value set by the caller and adds the rest",
			map[string]map[string]string{
				util.TLSProfileConfigMap: {
					"SSLProtocol":    "from-defaults",
					"SSLCipherSuite": "from-defaults",
				},
			},
			[]string{util.TLSProfileConfigMap},
			map[string]any{"SSLProtocol": "from-caller"},
			map[string]any{"SSLProtocol": "from-caller", "SSLCipherSuite": "from-defaults"}),
		Entry("lets earlier ConfigMaps win and skips the ones that are missing",
			map[string]map[string]string{
				"first":  {"SSLProtocol": "from-first"},
				"second": {"SSLProtocol": "from-second", "SSLCipherSuite": "only-in-second"},
			},
			[]string{"absent", "first", "second"},
			map[string]any{},
			map[string]any{"SSLProtocol": "from-first", "SSLCipherSuite": "only-in-second"}),
	)

	It("does not modify the Templates or ConfigOptions owned by the caller", func() {
		createDefaultsConfigMap(util.TLSProfileConfigMap, map[string]string{
			"SSLCipherSuite": "from-defaults",
		})
		configOptions := map[string]any{"SSLProtocol": "from-caller"}
		tmpls := []util.Template{sslTemplate("t", configOptions)}

		_, err := util.ApplyTemplateDefaults(ctx, h, tmpls, []string{util.TLSProfileConfigMap})

		Expect(err).NotTo(HaveOccurred())
		Expect(configOptions).To(HaveLen(1))
		Expect(tmpls[0].ConfigOptions).NotTo(HaveKey("SSLCipherSuite"))
	})

	It("leaves custom templates untouched", func() {
		createDefaultsConfigMap(util.TLSProfileConfigMap, map[string]string{
			"SSLCipherSuite": "from-defaults",
		})
		custom := sslTemplate("t", map[string]any{})
		custom.Type = util.TemplateTypeCustom

		got, err := util.ApplyTemplateDefaults(ctx, h, []util.Template{custom}, []string{util.TLSProfileConfigMap})

		Expect(err).NotTo(HaveOccurred())
		Expect(got[0].ConfigOptions).To(BeEmpty())
	})

	When("EnsureSecrets renders a template", func() {
		// createOwner - an object in the same namespace, so the rendered Secret
		// gets a controller reference rather than cross-namespace owner labels
		createOwner := func() *corev1.ConfigMap {
			owner := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "owner",
					Namespace: namespace,
				},
			}
			Expect(cClient.Create(ctx, owner)).To(Succeed())
			return owner
		}

		// renderSSLConf - run EnsureSecrets over a single ssl.conf template and
		// read the rendered file back out of the Secret it created
		renderSSLConf := func(name string, configOptions map[string]any) string {
			Expect(secret.EnsureSecrets(
				ctx, h, createOwner(),
				[]util.Template{sslTemplate(name, configOptions)},
				nil)).To(Succeed())

			got := &corev1.Secret{}
			Expect(cClient.Get(ctx, types.NamespacedName{
				Name:      name,
				Namespace: namespace,
			}, got)).To(Succeed())
			return string(got.Data["ssl.conf"])
		}

		It("falls back to the template defaults when no profile ConfigMap exists", func() {
			Expect(renderSSLConf("no-profile", map[string]any{})).To(ContainSubstring(
				"SSLCipherSuite ECDHE+AESGCM:DHE+AESGCM:!aNULL:!MD5:!RC4:!3DES\n"))
		})

		It("applies the profile ConfigMap without the template asking for it", func() {
			createDefaultsConfigMap(util.TLSProfileConfigMap, map[string]string{
				"SSLCipherSuite": "ECDHE-RSA-AES256-GCM-SHA384",
				"SSLProtocol":    "-all +TLSv1.3",
			})

			rendered := renderSSLConf("from-profile", map[string]any{})

			Expect(rendered).To(ContainSubstring("SSLCipherSuite ECDHE-RSA-AES256-GCM-SHA384\n"))
			Expect(rendered).To(ContainSubstring("SSLProtocol -all +TLSv1.3\n"))
		})

		It("lets an explicit ConfigOptions value win over the profile ConfigMap", func() {
			createDefaultsConfigMap(util.TLSProfileConfigMap, map[string]string{
				"SSLCipherSuite": "from-profile",
				"SSLProtocol":    "-all +TLSv1.3",
			})

			rendered := renderSSLConf("operator-override", map[string]any{
				"SSLCipherSuite": "set-by-operator",
			})

			Expect(rendered).To(ContainSubstring("SSLCipherSuite set-by-operator\n"))
			Expect(rendered).To(ContainSubstring("SSLProtocol -all +TLSv1.3\n"))
		})
	})
})
