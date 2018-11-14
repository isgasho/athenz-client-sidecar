package usecase

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	ntokend "ghe.corp.yahoo.co.jp/athenz/athenz-ntokend"
	"ghe.corp.yahoo.co.jp/athenz/athenz-tenant-sidecar/config"
	"ghe.corp.yahoo.co.jp/athenz/athenz-tenant-sidecar/handler"
	"ghe.corp.yahoo.co.jp/athenz/athenz-tenant-sidecar/infra"
	"ghe.corp.yahoo.co.jp/athenz/athenz-tenant-sidecar/router"
	"ghe.corp.yahoo.co.jp/athenz/athenz-tenant-sidecar/service"
)

func TestNew(t *testing.T) {
	type args struct {
		cfg config.Config
	}
	type test struct {
		name       string
		args       args
		beforeFunc func()
		checkFunc  func(Tenant, Tenant) error
		afterFunc  func()
		want       Tenant
		wantErr    error
	}
	tests := []test{
		{
			name: "Check error when new token service",
			args: args{
				cfg: config.Config{
					Token: config.Token{},
				},
			},
			wantErr: fmt.Errorf("invalid token refresh duration , time: invalid duration "),
		},
		func() test {
			keyKey := "dummyKey"
			key := "./assets/dummyServer.key"
			cfg := config.Config{
				Token: config.Token{
					AthenzDomain:    keyKey,
					ServiceName:     keyKey,
					PrivateKeyPath:  "_" + keyKey + "_",
					ValidateToken:   false,
					RefreshDuration: "1m",
					KeyVersion:      "1",
					Expiration:      "1m",
					NTokenPath:      "",
				},
				Server: config.Server{
					HealthzPath: "/dummyPath",
				},
			}

			return test{
				name: "Check success",
				args: args{
					cfg: cfg,
				},
				want: func() Tenant {
					os.Setenv(keyKey, key)
					defer os.Unsetenv(keyKey)
					token, err := createNtokend(cfg.Token)
					if err != nil {
						panic(err)
					}
					role := service.NewRoleService(cfg.Role, token.GetTokenProvider())

					serveMux := router.New(cfg.Server, handler.New(cfg.Proxy, infra.NewBuffer(cfg.Proxy.BufferSize), token.GetTokenProvider(), role.GetRoleProvider()))
					server := service.NewServer(cfg.Server, serveMux)

					return &tenantd{
						cfg:    cfg,
						token:  token,
						server: server,
						role:   role,
					}
				}(),
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				tt.beforeFunc()
			}
			if tt.afterFunc != nil {
				defer tt.afterFunc()
			}

			got, err := New(tt.args.cfg)
			if tt.wantErr == nil && err != nil {
				t.Errorf("failed to instantiate, err: %v", err)
				return
			} else if tt.wantErr != nil {
				if tt.wantErr.Error() != err.Error() {
					t.Errorf("error not the same, want: %v, got: %v", tt.wantErr, err)
				}
			}

			if tt.checkFunc != nil {
				err = tt.checkFunc(got, tt.want)
				if tt.wantErr == nil && err != nil {
					t.Errorf("compare check failed, err: %v", err)
					return
				}
			}
		})
	}
}

func Test_tenantd_Start(t *testing.T) {
	type fields struct {
		cfg    config.Config
		token  ntokend.TokenService
		server service.Server
		role   service.RoleService
	}
	type args struct {
		ctx context.Context
	}
	type test struct {
		name       string
		fields     fields
		args       args
		beforeFunc func() error
		checkFunc  func(chan []error, []error) error
		afterFunc  func()
		want       []error
	}
	tests := []test{
		func() test {
			keyKey := "dummyKey"
			key := "./assets/dummyServer.key"

			certKey := "dummy_cert"
			cert := "./assets/dummyServer.crt"

			cfg := config.Config{
				Token: config.Token{
					AthenzDomain:    keyKey,
					ServiceName:     keyKey,
					PrivateKeyPath:  "_" + keyKey + "_",
					ValidateToken:   false,
					RefreshDuration: "1m",
					KeyVersion:      "1",
					Expiration:      "1m",
					NTokenPath:      "",
				},
				Server: config.Server{
					HealthzPath: "/dummyPath",
					TLS: config.TLS{
						Enabled: true,
						CertKey: certKey,
						KeyKey:  keyKey,
					},
				},
			}

			ctx, cancelFunc := context.WithCancel(context.Background())

			os.Setenv(keyKey, key)
			os.Setenv(certKey, cert)

			return test{
				name: "Token updater works",
				fields: func() fields {
					token, err := createNtokend(cfg.Token)
					if err != nil {
						panic(err)
					}
					role := service.NewRoleService(cfg.Role, token.GetTokenProvider())

					serveMux := router.New(cfg.Server, handler.New(cfg.Proxy, infra.NewBuffer(cfg.Proxy.BufferSize), token.GetTokenProvider(), role.GetRoleProvider()))
					server := service.NewServer(cfg.Server, serveMux)

					return fields{
						cfg:    cfg,
						token:  token,
						server: server,
						role:   role,
					}
				}(),
				args: args{
					ctx: ctx,
				},
				checkFunc: func(got chan []error, want []error) error {
					time.Sleep(time.Millisecond * 200)
					cancelFunc()
					time.Sleep(time.Millisecond * 200)

					gotErr := <-got
					if !reflect.DeepEqual(gotErr, want) {
						return fmt.Errorf("Got: %v, want: %v", gotErr, want)
					}
					return nil
				},
				afterFunc: func() {
					os.Unsetenv(keyKey)
				},
				want: []error{context.Canceled},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.afterFunc != nil {
				defer tt.afterFunc()
			}
			if tt.beforeFunc != nil {
				if err := tt.beforeFunc(); err != nil {
					t.Errorf("Error : %v", err)
					return
				}
			}

			te := &tenantd{
				cfg:    tt.fields.cfg,
				token:  tt.fields.token,
				server: tt.fields.server,
				role:   tt.fields.role,
			}
			got := te.Start(tt.args.ctx)
			if err := tt.checkFunc(got, tt.want); err != nil {
				t.Errorf("Start function error: %v", err)
			}
		})
	}
}

