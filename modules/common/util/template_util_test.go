package util

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	. "github.com/onsi/gomega"
)

var (
	// path with templates files for testing relative to the package dir
	templatePath = "testdata/templates"
)

func TestAdd(t *testing.T) {

	t.Run("Sum of two int", func(t *testing.T) {
		g := NewWithT(t)

		s := add(1, 1)

		g.Expect(s).To(BeIdenticalTo(2))
	})
}

func TestLower(t *testing.T) {

	t.Run("Lower string", func(t *testing.T) {
		g := NewWithT(t)

		s := lower("FOObaR")

		g.Expect(s).To(BeIdenticalTo("foobar"))
	})
}

func TestGetTemplatesPath(t *testing.T) {
	// set the env var used to specify the template path in the container case
	os.Setenv("OPERATOR_TEMPLATES", templatePath)

	t.Run("Lower string", func(t *testing.T) {
		g := NewWithT(t)

		p, _ := GetTemplatesPath()

		g.Expect(p).To(BeIdenticalTo(templatePath))
	})
}

func TestGetAllTemplates(t *testing.T) {

	// get the package directory
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}

	// set the env var used to specify the template path in the container case
	os.Setenv("OPERATOR_TEMPLATES", filepath.Join(path.Dir(filename), templatePath))

	tests := []struct {
		name     string
		kind     string
		tmplType TType
		version  string
		want     []string
	}{
		{
			name:     "Get TemplateTypeConfig templates with no version",
			kind:     "testservice",
			tmplType: TemplateTypeConfig,
			version:  "",
			want: []string{
				filepath.Join(path.Dir(filename), templatePath, "testservice", "config", "config.json"),
				filepath.Join(path.Dir(filename), templatePath, "testservice", "config", "foo.conf"),
			},
		},
		{
			name:     "Get TemplateTypeScripts templates with version",
			kind:     "testservice",
			tmplType: TemplateTypeScripts,
			version:  "1.0",
			want: []string{
				filepath.Join(path.Dir(filename), templatePath, "testservice", "bin", "1.0", "init.sh"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			p, _ := GetTemplatesPath()
			g.Expect(p).To(BeADirectory())

			templatesFiles := GetAllTemplates(p, tt.kind, string(tt.tmplType), tt.version)

			g.Expect(templatesFiles).To(HaveLen(len(tt.want)))
			g.Expect(templatesFiles).Should(HaveEach(BeARegularFile()))
			g.Expect(templatesFiles).Should(ConsistOf(tt.want))
		})
	}
}

func TestGetTemplateData(t *testing.T) {

	// get the package directory
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}

	// set the env var used to specify the template path in the container case
	os.Setenv("OPERATOR_TEMPLATES", filepath.Join(path.Dir(filename), templatePath))

	tests := []struct {
		name  string
		tmpl  Template
		want  map[string]string
		error bool
	}{
		{
			name: "Render TemplateTypeConfig templates with no version",
			tmpl: Template{
				Name:         "testservice",
				Namespace:    "somenamespace",
				Type:         TemplateTypeConfig,
				InstanceType: "testservice",
				Version:      "",
				ConfigOptions: map[string]interface{}{
					"ServiceUser": "foo",
					"Count":       1,
					"Upper":       "BAR",
				},
				AdditionalTemplate: map[string]string{},
			},
			want: map[string]string{
				"config.json": "{\n    \"command\": \"/usr/sbin/httpd -DFOREGROUND\",\n}\n",
				"foo.conf":    "username = foo\ncount = 1\nadd = 3\nlower = bar\n",
			},
			error: false,
		},
		{
			name: "Render TemplateTypeScripts templates with version",
			tmpl: Template{
				Name:               "testservice",
				Namespace:          "somenamespace",
				Type:               TemplateTypeScripts,
				InstanceType:       "testservice",
				Version:            "1.0",
				AdditionalTemplate: map[string]string{},
			},
			want: map[string]string{
				"init.sh": "#!/bin//bash\nset -ex\n\necho foo\nexit 0\n",
			},
			error: false,
		},
		{
			name: "Render TemplateTypeConfig templates with AdditionalTemplate",
			tmpl: Template{
				Name:         "testservice",
				Namespace:    "somenamespace",
				Type:         TemplateTypeConfig,
				InstanceType: "testservice",
				Version:      "",
				ConfigOptions: map[string]interface{}{
					"ServiceUser": "foo",
					"Count":       1,
					"Upper":       "BAR",
					"Message":     "some common func",
				},
				AdditionalTemplate: map[string]string{"common.sh": "/common/common.sh"},
			},
			want: map[string]string{
				"config.json": "{\n    \"command\": \"/usr/sbin/httpd -DFOREGROUND\",\n}\n",
				"foo.conf":    "username = foo\ncount = 1\nadd = 3\nlower = bar\n",
				"common.sh":   "#!/bin/bash\nset -e\n\nfunction common_func {\n  echo some common func\n}\n",
			},
			error: false,
		},
		{
			name: "Render TemplateTypeNone templates with AdditionalTemplate",
			tmpl: Template{
				Name:         "testservice",
				Namespace:    "somenamespace",
				Type:         TemplateTypeNone,
				InstanceType: "testservice",
				Version:      "",
				ConfigOptions: map[string]interface{}{
					"ServiceUser": "foo",
					"Count":       1,
					"Upper":       "BAR",
					"Message":     "some common func",
				},
				AdditionalTemplate: map[string]string{"common.sh": "/common/common.sh"},
			},
			want: map[string]string{
				"common.sh": "#!/bin/bash\nset -e\n\nfunction common_func {\n  echo some common func\n}\n",
			},
			error: false,
		},
		{
			name: "Render TemplateTypeConfig templates with incomplete ConfigOptions",
			tmpl: Template{
				Name:         "testservice",
				Namespace:    "somenamespace",
				Type:         TemplateTypeConfig,
				InstanceType: "testservice",
				Version:      "",
				ConfigOptions: map[string]interface{}{
					"Count": 1,
					"Upper": "BAR",
				},
				AdditionalTemplate: map[string]string{},
			},
			want:  map[string]string{},
			error: true,
		},
		{
			name: "Render TemplateTypeConfig templates with AdditionamTemplate and incomplete ConfigOptions",
			tmpl: Template{
				Name:         "testservice",
				Namespace:    "somenamespace",
				Type:         TemplateTypeConfig,
				InstanceType: "testservice",
				Version:      "",
				ConfigOptions: map[string]interface{}{
					"ServiceUser": "foo",
					"Count":       1,
					"Upper":       "BAR",
				},
				AdditionalTemplate: map[string]string{"common.sh": "/common/common.sh"},
			},
			want:  map[string]string{},
			error: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			p, _ := GetTemplatesPath()
			g.Expect(p).To(BeADirectory())

			templatesFiles, err := GetTemplateData(tt.tmpl)

			if tt.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())

				g.Expect(templatesFiles).To(HaveLen(len(tt.want)))
				for k, v := range tt.want {
					g.Expect(templatesFiles).To(HaveKeyWithValue(k, v))
				}
			}
		})
	}
}
