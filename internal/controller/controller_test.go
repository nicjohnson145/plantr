package controller

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt"
	"github.com/nicjohnson145/hlp"
	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/v1"
	"github.com/nicjohnson145/plantr/internal/git"
	"github.com/nicjohnson145/plantr/internal/storage"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newControllerWithConfig(t *testing.T, conf ControllerConfig, repoConfig *pbv1.Config) *Controller {
	t.Helper()

	if conf.GitClient == nil {
		conf.GitClient = git.NewMockClient(t)
	}
	if conf.StorageClient == nil {
		conf.StorageClient = storage.NewMockClient(t)
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
	var (
		nodeID         = "some-node-id"
		challengeID    = "some-challenge-id"
		challengeValue = "some-challenge-value"
		signingKey     = []byte(`some-signing-key`)
		now            = time.Date(2095, time.May, 15, 15, 30, 0, 0, time.UTC)
	)

	t.Run("happy path - post challenge", func(t *testing.T) {
		store := storage.NewMockClient(t)
		store.
			EXPECT().
			ReadChallenge(mock.Anything, challengeID).
			Return(&storage.Challenge{ID: challengeID, Value: challengeValue}, nil)

		ctrl := newControllerWithConfig(
			t,
			ControllerConfig{
				StorageClient: store,
				JWTSigningKey: signingKey,
				NowFunc: func() time.Time {
					return now
				},
			},
			&pbv1.Config{
				Nodes: []*pbv1.Node{{Id: nodeID}},
			},
		)

		resp, err := ctrl.Login(context.Background(), connect.NewRequest(&pbv1.LoginRequest{
			NodeId:         nodeID,
			ChallengeId:    hlp.Ptr(challengeID),
			ChallengeValue: hlp.Ptr(challengeValue),
		}))
		require.NoError(t, err)

		require.NotEmpty(t, resp.Msg.Token)

		token, err := ParseJWT(*resp.Msg.Token, signingKey)
		require.NoError(t, err)

		wantToken := &Token{
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: now.Add(10 * 24 * time.Hour).Unix(),
			},
			NodeID: nodeID,
		}
		require.Equal(t, wantToken, token)
	})
}