func Test_createNtokend(t *testing.T) {
	type args struct {
		cfg config.Token
	}
	type test struct {
		name       string
		args       args
		beforeFunc func()
		checkFunc  func(got, want ntokend.TokenService) error
		afterFunc  func()
		want       ntokend.TokenService
		wantErr    error
	}
	tests := []test{
		{
			name: "refresh duration invalid",
			args: args{
				cfg: config.Token{
					RefreshDuration: "dummy",
				},
			},
			wantErr: fmt.Errorf("invalid token refresh duration %s, %v", "dummy", "time: invalid duration dummy"),
		},
		{
			name: "token expiration invalid",
			args: args{
				cfg: config.Token{
					RefreshDuration: "1s",
					Expiration:      "dummy",
				},
			},
			wantErr: fmt.Errorf("invalid token expiration %s, %v", "dummy", "time: invalid duration dummy"),
		},
		func() test {
			keyKey := "dummyKey"
			key := "notexists"

			return test{
				name: "Test error private key not exist",
				args: func() args {
					return args{
						cfg: config.Token{
							RefreshDuration: "1m",
							Expiration:      "1m",
							PrivateKeyPath:  "_" + keyKey + "_",
						},
					}
				}(),
				beforeFunc: func() {
					os.Setenv(keyKey, key)
				},
				afterFunc: func() {
					os.Unsetenv(keyKey)
				},
				wantErr: fmt.Errorf("invalid token certificate open %v", "notexists: no such file or directory"),
			}
		}(),
		func() test {
			keyKey := "dummyKey"
			key := "./assets/invalid_dummyServer.key"

			return test{
				name: "Test error private key not valid",
				args: func() args {

					return args{
						cfg: config.Token{
							RefreshDuration: "1m",
							Expiration:      "1m",
							PrivateKeyPath:  "_" + keyKey + "_",
							NTokenPath:      "",
						},
					}
				}(),
				beforeFunc: func() {
					os.Setenv(keyKey, key)
				},
				afterFunc: func() {
					os.Unsetenv(keyKey)
				},
				wantErr: fmt.Errorf(`failed to create ZMS SVC Token Builder
AthenzDomain:	
ServiceName:	
KeyVersion:	
Error: Unable to create signer: Unable to load private key`),
			}
		}(),
		func() test {
			keyKey := "dummyKey"
			key := "./assets/dummyServer.key"
			cfg := config.Token{
				AthenzDomain:    keyKey,
				ServiceName:     keyKey,
				NTokenPath:      "",
				PrivateKeyPath:  "_" + keyKey + "_",
				ValidateToken:   false,
				RefreshDuration: "1s",
				KeyVersion:      "1",
				Expiration:      "1s",
			}
			keyData, _ := ioutil.ReadFile(key)
			athenzDomain := config.GetActualValue(cfg.AthenzDomain)
			serviceName := config.GetActualValue(cfg.ServiceName)

			return test{
				name: "Check return value",
				args: args{
					cfg: cfg,
				},
				want: func() ntokend.TokenService {
					tok, err := ntokend.New(
						ntokend.RefreshDuration(time.Second), ntokend.TokenExpiration(time.Second), ntokend.KeyVersion(cfg.KeyVersion), ntokend.KeyData(keyData), ntokend.TokenFilePath(cfg.NTokenPath),
						ntokend.AthenzDomain(athenzDomain), ntokend.ServiceName(serviceName))

					if err != nil {
						panic(err)
					}

					return tok
				}(),
				beforeFunc: func() {
					os.Setenv(keyKey, key)
				},
				checkFunc: func(got, want ntokend.TokenService) error {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					got.StartTokenUpdater(ctx)
					want.StartTokenUpdater(ctx)
					time.Sleep(time.Millisecond * 50)

					g, err := got.GetToken()
					if err != nil {
						return fmt.Errorf("Got not found, err: %v", err)
					}
					w, err := want.GetToken()
					if err != nil {
						return fmt.Errorf("Want not found, err: %v", err)
					}
					parse := func(str string) map[string]string {
						m := make(map[string]string)
						for _, pair := range strings.Split(str, ";") {
							kv := strings.SplitN(pair, "=", 2)
							if len(kv) < 2 {
								continue
							}
							m[kv[0]] = kv[1]
						}
						return m
					}

					gm := parse(g)
					wm := parse(w)

					check := func(key string) bool {
						return gm[key] != wm[key]
					}

					if check("v") || check("d") || check("n") || check("k") || check("h") || check("i") {
						return fmt.Errorf("invalid token, got: %s, want: %s", g, w)
					}

					return nil
				},
				afterFunc: func() {
					os.Unsetenv(keyKey)
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				tt.beforeFunc()
			}
			if tt.afterFunc != nil {
				defer tt.afterFunc()
			}

			got, err := createNtokend(tt.args.cfg)
			if err != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("createNtokend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkFunc != nil {
				err = tt.checkFunc(got, tt.want)
				if tt.wantErr == nil && err != nil {
					t.Errorf("compare check failed, err: %v", err)
					return
				}
			}
		})
	}
}