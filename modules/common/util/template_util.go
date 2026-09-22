/*
Copyright 2022 Red Hat

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

package util

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//go:embed templates/common/config/*
var commonTemplates embed.FS

// TType - TemplateType
type TType string

const (
	// TemplateTypeScripts - scripts type
	TemplateTypeScripts TType = "bin"
	// TemplateTypeConfig - config type
	TemplateTypeConfig TType = "config"
	// TemplateTypeCustom - custom config type, the secret/cm will not get upated as it is exected that the content is owned by a user
	// if the configmap/secret does not exist on first check, it gets created
	TemplateTypeCustom TType = "custom"
	// TemplateTypeNone - none type, don't add configs from a directory, only files from AdditionalData
	TemplateTypeNone TType = "none"
)

// Template - config map and secret details
type Template struct {
	Name               string                 // name of the cm/secret to create based of the Template. Check secret/configmap pkg on details how it is used.
	Namespace          string                 // name of the nanmespace to create the cm/secret. Check secret/configmap pkg on details how it is used.
	Type               TType                  // type of the templates, see TTtypes
	InstanceType       string                 // the CRD name in lower case, to separate the templates for each CRD in /templates
	SecretType         corev1.SecretType      // Secrets only, defaults to "Opaque"
	AdditionalTemplate map[string]string      // templates which are common to multiple CRDs can be located in a shared folder and added via this type into the resulting CM/secret
	StringTemplate     map[string]string      // templates to render which are not accessable files, instead read by the caller from some other source, like a secret
	CustomData         map[string]string      // custom data which won't get rendered as a template and just added to the resulting cm/secret
	Labels             map[string]string      // labels to be set on the cm/secret
	Annotations        map[string]string      // Annotations set on cm/secret
	ConfigOptions      map[string]interface{} // map of parameters as input data to render the templates
	SkipSetOwner       bool                   // skip setting ownership on the associated configmap
	Version            string                 // optional version string to separate templates inside the InstanceType/Type directory. E.g. placementapi/config/18.0
	MultiTemplateDir   string                 // templates dir for multi-group operators, e.g. nova/api; requires InstanceType to be set
	CommonTemplates    []string               // list of common embedded templates to include (e.g. []string{"ssl.conf"}).
}

// GetTemplatesPath get path to templates, either running local or deployed as container
func GetTemplatesPath() (string, error) {

	templates := os.Getenv("OPERATOR_TEMPLATES")
	templatesPath := ""
	if templates == "" {
		// support local testing with 'up local'
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		templatesPath = path.Join(cwd, "/templates")
	} else {
		// deployed as a container
		templatesPath = templates
	}

	return templatesPath, nil
}

// GetAllTemplates - get all template files
//
// The structure of the folder is, base path, subdir, templateType, version
//   - path - base path of the templates folder
//   - subdir - directory under that root: one segment (legacy InstanceType, e.g. NovaAPI) or
//     multi-segment (e.g. nova/api from MultiTemplateDir)
//   - templateType - TType of the templates. When the templates got rendered and added to a CM
//     this information is e.g. used for the permissions they get mounted into the pod
//   - version - if there need to be templates for different versions, they can be stored in a version subdir
//
// Sub directories inside the specified directory with the above parameters get ignored.
func GetAllTemplates(path string, subdir string, templateType string, version string) []string {

	templatePath := filepath.Join(path, subdir, templateType, "*")

	if version != "" {
		templatePath = filepath.Join(path, subdir, templateType, version, "*")
	}

	templatesFiles, err := filepath.Glob(templatePath)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	// remove any subdiretories from templatesFiles
	for index := 0; index < len(templatesFiles); index++ {
		fi, err := os.Stat(templatesFiles[index])
		if err != nil {
			fmt.Print(err)
			os.Exit(1)
		}
		if fi.Mode().IsDir() {
			templatesFiles = RemoveIndex(templatesFiles, index)
			index = -1 // restart from the beginning
		}
	}

	return templatesFiles
}

// ExecuteTemplate creates a template from the file and
// execute it with the specified data
func ExecuteTemplate(templateFile string, data interface{}) (string, error) {

	b, err := os.ReadFile(templateFile)
	if err != nil {

		return "", err
	}

	file := string(b)

	renderedTemplate, err := ExecuteTemplateData(file, data)
	if err != nil {
		return "", err
	}
	return renderedTemplate, nil
}

// template function to indent the template with n tabs
func indent(n int, in string) string {
	var out string
	s := bufio.NewScanner(bytes.NewReader([]byte(in)))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		for i := 0; i < n; i++ {
			line = "\t" + line
		}
		out += line + "\n"
	}
	return out
}

// template function to remove empty lines if there are > n continuous empty lines
func removeNewLines(n int, in string) string {
	var out string
	s := bufio.NewScanner(bytes.NewReader([]byte(in)))

	// Variable to keep track of consecutive empty lines
	emptyLineCount := 0
	for s.Scan() {
		line := s.Text()

		if strings.TrimSpace(line) == "" {
			emptyLineCount++
			// If we have already seen more then n empty lines, skip this one
			if emptyLineCount > n {
				continue
			}
		} else {
			// Reset the empty line counter when we encounter a non-empty line
			emptyLineCount = 0
		}

		out += line + "\n"
	}
	return out
}

// This function removes extra space and new-lines from conf data.
func removeNewLinesInSections(in string) string {
	var out string
	s := bufio.NewScanner(bytes.NewReader([]byte(in)))

	for s.Scan() {
		line := strings.TrimSpace(s.Text())

		if line != "" {
			if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
				// new section-header
				if len(out) > 0 {
					out += "\n"
				}
			}

			out += line + "\n"
		}
	}

	return out
}

// template function to increment an int
func add(x, y int) int {
	return x + y
}

// template function to lower a string
func lower(s string) string {
	return strings.ToLower(s)
}

// ExecuteTemplateData creates a template from string and
// execute it with the specified data
func ExecuteTemplateData(templateData string, data interface{}) (string, error) {
	// template functions
	var tmpl *template.Template

	// template function which allows to execute a template from within
	// a template file.
	// name - name of the template as defined with `{{define "some-template"}}your template{{end}}
	// data - data to pass into to render the template for all can use `.`
	execTempl := func(name string, data interface{}) (string, error) {
		buf := &bytes.Buffer{}
		err := tmpl.ExecuteTemplate(buf, name, data)
		return buf.String(), err
	}

	var buff bytes.Buffer
	var err error
	funcs := template.FuncMap{
		"add":                      add,
		"execTempl":                execTempl,
		"indent":                   indent,
		"lower":                    lower,
		"removeNewLines":           removeNewLines,
		"removeNewLinesInSections": removeNewLinesInSections,
	}
	tmpl, err = template.New("tmp").Option("missingkey=error").Funcs(funcs).Parse(templateData)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(&buff, data)
	if err != nil {
		return "", err
	}
	return buff.String(), nil
}

// ExecuteTemplateFile - creates a template from the file and
// execute it with the specified data
func ExecuteTemplateFile(filename string, data interface{}) (string, error) {

	templates := os.Getenv("OPERATOR_TEMPLATES")
	filepath := ""
	if templates == "" {
		// support local testing with 'up local'
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		filepath = path.Join(cwd, "/templates/"+filename)
	} else {
		// deployed as a container
		filepath = path.Join(templates, filename)
	}

	b, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	file := string(b)

	return ExecuteTemplateData(file, data)
}

// GetTemplateData - Renders templates specified via Template struct
//
// Check the TType const and Template type for more details on defining the template.
func GetTemplateData(t Template) (map[string]string, error) {
	opts := t.ConfigOptions

	// get templates base path, either running local or deployed as container
	templatesPath, err := GetTemplatesPath()
	if err != nil {
		return nil, err
	}

	data := make(map[string]string)

	// Render requested common embedded templates as fallback defaults.
	// Note that local operator templates rendered below overwrite the common
	// ones: if they share the same filename, service operators definition take
	// precedence
	if len(t.CommonTemplates) > 0 {
		commonData, err := GetCommonTemplates(opts)
		if err != nil {
			return nil, err
		}
		for _, name := range t.CommonTemplates {
			v, ok := commonData[name]
			if !ok {
				return nil, fmt.Errorf("%w: %s", ErrCommonTemplateNotFound, name)
			}
			data[name] = v
		}
	}

	if t.Type != TemplateTypeNone {
		// If MultiTemplateDir is set but InstanceType is not, return an error
		// though we do not use InstanceType here, but it is used in secret.go
		if t.MultiTemplateDir != "" && t.InstanceType == "" {
			return nil, ErrInstanceTypeUnsetWithMultiTemplateDir
		}

		var templateSubdir string
		// if MultiTemplateDir is set, it will take precedence over InstanceType
		// otherwise, InstanceType is used to create the template subdir based on CRD name
		if t.MultiTemplateDir != "" {
			templateSubdir = t.MultiTemplateDir
		} else {
			templateSubdir = strings.ToLower(t.InstanceType)
		}
		if templateSubdir == "" {
			return nil, fmt.Errorf("%w: type is %q", ErrTemplateSubdirUnset, t.Type)
		}
		templatesFiles := GetAllTemplates(templatesPath, templateSubdir, string(t.Type), string(t.Version))

		// render all template files
		for _, file := range templatesFiles {
			renderedData, err := ExecuteTemplate(file, opts)
			if err != nil {
				return data, err
			}
			data[filepath.Base(file)] = renderedData
		}
	}
	// add additional template files from different directory, which
	// e.g. can be common to multiple controllers
	for filename, file := range t.AdditionalTemplate {
		renderedTemplate, err := ExecuteTemplateFile(file, opts)
		if err != nil {
			return nil, err
		}
		data[filename] = renderedTemplate
	}

	// render templates passed in as string via the StringTemplate
	for filename, tmplData := range t.StringTemplate {
		renderedTemplate, err := ExecuteTemplateData(tmplData, opts)

		if err != nil {
			return nil, err
		}
		data[filename] = renderedTemplate
	}

	return data, nil
}

// GetCommonTemplates renders the common config templates (ssl.conf, etc.)
// shipped with lib-common using the provided configOptions, and returns
// map[filename]renderedContent.
// Callers merge the result into their Template.CustomData before calling
// EnsureSecrets/EnsureConfigMaps so that common templates are included in the
// generated Secret/ConfigMaps without each operator duplicating the files.
func GetCommonTemplates(configOptions map[string]any) (map[string]string, error) {
	dir := "templates/common/config"
	entries, err := fs.ReadDir(commonTemplates, dir)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := fs.ReadFile(commonTemplates, path.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		rendered, err := ExecuteTemplateData(string(data), configOptions)
		if err != nil {
			return nil, err
		}
		result[e.Name()] = rendered
	}
	return result, nil
}

// TLSProfileConfigMap is the well-known ConfigMap the openstack-operator
// maintains from the cluster-wide TLS security profile on the OpenShift
// APIServer CR. Its keys (e.g. SSLCipherSuite, SSLProtocol) are merged as
// defaults into the ConfigOptions of every rendered Template, so service
// operators inherit the cluster TLS settings without any code of their own.
//
// It is a ConfigMap rather than a Secret on purpose: cipher suites and protocol
// lists are derived from a cluster-readable API object and carry nothing
// sensitive, so they should stay inspectable with `oc get cm`.
//
// Its lifecycle belongs entirely to the openstack-operator. lib-common only
// reads it: it is never created, updated or required here, and if it is absent
// every Template simply keeps the defaults it had before.
const TLSProfileConfigMap = "openstack-ssl-profile"

// DefaultTemplateConfigMaps - well-known ConfigMaps whose data is merged into
// template ConfigOptions as defaults by EnsureSecrets and EnsureConfigMaps.
// Earlier entries win over later ones; a caller's own ConfigOptions wins over
// all of them.
var DefaultTemplateConfigMaps = []string{TLSProfileConfigMap}

// ApplyTemplateDefaults returns a copy of tmpls in which the data of the named
// ConfigMaps has been merged underneath each Template's ConfigOptions. Keys
// already present in ConfigOptions are left untouched, so a value set explicitly
// by a service operator always wins over a cluster-wide default, and earlier
// ConfigMaps in the list win over later ones. Neither the given slice nor the
// caller's ConfigOptions maps are modified.
//
// A ConfigMap that does not exist contributes nothing: that is the normal case
// on clusters where no cluster-wide profile has been published, and on plain
// Kubernetes, leaving templates on their own hardcoded defaults. Any other read
// error is returned rather than skipped, because rendering a config with
// silently substituted fallbacks could downgrade settings the cluster
// administrator mandated.
func ApplyTemplateDefaults(
	ctx context.Context,
	h *helper.Helper,
	tmpls []Template,
	configMapNames []string,
) ([]Template, error) {
	out := make([]Template, len(tmpls))

	for i, t := range tmpls {
		// custom templates are not rendered, so they cannot consume defaults
		if t.Type == TemplateTypeCustom {
			out[i] = t
			continue
		}

		defaults, err := getTemplateDefaults(ctx, h, t.Namespace, configMapNames)
		if err != nil {
			return nil, err
		}

		if len(defaults) > 0 {
			merged := make(map[string]any, len(defaults)+len(t.ConfigOptions))
			for k, v := range defaults {
				merged[k] = v
			}
			for k, v := range t.ConfigOptions {
				merged[k] = v
			}
			t.ConfigOptions = merged
		}
		out[i] = t
	}

	return out, nil
}

// getTemplateDefaults reads the named ConfigMaps from namespace and returns
// their data as a single map, with earlier ConfigMaps in the list taking
// precedence over later ones.
func getTemplateDefaults(
	ctx context.Context,
	h *helper.Helper,
	namespace string,
	configMapNames []string,
) (map[string]any, error) {
	defaults := map[string]any{}

	for _, name := range configMapNames {
		cm := &corev1.ConfigMap{}
		err := h.GetClient().Get(ctx, types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		}, cm)
		if err != nil {
			if k8s_errors.IsNotFound(err) {
				continue
			}
			return nil, fmt.Errorf("error reading template defaults from ConfigMap %s/%s: %w", namespace, name, err)
		}

		for k, v := range cm.Data {
			if _, exists := defaults[k]; !exists {
				defaults[k] = v
			}
		}
	}

	return defaults, nil
}

// isDefaultTemplateConfigMap reports whether obj is one of the well-known
// ConfigMaps listed in DefaultTemplateConfigMaps. It matches on name only: a
// ConfigMap with a well-known name is relevant to the service CRs in its own
// namespace, which the watch set up by WatchDefaultTemplateConfigMap scopes
// separately.
func isDefaultTemplateConfigMap(obj client.Object) bool {
	for _, name := range DefaultTemplateConfigMaps {
		if obj.GetName() == name {
			return true
		}
	}
	return false
}

// WatchDefaultTemplateConfigMap returns an event handler for
//
//	ctrl.NewControllerManagedBy(mgr).Watches(&corev1.ConfigMap{}, ...)
//
// When one of the well-known ConfigMaps listed in DefaultTemplateConfigMaps
// is updated or deleted, every CR of the kind of cr in the ConfigMap's
// namespace is enqueued for reconciliation, so the service re-renders its
// templates against the new defaults. Events for other ConfigMaps are
// ignored. Only the kind of cr is used, not its contents.
func WatchDefaultTemplateConfigMap(c client.Client, cr client.Object) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(defaultTemplateConfigMapMapFunc(c, cr))
}

// defaultTemplateConfigMapMapFunc enqueues every CR of the kind of cr in the
// changed ConfigMap's namespace, but only when the ConfigMap is one of the
// well-known template-defaults ConfigMaps. It never reads the ConfigMap
// itself, so it also applies to deletion events.
func defaultTemplateConfigMapMapFunc(c client.Client, cr client.Object) handler.MapFunc {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		if !isDefaultTemplateConfigMap(obj) {
			return nil
		}
		gvk, err := apiutil.GVKForObject(cr, c.Scheme())
		if err != nil {
			return nil
		}
		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(gvk)
		if err := c.List(ctx, list, client.InNamespace(obj.GetNamespace())); err != nil {
			return nil
		}
		requests := make([]reconcile.Request, 0, len(list.Items))
		for _, item := range list.Items {
			requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			}})
		}
		return requests
	}
}
