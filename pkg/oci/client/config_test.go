// Copyright 2017 Oracle and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/oracle/oci-flexvolume-driver/pkg/oci/instancemeta"
)

func TestConfigDefaulting(t *testing.T) {
	expectedCompartmentOCID := "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	expectedRegion := "us-phoenix-1"
	expectedRegionKey := "phx"

	cfg := &Config{metadata: instancemeta.NewMock(
		&instancemeta.InstanceMetadata{
			CompartmentOCID: expectedCompartmentOCID,
			Region:          expectedRegionKey, // instance metadata API only returns the region key
		},
	)}

	err := cfg.setDefaults()
	if err != nil {
		t.Fatalf("cfg.setDefaults() => %v, expected no error", err)
	}

	if cfg.Auth.Region != expectedRegion {
		t.Fatalf("Expected cfg.Region = %q, got %q", cfg.Auth.Region, expectedRegion)
	}
	if cfg.Auth.RegionKey != expectedRegionKey {
		t.Fatalf("Expected cfg.RegionKey = %q, got %q", cfg.Auth.RegionKey, expectedRegionKey)
	}

	if cfg.Auth.CompartmentOCID != expectedCompartmentOCID {
		t.Fatalf("Expected cfg.CompartmentOCID = %q, got %q", cfg.Auth.CompartmentOCID, expectedCompartmentOCID)
	}
}

