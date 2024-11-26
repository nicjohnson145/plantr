package controller

import (
	"context"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt"
	"github.com/nicjohnson145/hlp"
	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
	"github.com/nicjohnson145/plantr/internal/interceptors"
	"github.com/nicjohnson145/plantr/internal/parsingv2"
	"github.com/nicjohnson145/plantr/internal/token"
	"github.com/nicjohnson145/plantr/internal/vault"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func TestController_CollectSeeds(t *testing.T) {
	t.Parallel()

	t.Run("basic deduplication smokes", func(t *testing.T) {
		t.Parallel()

		ctrl := newControllerWithConfig(
			t,
			ControllerConfig{},
			hlp.Must(parsingv2.ParseFS(os.DirFS("./testdata/collect-seeds/basic"))),
		)

		got, _, err := ctrl.collectSeeds("01JD340PST4R6PY8EDZ5JW127T")
		require.NoError(t, err)

		require.ElementsMatch(
			t,
			[]*parsingv2.Seed{
				{
					Element: &parsingv2.ConfigFile{
						TemplateContent: "hello from template-one\n",
						Destination:     "~/template-one",
					},
				},
				{
					Element: &parsingv2.ConfigFile{
						TemplateContent: "hello from template-two\n",
						Destination:     "~/template-two",
					},
				},
				{
					Element: &parsingv2.ConfigFile{
						TemplateContent: "hello from template-three\n",
						Destination:     "~/template-three",
					},
				},
			},
			got,
		)
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
				VaultClient: vault.NewNoop(vault.NoopConfig{}),
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
						ID: nodeID,
						Roles: []string{"foo"},
						UserHome: "/home/fake-user",
					},
				},
			},
		)

		got, err := ctrl.GetSyncData(ctx, connect.NewRequest(&pbv1.GetSyncDataRequest{}))
		require.NoError(t, err)
		require.Equal(
			t,
			connect.NewResponse(&pbv1.GetSyncDataResponse{
				Seeds: []*pbv1.Seed{
					{Element: &pbv1.Seed_ConfigFile{ConfigFile: &pbv1.ConfigFile{
						Content:     "Hello from static-foo-value",
						Destination: "/home/fake-user/foo/bar",
					}}},
				},
			}),
			got,
		)
	})
}
