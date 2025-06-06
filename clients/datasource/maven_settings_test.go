// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datasource_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/osv-scalibr/clients/datasource"
)

func TestParseMavenSettings(t *testing.T) {
	t.Setenv("MAVEN_SETTINGS_TEST_USR", "UsErNaMe")
	t.Setenv("MAVEN_SETTINGS_TEST_PWD", "P455W0RD")
	t.Setenv("MAVEN_SETTINGS_TEST_SID", "my-cool-server")
	t.Setenv("MAVEN_SETTINGS_TEST_NIL", "")
	want := datasource.MavenSettingsXML{
		Servers: []datasource.MavenSettingsXMLServer{
			{
				ID:       "server1",
				Username: "user",
				Password: "pass",
			},
			{
				ID:       "server2",
				Username: "UsErNaMe",
				Password: "~~P455W0RD~~",
			},
			{
				ID:       "my-cool-server",
				Username: "${env.maven_settings_test_usr}-",
				Password: "${env.MAVEN_SETTINGS_TEST_BAD}",
			},
		},
	}

	got := datasource.ParseMavenSettings("./testdata/maven_settings/settings.xml")

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ParseMavenSettings() (-want +got):\n%s", diff)
	}
}

func TestMakeMavenAuth(t *testing.T) {
	globalSettings := datasource.MavenSettingsXML{
		Servers: []datasource.MavenSettingsXMLServer{
			{
				ID:       "global",
				Username: "global-user",
				Password: "global-pass",
			},
			{
				ID:       "overwrite1",
				Username: "original-user",
				Password: "original-pass",
			},
			{
				ID:       "overwrite2",
				Username: "user-to-be-deleted",
				// no password
			},
		},
	}
	userSettings := datasource.MavenSettingsXML{
		Servers: []datasource.MavenSettingsXMLServer{
			{
				ID:       "user",
				Username: "user",
				Password: "pass",
			},
			{
				ID:       "overwrite1",
				Username: "new-user",
				Password: "new-pass",
			},
			{
				ID: "overwrite2",
				// no username
				Password: "lone-password",
			},
		},
	}

	wantSupportedMethods := []datasource.HTTPAuthMethod{datasource.AuthDigest, datasource.AuthBasic}
	want := map[string]*datasource.HTTPAuthentication{
		"global": {
			SupportedMethods: wantSupportedMethods,
			AlwaysAuth:       false,
			Username:         "global-user",
			Password:         "global-pass",
		},
		"user": {
			SupportedMethods: wantSupportedMethods,
			AlwaysAuth:       false,
			Username:         "user",
			Password:         "pass",
		},
		"overwrite1": {
			SupportedMethods: wantSupportedMethods,
			AlwaysAuth:       false,
			Username:         "new-user",
			Password:         "new-pass",
		},
		"overwrite2": {
			SupportedMethods: wantSupportedMethods,
			AlwaysAuth:       false,
			Username:         "",
			Password:         "lone-password",
		},
	}

	got := datasource.MakeMavenAuth(globalSettings, userSettings)
	if diff := cmp.Diff(want, got, cmpopts.IgnoreUnexported(datasource.HTTPAuthentication{})); diff != "" {
		t.Errorf("MakeMavenAuth() (-want +got):\n%s", diff)
	}
}