func TestConfigSetRegion(t *testing.T) {
	var testCases = []struct {
		in          string
		region      string
		shortRegion string
		shouldErr   bool
	}{
		{"us-phoenix-1", "us-phoenix-1", "phx", false},
		{"US-PHOENIX-1", "us-phoenix-1", "phx", false},
		{"phx", "us-phoenix-1", "phx", false},
		{"PHX", "us-phoenix-1", "phx", false},

		{"us-ashburn-1", "us-ashburn-1", "iad", false},
		{"US-ASHBURN-1", "us-ashburn-1", "iad", false},
		{"iad", "us-ashburn-1", "iad", false},
		{"IAD", "us-ashburn-1", "iad", false},

		{"eu-frankfurt-1", "eu-frankfurt-1", "fra", false},
		{"EU-FRANKFURT-1", "eu-frankfurt-1", "fra", false},
		{"fra", "eu-frankfurt-1", "fra", false},
		{"FRA", "eu-frankfurt-1", "fra", false},

		// error cases
		{"us-east", "", "", true},
		{"", "", "", true},
	}

	for _, tt := range testCases {
		t.Run(tt.in, func(t *testing.T) {
			cfg := &Config{}
			err := cfg.setRegionFields(tt.in)
			if err != nil {
				if !tt.shouldErr {
					t.Errorf("SetRegionFields(%q) unexpected error: %v", tt.in, err)
				}
			}

			if cfg.Auth.Region != tt.region {
				t.Errorf("SetRegionFields(%q) => {Region: %q}; want {Region: %q}", tt.in, cfg.Auth.Region, tt.region)
			}
			if cfg.Auth.RegionKey != tt.shortRegion {
				t.Errorf("SetRegionFields(%q) => {RegionShortName: %q}; want {RegionShortName: %q}", tt.in, cfg.Auth.RegionKey, tt.shortRegion)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	testCases := []struct {
		name string
		in   *Config
		errs field.ErrorList
	}{
		{
			name: "valid",
			in: &Config{
				Auth: AuthConfig{
					Region:          "us-phoenix-1",
					RegionKey:       "phx",
					CompartmentOCID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					TenancyOCID:     "ocid1.tennancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserOCID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:      "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:     "aa:bb:cc:dd:ee:ff:gg:hh:ii:jj:kk:ll:mm:nn:oo:pp",
					VcnOCID:         "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			errs: field.ErrorList{},
		}, {
			name: "missing_region",
			in: &Config{
				Auth: AuthConfig{
					RegionKey:       "phx",
					CompartmentOCID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					TenancyOCID:     "ocid1.tennancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserOCID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:      "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:     "aa:bb:cc:dd:ee:ff:gg:hh:ii:jj:kk:ll:mm:nn:oo:pp",
					VcnOCID:         "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.region", BadValue: ""},
			},
		}, {
			name: "missing_region_key",
			in: &Config{
				Auth: AuthConfig{
					Region:          "us-phoenix-1",
					CompartmentOCID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					TenancyOCID:     "ocid1.tennancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserOCID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:      "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:     "aa:bb:cc:dd:ee:ff:gg:hh:ii:jj:kk:ll:mm:nn:oo:pp",
					VcnOCID:         "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.region_key", BadValue: ""},
			},
		}, {
			name: "missing_tenancy_ocid",
			in: &Config{
				Auth: AuthConfig{
					Region:          "us-phoenix-1",
					RegionKey:       "phx",
					CompartmentOCID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserOCID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:      "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:     "aa:bb:cc:dd:ee:ff:gg:hh:ii:jj:kk:ll:mm:nn:oo:pp",
					VcnOCID:         "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.tenancy", BadValue: ""},
			},
		}, {
			name: "missing_user_ocid",
			in: &Config{
				Auth: AuthConfig{
					Region:          "us-phoenix-1",
					RegionKey:       "phx",
					CompartmentOCID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					TenancyOCID:     "ocid1.tennancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:      "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:     "aa:bb:cc:dd:ee:ff:gg:hh:ii:jj:kk:ll:mm:nn:oo:pp",
					VcnOCID:         "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.user", BadValue: ""},
			},
		}, {
			name: "missing_key_file",
			in: &Config{
				Auth: AuthConfig{
					Region:          "us-phoenix-1",
					RegionKey:       "phx",
					CompartmentOCID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					TenancyOCID:     "ocid1.tennancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserOCID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Fingerprint:     "aa:bb:cc:dd:ee:ff:gg:hh:ii:jj:kk:ll:mm:nn:oo:pp",
					VcnOCID:         "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.key", BadValue: ""},
			},
		}, {
			name: "missing_fingerprint",
			in: &Config{
				Auth: AuthConfig{
					Region:          "us-phoenix-1",
					RegionKey:       "phx",
					CompartmentOCID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					TenancyOCID:     "ocid1.tennancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserOCID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:      "-----BEGIN RSA PRIVATE KEY----- (etc)",
					VcnOCID:         "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.fingerprint", BadValue: ""},
			},
		}, {
			name: "missing_vcn",
			in: &Config{
				Auth: AuthConfig{
					Region:          "us-phoenix-1",
					RegionKey:       "phx",
					CompartmentOCID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					TenancyOCID:     "ocid1.tennancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserOCID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:      "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:     "aa:bb:cc:dd:ee:ff:gg:hh:ii:jj:kk:ll:mm:nn:oo:pp",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.vcn", BadValue: ""},
			},
		}, {
			name: "valid with instance principals enabled",
			in: &Config{
				UseInstancePrincipals: true,
				Auth: AuthConfig{
					VcnOCID:   "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					RegionKey: "phx",
				},
			},
			errs: field.ErrorList{},
		}, {
			name: "mixing instance principals with other auth flags",
			in: &Config{
				UseInstancePrincipals: true,
				Auth: AuthConfig{
					Region:      "us-phoenix-1",
					TenancyOCID: "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserOCID:    "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:  "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint: "aa:bb:cc:dd:ee:ff:gg:hh:ii:jj:kk:ll:mm:nn:oo:pp",
					VcnOCID:     "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					RegionKey:   "phx",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeForbidden, Field: "auth.region", Detail: "cannot be used when useInstancePrincipals is enabled", BadValue: ""},
				&field.Error{Type: field.ErrorTypeForbidden, Field: "auth.tenancy", Detail: "cannot be used when useInstancePrincipals is enabled", BadValue: ""},
				&field.Error{Type: field.ErrorTypeForbidden, Field: "auth.user", Detail: "cannot be used when useInstancePrincipals is enabled", BadValue: ""},
				&field.Error{Type: field.ErrorTypeForbidden, Field: "auth.key", Detail: "cannot be used when useInstancePrincipals is enabled", BadValue: ""},
				&field.Error{Type: field.ErrorTypeForbidden, Field: "auth.fingerprint", Detail: "cannot be used when useInstancePrincipals is enabled", BadValue: ""},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateConfig(tt.in)
			if !reflect.DeepEqual(result, tt.errs) {
				t.Errorf("ValidateConfig(%+v)\n=> %+v\nExpected: %+v", tt.in, result, tt.errs)
			}
		})
	}
}
