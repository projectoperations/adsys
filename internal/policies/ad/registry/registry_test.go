package registry_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ubuntu/adsys/internal/policies/ad/registry"
	"github.com/ubuntu/adsys/internal/policies/entry"
)

func TestDecodePolicy(t *testing.T) {
	t.Parallel()

	defaultKey := `Software/Canonical/Ubuntu/ValueName`
	defaultData := "BA"
	tests := map[string]struct {
		want    []entry.Entry
		wantErr bool
	}{
		"one element, string value": {
			want: []entry.Entry{
				{
					Key:   defaultKey,
					Value: defaultData,
				},
			}},
		"one element, decimal value": {
			want: []entry.Entry{
				{
					Key:   defaultKey,
					Value: "1234",
				},
			}},
		"one element, multitext value": {
			want: []entry.Entry{
				{
					Key:   defaultKey,
					Value: "B\nA",
				},
			}},
		"two elements": {
			want: []entry.Entry{
				{
					Key:   defaultKey,
					Value: "1",
				},
				{
					Key:   `Software/Policies/Canonical/Ubuntu/Directory UI/QueryLimit`,
					Value: "12345",
				},
			}},
		"one element, disabled": {
			want: []entry.Entry{
				{
					Key:      defaultKey,
					Value:    "",
					Disabled: true,
				},
			}},

		// basic type: no container, no children
		"basic type, enabled": {
			want: []entry.Entry{
				{
					Key:      `Software/Policies/Ubuntu/privilege/allow-local-admins/all`,
					Value:    "1",
					Disabled: false,
				},
			}},
		"basic type, disabled": {
			want: []entry.Entry{
				{
					Key:      `Software/Policies/Ubuntu/privilege/allow-local-admins/all`,
					Disabled: true,
				},
			}},

		// Container and options test cases
		"container with default elements override empty option values": {
			want: []entry.Entry{
				{
					Key:   `Software/Container/Child`,
					Value: "containerDefaultValueForChild",
				},
			}},
		"container with default elements are ignored on non empty option values": {
			want: []entry.Entry{
				{
					Key:   `Software/Container/Child`,
					Value: "MyValue",
				},
			}},
		"container with missing default element for option values have empty strings": {
			want: []entry.Entry{
				{
					Key:   `Software/Container/Child2`,
					Value: "",
				},
			}},
		"container with default elements are ignored on int option values (always have values)": {
			want: []entry.Entry{
				{
					Key:   `Software/Container/Child`,
					Value: "2",
				},
			}},
		// This ignores child value because container is disabled
		"disabled container with disabled option values": {
			want: []entry.Entry{
				{
					Key:      `Software/Container/Child`,
					Value:    "",
					Disabled: true,
				},
			}},
		// Both container and child are disabled
		"disabled container disables its option values": {
			want: []entry.Entry{
				{
					Key:      `Software/Container/Child`,
					Value:    "",
					Disabled: true,
				},
			}},
		"container with meta elements and default without value on options": {
			want: []entry.Entry{
				{
					Key:   `Software/Container/Child`,
					Value: "containerDefaultValueForChild",
					Meta:  "containerMetaValueForChild",
				},
			}},
		"container with meta elements and value on options": {
			want: []entry.Entry{
				{
					Key:   `Software/Container/Child`,
					Value: "MyValue",
					Meta:  "containerMetaValueForChild",
				},
			}},
		"one container with 2 children don’t mix their default values": {
			want: []entry.Entry{
				{
					Key:   `Software/Container1/Child1`,
					Value: "container1DefaultValueForChild1",
				},
				{
					Key:   `Software/Container1/Child2`,
					Value: "container1DefaultValueForChild2",
				},
			}},
		"two containers don’t mix their default values when redefined": {
			want: []entry.Entry{
				{
					Key:   `Software/Container1/Child1`,
					Value: "container1DefaultValueForChild1",
				},
				{
					Key:   `Software/Container1/Child2`,
					Value: "container1DefaultValueForChild2",
				},
				{
					Key:   `Software/Container2/Child1`,
					Value: "container2DefaultValueForChild1",
				},
				{
					Key: `Software/Container2/Child2`,
					// we didn't set default values for Child2 on Container2: keep empty (no leftover for Child1)
					Value: "",
				},
			}},
		"one container with 2 children don’t mix their meta values": {
			want: []entry.Entry{
				{
					Key:  `Software/Container1/Child1`,
					Meta: "container1MetaValueForChild1",
				},
				{
					Key:  `Software/Container1/Child2`,
					Meta: "container1MetaValueForChild2",
				},
			}},

		"semicolon in data": {
			want: []entry.Entry{
				{
					Key:   defaultKey,
					Value: "B;A",
				},
			}},

		"section separators in data": {
			want: []entry.Entry{
				{
					Key:   defaultKey,
					Value: "BA][C]",
				},
			}},
		"header only": {},

		"exotic return type":                  {wantErr: true},
		"invalid decimal value":               {wantErr: true},
		"invalid header, header doesnt match": {wantErr: true},
		"invalid header, header too short":    {wantErr: true},
		"invalid header, file truncated":      {wantErr: true},
		"invalid container default values":    {wantErr: true},
		"no header":                           {wantErr: true},
		"empty file":                          {wantErr: true},
		"section not closed":                  {wantErr: true},
		"missing field":                       {wantErr: true},
		"key is not utf16":                    {wantErr: true},
		"value is not utf16":                  {wantErr: true},
		"empty key":                           {wantErr: true},
		"empty value":                         {wantErr: true},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			name := name
			t.Parallel()

			f, err := os.Open(policyFilePath(name))
			if err != nil {
				t.Fatalf("Can't open registry file: %s", policyFilePath(name))
			}
			defer f.Close()

			rules, err := registry.DecodePolicy(f)
			if tc.wantErr {
				require.NotNil(t, err, "readPolicy returned no error when expecting one")
			} else {
				require.NoError(t, err, "readPolicy returned an error when expecting none")
			}

			require.Equalf(t, tc.want, rules, "expected value from readPolicy doesn't match")
		})
	}
}

func policyFilePath(name string) string {
	return filepath.Join("testdata", strings.ReplaceAll(strings.ReplaceAll(name, ",", "_"), " ", "_")+".pol")
}
