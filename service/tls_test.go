/*
Copyright (C)  2018 Yahoo Japan Corporation Athenz team.

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
package service

import (
	"crypto/tls"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/yahoojapan/athenz-client-sidecar/config"
)

func TestNewTLSConfig(t *testing.T) {
	type args struct {
		CertKeyPath string
		KeyKeyPath  string
		CAKeyPath   string
		cfg         config.TLS
	}
	type test struct {
		name       string
		args       args
		want       *tls.Config
		beforeFunc func(args args)
		afterFunc  func(args args)
		checkFunc  func(*tls.Config, *tls.Config) error
		wantErr    error
	}

	defaultArgs := args{
		CertKeyPath: "./assets/dummyServer.crt",
		KeyKeyPath:  "./assets/dummyServer.key",
		CAKeyPath:   "./assets/dummyCa.pem",
		cfg: config.TLS{
			Cert: "_test_cert_",
			Key:  "_test_key_",
			CA:   "_test_ca_",
		},
	}

	tests := []test{
		{
			name: "Check the minimum SSL/TLS version",
			args: defaultArgs,
			want: &tls.Config{
				MinVersion: tls.VersionTLS12,
				// CurvePreferences: []tls.CurveID{
				// 	tls.CurveP521,
				// 	tls.CurveP384,
				// 	tls.CurveP256,
				// 	tls.X25519,
				// },
				// SessionTicketsDisabled: true,
				// Certificates: func() []tls.Certificate {
				// 	cert, _ := tls.LoadX509KeyPair(defaultArgs.CertKeyPath, defaultArgs.KeyKeyPath)
				// 	return []tls.Certificate{cert}
				// }(),
				// ClientAuth: tls.RequireAndVerifyClientCert,
			},
			beforeFunc: func(args args) {
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"), args.CertKeyPath)
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"), args.KeyKeyPath)
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"), args.CAKeyPath)
			},
			afterFunc: func(args args) {
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"))
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"))
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"))
			},
			checkFunc: func(got, want *tls.Config) error {
				if got.MinVersion != want.MinVersion {
					return fmt.Errorf("MinVersion unmatched: got: %d  want: %d", got.MinVersion, want.MinVersion)
				}
				return nil
			},
		},
		{
			name: "Check values of 'CurvePreferences'",
			args: defaultArgs,
			want: &tls.Config{
				// MinVersion: tls.VersionTLS12,
				CurvePreferences: []tls.CurveID{
					tls.CurveP256,
					tls.CurveP384,
					tls.CurveP521,
					tls.X25519,
				},
				// SessionTicketsDisabled: true,
				// Certificates: func() []tls.Certificate {
				// 	cert, _ := tls.LoadX509KeyPair(defaultArgs.CertKeyPath, defaultArgs.KeyKeyPath)
				// 	return []tls.Certificate{cert}
				// }(),
				// ClientAuth: tls.RequireAndVerifyClientCert,
			},
			beforeFunc: func(args args) {
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"), args.CertKeyPath)
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"), args.KeyKeyPath)
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"), args.CAKeyPath)
			},
			afterFunc: func(args args) {
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"))
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"))
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"))
			},
			checkFunc: func(got, want *tls.Config) error {
				if len(got.CurvePreferences) != len(want.CurvePreferences) {
					return fmt.Errorf("The length of CurvePreferences not matched: got: %d  want: %d", len(got.CurvePreferences), len(want.CurvePreferences))
				}

				for _, actual := range got.CurvePreferences {
					var hasValue bool
					for _, expected := range want.CurvePreferences {
						if actual == expected {
							hasValue = true
							break
						}
					}
					if !hasValue {
						return fmt.Errorf("The actual value of CurvePreferences not found in the expected: got: %v  want: %v", got.CurvePreferences, want.CurvePreferences)
					}
				}

				return nil
			},
		},
		{
			name: "Check whether the value of 'SessionTicketsDisabled' is true",
			args: defaultArgs,
			want: &tls.Config{
				// MinVersion: tls.VersionTLS12,
				// CurvePreferences: []tls.CurveID{
				// 	tls.CurveP256,
				// 	tls.CurveP384,
				// 	tls.CurveP521,
				// 	tls.X25519,
				// },
				SessionTicketsDisabled: true,
				// Certificates: func() []tls.Certificate {
				// 	cert, _ := tls.LoadX509KeyPair(defaultArgs.CertKeyPath, defaultArgs.KeyKeyPath)
				// 	return []tls.Certificate{cert}
				// }(),
				// ClientAuth: tls.RequireAndVerifyClientCert,
			},
			beforeFunc: func(args args) {
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"), args.CertKeyPath)
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"), args.KeyKeyPath)
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"), args.CAKeyPath)
			},
			afterFunc: func(args args) {
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"))
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"))
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"))
			},
			checkFunc: func(got, want *tls.Config) error {
				if got.SessionTicketsDisabled != want.SessionTicketsDisabled {
					return fmt.Errorf("SessionTicketsDisabled is not as same as wanted: got: %t  want: %t", got.SessionTicketsDisabled, want.SessionTicketsDisabled)
				}
				return nil
			},
		},
		{
			name: "Check certificates",
			args: defaultArgs,
			want: &tls.Config{
				// MinVersion: tls.VersionTLS12,
				// CurvePreferences: []tls.CurveID{
				// 	tls.CurveP256,
				// 	tls.CurveP384,
				// 	tls.CurveP521,
				// 	tls.X25519,
				// },
				// SessionTicketsDisabled: true,
				Certificates: func() []tls.Certificate {
					cert, _ := tls.LoadX509KeyPair(defaultArgs.CertKeyPath, defaultArgs.KeyKeyPath)
					return []tls.Certificate{cert}
				}(),
				// ClientAuth: tls.RequireAndVerifyClientCert,
			},
			beforeFunc: func(args args) {
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"), args.CertKeyPath)
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"), args.KeyKeyPath)
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"), args.CAKeyPath)
			},
			afterFunc: func(args args) {
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"))
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"))
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"))
			},
			checkFunc: func(got, want *tls.Config) error {
				if !reflect.DeepEqual(got.Certificates, want.Certificates) {
					return fmt.Errorf("Certificates unmatched: got: %v  want: %v", got.Certificates, want.Certificates)
				}
				return nil
			},
		},
		{
			name: "Check whether ClientAuth is 'RequireAndVerifyClientCert'",
			args: defaultArgs,
			want: &tls.Config{
				// MinVersion: tls.VersionTLS12,
				// CurvePreferences: []tls.CurveID{
				// 	tls.CurveP256,
				// 	tls.CurveP384,
				// 	tls.CurveP521,
				// 	tls.X25519,
				// },
				// SessionTicketsDisabled: true,
				// Certificates: func() []tls.Certificate {
				// 	cert, _ := tls.LoadX509KeyPair(defaultArgs.CertKeyPath, defaultArgs.KeyKeyPath)
				// 	return []tls.Certificate{cert}
				// }(),
				ClientAuth: tls.RequireAndVerifyClientCert,
			},
			beforeFunc: func(args args) {
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"), args.CertKeyPath)
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"), args.KeyKeyPath)
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"), args.CAKeyPath)
			},
			afterFunc: func(args args) {
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"))
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"))
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"))
			},
			checkFunc: func(got, want *tls.Config) error {
				if got.ClientAuth != want.ClientAuth {
					return fmt.Errorf("ClientAuth is not 'RequireAndVerifyClientCert': got: %d  want: %d", got.ClientAuth, want.ClientAuth)
				}
				return nil
			},
		},
		{
			name: "Request without cert file",
			args: defaultArgs,
			beforeFunc: func(args args) {
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"), args.KeyKeyPath)
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"), args.CAKeyPath)
			},
			afterFunc: func(args args) {
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"))
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"))
			},
			wantErr: fmt.Errorf("Cert/Key path not found"),
		},
		{
			name: "Request without key file",
			args: defaultArgs,
			beforeFunc: func(args args) {
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"), args.CertKeyPath)
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"), args.CAKeyPath)
			},
			afterFunc: func(args args) {
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"))
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"))
			},
			wantErr: fmt.Errorf("Cert/Key path not found"),
		},
		{
			name: "Check whether ClientAuth is 'NoClientCert' when requesting without CA file",
			args: defaultArgs,
			want: &tls.Config{
				// MinVersion: tls.VersionTLS12,
				// CurvePreferences: []tls.CurveID{
				// 	tls.CurveP256,
				// 	tls.CurveP384,
				// 	tls.CurveP521,
				// 	tls.X25519,
				// },
				// SessionTicketsDisabled: true,
				// Certificates: func() []tls.Certificate {
				// 	cert, _ := tls.LoadX509KeyPair(defaultArgs.CertKeyPath, defaultArgs.KeyKeyPath)
				// 	return []tls.Certificate{cert}
				// }(),
				ClientAuth: tls.NoClientCert,
			},
			beforeFunc: func(args args) {
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"), args.CertKeyPath)
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"), args.KeyKeyPath)
			},
			afterFunc: func(args args) {
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"))
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"))
			},
			checkFunc: func(got, want *tls.Config) error {
				if got.ClientAuth != want.ClientAuth {
					return fmt.Errorf("ClientAuth is not 'NoClientCert': got: %d  want: %d", got.ClientAuth, want.ClientAuth)
				}
				return nil
			},
		},
		{
			name: "Request with invalid cert file",
			args: defaultArgs,
			beforeFunc: func(args args) {
				args.CertKeyPath = "./assets/invalid_dummyServer.crt"
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"), args.CertKeyPath)
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"), args.KeyKeyPath)
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"), args.CAKeyPath)
			},
			afterFunc: func(args args) {
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"))
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"))
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"))
			},
			wantErr: fmt.Errorf("tls: failed to find any PEM data in certificate input"),
		},
		{
			name: "Request with invalid key file",
			args: defaultArgs,
			beforeFunc: func(args args) {
				args.KeyKeyPath = "./assets/invalid_dummyServer.key"
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"), args.CertKeyPath)
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"), args.KeyKeyPath)
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"), args.CAKeyPath)
			},
			afterFunc: func(args args) {
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"))
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"))
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"))
			},
			wantErr: fmt.Errorf("tls: failed to find any PEM data in key input"),
		},
		{
			name: "Request with invalid CA file",
			args: defaultArgs,
			beforeFunc: func(args args) {
				args.CAKeyPath = "./assets/invalid_dummyCa.pem"
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"), args.CertKeyPath)
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"), args.KeyKeyPath)
				os.Setenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"), args.CAKeyPath)
			},
			afterFunc: func(args args) {
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Cert, "_"), "_"))
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.Key, "_"), "_"))
				os.Unsetenv(strings.TrimPrefix(strings.TrimSuffix(args.cfg.CA, "_"), "_"))
			},
			wantErr: fmt.Errorf("Certification Failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				tt.beforeFunc(tt.args)
			}
			if tt.afterFunc != nil {
				defer tt.afterFunc(tt.args)
			}

			got, err := NewTLSConfig(tt.args.cfg)

			if tt.wantErr == nil && err != nil {
				t.Errorf("NewTLSConfig() error: %v  wantErr: %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Error should occur: want error: %v  want: %v", err, tt.wantErr)
					return
				}
				// Here is comparing error message with expected
				if err.Error() != tt.wantErr.Error() {
					t.Errorf("Assertion failed: got: %v  want: %v", err, tt.wantErr)
					return
				}
			}

			if tt.checkFunc != nil {
				err = tt.checkFunc(got, tt.want)
				if err != nil {
					t.Errorf("NewTLSConfig() error = %v", err)
					return
				}
			}
		})
	}
}

func TestNewX509CertPool(t *testing.T) {
	type args struct {
		path string
	}
	type test struct {
		name    string
		args    args
		wantErr error
	}

	tests := []test{
		{
			name: "Get CA cert pool",
			args: args{
				path: "./assets/dummyCa.pem",
			},
		},
		{
			name: "Missing CA file",
			args: args{
				path: "/tmp/CAfilenotfound.pem",
			},
			wantErr: fmt.Errorf("no such file or directory"),
		},
		{
			name: "Argument not specified",
			args: args{
				path: "",
			},
			wantErr: fmt.Errorf("no such file or directory"),
		},
		{
			name: "Request with invalid CA file",
			args: args{
				path: "./assets/invalid_dummyCa.pem",
			},
			wantErr: fmt.Errorf("Certification Failed"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewX509CertPool(tt.args.path)
			if err != nil && tt.wantErr == nil {
				t.Errorf("NewX509CertPool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("want error: %v  got: %v", tt.wantErr, err)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr.Error()) {
					t.Errorf("NewX509CertPool() error: %v, want: %v", err, tt.wantErr)
					return
				}
			} else {
				if got == nil {
					t.Errorf("CertPool should not be empty: got: %v", got)
					return
				}
			}
		})
	}
}
