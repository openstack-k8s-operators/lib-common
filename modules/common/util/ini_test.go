package util // nolint:revive

import (
	"testing"

	. "github.com/onsi/gomega" // nolint:revive
)

var (
	defaultTestIniOption = IniOption{
		Section: "foo",
		Key:     "s3_store_cacert",
		Value:   "/etc/pki/tls/certs/ca-bundle.crt",
		Unique:  true,
	}

	aliasKeyIniOption = IniOption{
		Section: "pci",
		Key:     "alias",
		Value:   "{ \"device_type\": \"type-VF\", \"resource_class\": \"CUSTOM_A16_16A\", \"name\": \"A16_16A\" }",
		Unique:  false,
	}

	deviceSpecIniOption = IniOption{
		Section: "pci",
		Key:     "device_spec",
		Value:   "{ \"vendor_id\": \"10de\", \"product_id\": \"25b6\", \"address\": \"0000:25:00.6\", \"resource_class\": \"CUSTOM_A16_8A\", \"managed\": \"yes\" }",
		Unique:  false,
	}
)

var tests = []struct {
	name     string
	input    string
	option   IniOption
	expected string
	err      string
}{
	{
		name:     "empty customServiceConfig",
		input:    "",
		option:   defaultTestIniOption,
		expected: "",
		err:      "",
	},
	{
		name: "field to non-existing section",
		input: `[DEFAULT]
debug=true
enabled_backends = backend:s3
[foo]
bar = bar
foo = foo
[backend]
option1 = value1`,
		option: IniOption{
			Section: "bar",
			Key:     "foo",
			Value:   "foo",
		},
		expected: `[DEFAULT]
debug=true
enabled_backends = backend:s3
[foo]
bar = bar
foo = foo
[backend]
option1 = value1`,
		err: "could not patch target section: bar",
	},
	{
		name: "add new ini line to a section in the middle",
		input: `[DEFAULT]
debug=true
enabled_backends = backend:s3
[foo]
bar = bar
foo = foo
[backend]
option1 = value1`,
		option: defaultTestIniOption,
		expected: `[DEFAULT]
debug=true
enabled_backends = backend:s3
[foo]
s3_store_cacert = /etc/pki/tls/certs/ca-bundle.crt
bar = bar
foo = foo
[backend]
option1 = value1`,
		err: "",
	},
	{
		name: "section is not found, return it unchanged",
		input: `[DEFAULT]
debug=true
enabled_backends = backend:s3
[backend]
s3_store_cacert = /etc/pki/tls/certs/ca-bundle.crt `,
		option: defaultTestIniOption,
		expected: `[DEFAULT]
debug=true
enabled_backends = backend:s3
[backend]
s3_store_cacert = /etc/pki/tls/certs/ca-bundle.crt `,
		err: "could not patch target section: foo",
	},
	{
		name: "Add option to a section at the very beginning of customServiceConfig",
		input: `[foo]
bar = bar
foo = foo
[backend]
option1 = value1 `,
		option: defaultTestIniOption,
		expected: `[foo]
s3_store_cacert = /etc/pki/tls/certs/ca-bundle.crt
bar = bar
foo = foo
[backend]
option1 = value1 `,
		err: "",
	},
	{
		name: "Add option to a section at the very bottom of customServiceConfig",
		input: `[DEFAULT]
debug=true
enabled_backends = backend:s3
[backend]
# this is a comment
option1 = value1
[foo]
# this is a comment
bar = bar
foo = foo`,
		option: defaultTestIniOption,
		expected: `[DEFAULT]
debug=true
enabled_backends = backend:s3
[backend]
# this is a comment
option1 = value1
[foo]
s3_store_cacert = /etc/pki/tls/certs/ca-bundle.crt
# this is a comment
bar = bar
foo = foo`,
		err: "",
	},
	{
		name: "Add option to an empty target section",
		input: `[DEFAULT]
debug=true
[foo]`,
		option: defaultTestIniOption,
		expected: `[DEFAULT]
debug=true
[foo]
s3_store_cacert = /etc/pki/tls/certs/ca-bundle.crt`,
		err: "",
	},
	{
		name: "key/value already present in the target section",
		input: `[DEFAULT]
debug=true
[foo]
s3_store_cacert = /my/custom/path/ca-bundle.crt`,
		option: defaultTestIniOption,
		expected: `[DEFAULT]
debug=true
[foo]
s3_store_cacert = /my/custom/path/ca-bundle.crt`,
		err: "key already exists in section: key s3_store_cacert in section foo",
	},
	{
		name: "add new ini line anyway even though a section contains the same key ",
		input: `[pci]
device_spec = { "vendor_id": "10de", "product_id": "25b6", "address": "0000:25:00.4", "resource_class": "CUSTOM_A16_16A", "managed": "no" }
device_spec = { "vendor_id": "10de", "product_id": "25b6", "address": "0000:25:00.5", "resource_class": "CUSTOM_A16_8A", "managed": "no" }
alias = { "device_type": "type-VF", "resource_class": "CUSTOM_A16_8A", "name": "A16_8A" }`,
		option: aliasKeyIniOption,
		expected: `[pci]
alias = { "device_type": "type-VF", "resource_class": "CUSTOM_A16_16A", "name": "A16_16A" }
device_spec = { "vendor_id": "10de", "product_id": "25b6", "address": "0000:25:00.4", "resource_class": "CUSTOM_A16_16A", "managed": "no" }
device_spec = { "vendor_id": "10de", "product_id": "25b6", "address": "0000:25:00.5", "resource_class": "CUSTOM_A16_8A", "managed": "no" }
alias = { "device_type": "type-VF", "resource_class": "CUSTOM_A16_8A", "name": "A16_8A" }`,
		err: "",
	},
	{
		name: "add new ini line anyway even though a section contains the same key ",
		input: `[pci]
device_spec = { "vendor_id": "10de", "product_id": "25b6", "address": "0000:25:00.4", "resource_class": "CUSTOM_A16_16A", "managed": "no" }
device_spec = { "vendor_id": "10de", "product_id": "25b6", "address": "0000:25:00.5", "resource_class": "CUSTOM_A16_8A", "managed": "no" }
alias = { "device_type": "type-VF", "resource_class": "CUSTOM_A16_16A", "name": "A16_16A" }
alias = { "device_type": "type-VF", "resource_class": "CUSTOM_A16_8A", "name": "A16_8A" }`,
		option: deviceSpecIniOption,
		expected: `[pci]
device_spec = { "vendor_id": "10de", "product_id": "25b6", "address": "0000:25:00.6", "resource_class": "CUSTOM_A16_8A", "managed": "yes" }
device_spec = { "vendor_id": "10de", "product_id": "25b6", "address": "0000:25:00.4", "resource_class": "CUSTOM_A16_16A", "managed": "no" }
device_spec = { "vendor_id": "10de", "product_id": "25b6", "address": "0000:25:00.5", "resource_class": "CUSTOM_A16_8A", "managed": "no" }
alias = { "device_type": "type-VF", "resource_class": "CUSTOM_A16_16A", "name": "A16_16A" }
alias = { "device_type": "type-VF", "resource_class": "CUSTOM_A16_8A", "name": "A16_8A" }`,
		err: "",
	},
}

func TestExtendCustomServiceConfig(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			output, err := ExtendCustomServiceConfig(tt.input, tt.option)
			g.Expect(output).To(Equal(tt.expected))
			if err != nil {
				// check the string matches the expected error message
				g.Expect(err.Error()).To(Equal(tt.err))
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
		})
	}
}
