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

func TestIndent(t *testing.T) {

	t.Run("Indent string", func(t *testing.T) {
		g := NewWithT(t)
		const in = `foo
bar`
		// 5 tabs and line break
		const expct = `					foo
					bar
`

		s := indent(5, in)

		g.Expect(s).To(BeIdenticalTo(expct))
	})
}

func TestRemoveNewLines(t *testing.T) {

	t.Run("Remove duplicate new lines", func(t *testing.T) {
		g := NewWithT(t)
		const in = `	foo

  bar


foo




bar`

		const expct = `	foo

  bar

foo

bar
`

		s := removeNewLines(1, in)

		g.Expect(s).To(BeIdenticalTo(expct))
	})
}

func TestExecTempl(t *testing.T) {

	t.Run("ExecTempl", func(t *testing.T) {
		g := NewWithT(t)
		const myTmpl = `{{define "my-template"}}my-template


content



with empty lines



to
remove
{{end}}
See result:
{{$var := execTempl "my-template" . | removeNewLines 1}}
{{$var}}`

		// render template using execTempl and remove more then 1 continuous empty lines
		const expct = `
See result:

my-template

content

with empty lines

to
remove
`
		renderedTemplate, err := ExecuteTemplateData(myTmpl, "")
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(renderedTemplate).To(BeIdenticalTo(expct))
	})
}

