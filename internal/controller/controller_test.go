package controller

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/nicjohnson145/hlp"
	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
	"github.com/nicjohnson145/plantr/internal/interceptors"
	"github.com/nicjohnson145/plantr/internal/parsingv2"
	"github.com/nicjohnson145/plantr/internal/token"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"
)

func newControllerWithConfig(t *testing.T, conf ControllerConfig, repoConfig *parsingv2.Config) *Controller {
	t.Helper()

	if conf.GitClient == nil {
		conf.GitClient = NewMockGitClient(t)
	}
	if conf.StorageClient == nil {
		conf.StorageClient = NewMockStorageClient(t)
	}
	if conf.JWTDuration.Seconds() == 0 {
		conf.JWTDuration = 10 * 24 * time.Hour
	}
	if conf.JWTSigningKey == nil {
		conf.JWTSigningKey = []byte(`some-signing-key`)
	}

	ctrl, err := NewController(conf)
	require.NoError(t, err)

	ctrl.config = repoConfig

	return ctrl
}

func seedsEqual(t *testing.T, want []*parsingv2.Seed, got []*parsingv2.Seed) {
	t.Helper()

	opts := []cmp.Option{
		cmpopts.IgnoreFields(parsingv2.Seed{}, "Hash"),
	}

	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Logf("Mismatch (-want +got):\n%s", diff)
		t.FailNow()
	}
}

func pbEqual(t *testing.T, want any, got any) {
	t.Helper()

	opts := []cmp.Option{
		protocmp.Transform(),
		protocmp.IgnoreFields(&pbv1.Seed_Metadata{}, "hash"),
	}

	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Logf("Mismatch (-want +got):\n%s", diff)
		t.FailNow()
	}
}

func TestController_Login(t *testing.T) {
	t.Parallel()

	var (
		nodeID         = "some-node-id"
		challengeID    = "some-challenge-id"
		challengeValue = "some-challenge-value"
		signingKey     = []byte(`some-signing-key`)
		now            = time.Date(2095, time.May, 15, 15, 30, 0, 0, time.UTC)
	)

	t.Run("happy path - post challenge", func(t *testing.T) {
		t.Parallel()

		store := NewMockStorageClient(t)
		store.
			EXPECT().
			ReadChallenge(mock.Anything, challengeID).
			Return(&Challenge{ID: challengeID, Value: challengeValue}, nil)

		ctrl := newControllerWithConfig(
			t,
			ControllerConfig{
				StorageClient: store,
				JWTSigningKey: signingKey,
				NowFunc: func() time.Time {
					return now
				},
			},
			&parsingv2.Config{
				Nodes: []*parsingv2.Node{{ID: nodeID}},
			},
		)

		resp, err := ctrl.Login(context.Background(), connect.NewRequest(&pbv1.LoginRequest{
			NodeId:         nodeID,
			ChallengeId:    hlp.Ptr(challengeID),
			ChallengeValue: hlp.Ptr(challengeValue),
		}))
		require.NoError(t, err)

		require.NotEmpty(t, resp.Msg.Token)

		gotToken, err := token.ParseJWT(*resp.Msg.Token, signingKey)
		require.NoError(t, err)

		wantToken := &token.Token{
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: now.Add(10 * 24 * time.Hour).Unix(),
			},
			NodeID: nodeID,
		}
		require.Equal(t, wantToken, gotToken)
	})
}

func TestController_GetSyncData(t *testing.T) {
	t.Parallel()

	const (
		nodeID = "some-node-id"
	)

	ctx := interceptors.SetTokenOnContext(context.Background(), &token.Token{
		NodeID: nodeID,
	})

	t.Run("config smokes", func(t *testing.T) {
		t.Parallel()

		ctrl := newControllerWithConfig(
			t,
			ControllerConfig{
				VaultClient: NewNoopVault(NoopVaultConfig{}),
			},
			&parsingv2.Config{
				Roles: map[string][]*parsingv2.Seed{
					"foo": {
						{Element: &parsingv2.ConfigFile{
							TemplateContent: "Hello from {{ .Vault.foo }}",
							Destination:     "~/foo/bar",
						}},
					},
				},
				Nodes: []*parsingv2.Node{
					{
						ID:       nodeID,
						Roles:    []string{"foo"},
						UserHome: "/home/fake-user",
					},
				},
			},
		)

		got, err := ctrl.GetSyncData(ctx, connect.NewRequest(&pbv1.GetSyncDataRequest{}))
		require.NoError(t, err)
		pbEqual(
			t,
			&pbv1.GetSyncDataResponse{
				Seeds: []*pbv1.Seed{
					{
						Metadata: &pbv1.Seed_Metadata{
							DisplayName: "~/foo/bar",
						},
						Element: &pbv1.Seed_ConfigFile{
							ConfigFile: &pbv1.ConfigFile{
								Content:     "Hello from static-foo-value",
								Destination: "/home/fake-user/foo/bar",
							},
						},
					},
				},
			},
			got.Msg,
		)
	})
}

func TestValidateGithubRequest(t *testing.T) {
	t.Run("docs example", func(t *testing.T) {
		ctrl := newControllerWithConfig(
			t,
			ControllerConfig{
				GithubWebhookSecret: []byte("It's a Secret to Everybody"),
			},
			nil,
		)

		req, err := http.NewRequest(
			http.MethodPost,
			"http://fake.example.com",
			bytes.NewBuffer([]byte("Hello, World!")),
		)
		require.NoError(t, err)
		req.Header.Add("X-Hub-Signature-256", "sha256=757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17")

		_, err = ctrl.validateGithubRequest(req)
		require.NoError(t, err)
	})
}
