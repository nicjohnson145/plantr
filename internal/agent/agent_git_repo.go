package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/nicjohnson145/hlp"
	controllerv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
	"github.com/nicjohnson145/plantr/internal/util"
)

func (a *Agent) executeSeed_gitRepo(ctx context.Context, pbseed *controllerv1.Seed) (*InventoryRow, error) {
	pbRepo := pbseed.Element.(*controllerv1.Seed_GitRepo).GitRepo

	// Does the location already exist?
	exists, err := util.PathExists(pbRepo.Location)
	if err != nil {
		return nil, fmt.Errorf("error determining repo existence: %w", err)
	}

	var repoFunc func(*controllerv1.GitRepo) (*git.Repository, error)
	if !exists {
		repoFunc = a.checkoutRepo
	} else {
		repoFunc = a.openRepo
	}

	repo, err := repoFunc(pbRepo)
	if err != nil {
		return nil, fmt.Errorf("error checking out repo: %w", err)
	}

	// Fetch latest, so we dont try to lookup a tag that doesnt exist on our local
	if err := a.fetchLatest(repo); err != nil {
		return nil, fmt.Errorf("error fetching latest: %w", err)
	}

	// Translate our desired reference into a commit hash
	wantHash, err := a.translateToHash(pbRepo, repo)
	if err != nil {
		return nil, fmt.Errorf("error translating desired ref: %w", err)
	}

	// Checkout that new desired hash
	if err := a.checkoutCommit(repo, wantHash); err != nil {
		return nil, fmt.Errorf("error checking out commit: %w", err)
	}

	return &InventoryRow{
		Path: hlp.Ptr(pbRepo.Location),
	}, nil
}

func (a *Agent) checkoutRepo(pbRepo *controllerv1.GitRepo) (*git.Repository, error) {
	// ensure any containing directories already exist
	if err := os.MkdirAll(filepath.Dir(pbRepo.Location), 0775); err != nil {
		return nil, fmt.Errorf("error ensuring containing directories: %w", err)
	}

	repo, err := git.PlainClone(pbRepo.Location, false, &git.CloneOptions{
		URL: pbRepo.Url,
	})
	if err != nil {
		return nil, fmt.Errorf("error cloning repo: %w", err)
	}

	return repo, nil
}

func (a *Agent) openRepo(pbRepo *controllerv1.GitRepo) (*git.Repository, error) {
	repo, err := git.PlainOpen(pbRepo.Location)
	if err != nil {
		return nil, fmt.Errorf("error opening existing repo: %w", err)
	}
	return repo, nil
}

func (a *Agent) fetchLatest(repo *git.Repository) error {
	if err := repo.Fetch(&git.FetchOptions{}); err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("error fetching: %w", err)
	}
	return nil
}

func (a *Agent) translateToHash(pbRepo *controllerv1.GitRepo, repo *git.Repository) (string, error) {
	switch concrete := pbRepo.Ref.(type) {
	case *controllerv1.GitRepo_Commit:
		return concrete.Commit, nil
	case *controllerv1.GitRepo_Tag:
		h, err := repo.ResolveRevision(plumbing.Revision(concrete.Tag))
		if err != nil {
			return "", fmt.Errorf("error resolving tag: %w", err)
		}
		return h.String(), nil
	default:
		return "", fmt.Errorf("unhandled reference type %T", concrete)
	}
}

func (a *Agent) checkoutCommit(repo *git.Repository, desired string) error {
	tree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("error getting work tree: %w", err)
	}

	if err := tree.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(desired)}); err != nil {
		return fmt.Errorf("error executing checkout: %w", err)
	}

	return nil
}