func TestRemoveNewLinesInSections(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		cleaned string
	}{
		{
			name:    "Empty input",
			raw:     "",
			cleaned: "",
		},
		{
			name:    "Single empty line",
			raw:     "\n",
			cleaned: "",
		},
		{
			name:    "Two empty lines",
			raw:     "\n\n",
			cleaned: "",
		},
		{
			name:    "Insert newline at end of file",
			raw:     "foo",
			cleaned: "foo\n",
		},
		{
			name:    "Remove starting empty line",
			raw:     "\nfoo",
			cleaned: "foo\n",
		},
		{
			name:    "Remove starting empty lines",
			raw:     "\n\nfoo",
			cleaned: "foo\n",
		},
		{
			name:    "Remove extra empty line at the end",
			raw:     "foo\n\n",
			cleaned: "foo\n",
		},
		{
			name:    "Remove extra empty lines at the end",
			raw:     "foo\n\n\n",
			cleaned: "foo\n",
		},
		{
			name:    "Keep subsequent data lines",
			raw:     "foo\nbar",
			cleaned: "foo\nbar\n",
		},
		{
			name:    "Remove empty line between subsequent data",
			raw:     "foo\n\nbar",
			cleaned: "foo\nbar\n",
		},
		{
			name:    "Extra spaces around data lines are not kept",
			raw:     "\n\n  foo  \n\n  bar  ",
			cleaned: "foo\nbar\n",
		},
		{
			name:    "Extra spaces around section lines are not kept",
			raw:     "\n\n  [foo]  \n\n  [bar]  ",
			cleaned: "[foo]\n\n[bar]\n",
		},
		{
			name:    "Remove extra lines with spaces only",
			raw:     " \n  \nfoo\n \nbar\n  \n  ",
			cleaned: "foo\nbar\n",
		},
		{
			name:    "Remove starting empty line from section header",
			raw:     "\n[foo]",
			cleaned: "[foo]\n",
		},
		{
			name:    "Remove starting empty lines from section header",
			raw:     "\n\n[foo]",
			cleaned: "[foo]\n",
		},
		{
			name:    "Remove extra empty line after section header",
			raw:     "[foo]\n\n",
			cleaned: "[foo]\n",
		},
		{
			name:    "Remove extra empty lines after section header",
			raw:     "[foo]\n\n\n",
			cleaned: "[foo]\n",
		},
		{
			name:    "Insert empty line after section header at the end",
			raw:     "[foo]",
			cleaned: "[foo]\n",
		},
		{
			name:    "Keep one empty line between section headers",
			raw:     "[foo]\n\n[bar]",
			cleaned: "[foo]\n\n[bar]\n",
		},
		{
			name:    "Insert one empty line between section headers",
			raw:     "[foo]\n[bar]",
			cleaned: "[foo]\n\n[bar]\n",
		},
		{
			name:    "Remove more empty lines between section headers",
			raw:     "[foo]\n\n\n[bar]",
			cleaned: "[foo]\n\n[bar]\n",
		},
		{
			name:    "Remove extra empty line between section header and data",
			raw:     "[foo]\n\nbar",
			cleaned: "[foo]\nbar\n",
		},
		{
			name:    "Remove extra empty lines between section header and data",
			raw:     "[foo]\n\n\nbar",
			cleaned: "[foo]\nbar\n",
		},
		{
			name:    "Insert extra line between sections",
			raw:     "[foo]\nbar\n[goo]\nbaz",
			cleaned: "[foo]\nbar\n\n[goo]\nbaz\n",
		},
		{
			name:    "Remove extra lines between sections",
			raw:     "[foo]\nbar\n\n\n[goo]\nbaz",
			cleaned: "[foo]\nbar\n\n[goo]\nbaz\n",
		},
		{
			name:    "Insert no new line when there is a parameter value which brackets",
			raw:     "[foo]\nkey=[value]\n[bar]",
			cleaned: "[foo]\nkey=[value]\n\n[bar]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			cleaned := removeNewLinesInSections(tt.raw)
			g.Expect(cleaned).To(Equal(tt.cleaned))
		})
	}
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
				filepath.Join(path.Dir(filename), templatePath, "testservice", "config", "bar.conf"),
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
				"bar.conf":    "[DEFAULT]\nstate_path = /var/lib/nova\ndebug=true\nsome_parameter_with_brackets=[test]\ncompute_driver = libvirt.LibvirtDriver\n\n[oslo_concurrency]\nlock_path = /var/lib/nova/tmp\n",
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
				"bar.conf":    "[DEFAULT]\nstate_path = /var/lib/nova\ndebug=true\nsome_parameter_with_brackets=[test]\ncompute_driver = libvirt.LibvirtDriver\n\n[oslo_concurrency]\nlock_path = /var/lib/nova/tmp\n",
				"config.json": "{\n    \"command\": \"/usr/sbin/httpd -DFOREGROUND\",\n}\n",
				"foo.conf":    "username = foo\ncount = 1\nadd = 3\nlower = bar\n",
				"common.sh":   "#!/bin/bash\nset -e\n\nfunction common_func {\n  echo some common func\n}\n",
			},
			error: false,
		},
		{
			name: "Render TemplateTypeConfig templates with StringTemplate",
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
				StringTemplate: map[string]string{"common.sh": `#!/bin/bash
set -e

function common_func {
  echo {{ .Message }}
}`},
			},
			want: map[string]string{
				"bar.conf":    "[DEFAULT]\nstate_path = /var/lib/nova\ndebug=true\nsome_parameter_with_brackets=[test]\ncompute_driver = libvirt.LibvirtDriver\n\n[oslo_concurrency]\nlock_path = /var/lib/nova/tmp\n",
				"config.json": "{\n    \"command\": \"/usr/sbin/httpd -DFOREGROUND\",\n}\n",
				"foo.conf":    "username = foo\ncount = 1\nadd = 3\nlower = bar\n",
				"common.sh":   "#!/bin/bash\nset -e\n\nfunction common_func {\n  echo some common func\n}",
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
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(templatesFiles).To(HaveLen(len(tt.want)))
				for k, v := range tt.want {
					g.Expect(templatesFiles).To(HaveKeyWithValue(k, v))
				}
			}
		})
	}
}

// Run the new line section cleaning twice on an input and ensure that the second cleaning
// does nothing as the first run cleaned everything
// This was failing due to empty line handling between sections is unstable.
func TestRemoveNewLinesInSectionsIsStable(t *testing.T) {
	g := NewWithT(t)

	input := `
[foo]
boo=1
bar=1
[goo]
baz=1
`
	cleaned := removeNewLinesInSections(input)
	cleaned2 := removeNewLinesInSections(cleaned)

	g.Expect(cleaned2).To(Equal(cleaned))
}
