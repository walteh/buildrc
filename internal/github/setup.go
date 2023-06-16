package github

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v53/github"
	"github.com/rs/zerolog"
)

func GetCurrentRunTags(ctx context.Context) (string, string, error) {
	tags, err := GetCurrentCommitTags()
	if err != nil {
		return "", "", err
	}

	brc, err := GetNameForThisBuildrcCommitTagPrefix()
	if err != nil {
		return "", "", err
	}

	release := ""
	buildrc := ""

	for _, tag := range tags {
		if strings.HasPrefix(tag, brc) {
			buildrc = tag
		} else if strings.HasPrefix(tag, "v") {
			release = tag
		}
	}

	return release, buildrc, nil

}

func (me *GithubClient) Setup(ctx context.Context, major int) error {

	// create the release for this build
	rel, _, err := GetCurrentRunTags(ctx)
	if err != nil {
		return err
	}

	if rel != "" {
		return nil
	}

	_, err = me.EnsureRelease(ctx, semver.New(uint64(major), 0, 0, "", ""))
	if err != nil {
		return err
	}

	return nil

}

func (me *GithubClient) ShouldBuild(ctx context.Context) (bool, string, error) {

	_, brc, err := GetCurrentRunTags(ctx)
	if err != nil {
		return false, "", err
	}

	if brc != "" {
		return false, "commit is already tagged by buildrc", nil
	}

	name, err := GetCurrentBranch()
	if err != nil {
		return false, "", err
	}

	if name != "main" {
		return true, "not on main branch", nil
	}

	branch, err := me.GetBranch(ctx, "main")
	if err != nil {
		return false, "", err
	}

	num, err := me.GetClosedPullRequestFromCommit(ctx, branch.GetCommit())
	if err != nil {
		return false, "", err
	}

	if num == nil {
		return true, "not a PR merge commit", nil
	} else {
		return false, fmt.Sprintf("PR #%d merged and matches commit tree, its build will be the same", num.GetNumber()), nil
	}
}

// ComputeSHA256Hash computes the SHA256 hash of a file
func ComputeSHA256Hash(filePath string) (string, error) {
	// Open the file
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Create a hasher
	h := sha256.New()

	// Read the file into the hasher
	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}

	// Get the final hash
	hash := h.Sum(nil)

	return hex.EncodeToString(hash), nil
}

func (me *GithubClient) Upload(ctx context.Context, file string) error {

	rel, _, err := GetCurrentRunTags(ctx)
	if err != nil {
		return err
	}

	if rel == "" {
		return fmt.Errorf("no release found")
	}

	rele, _, err := me.client.Repositories.GetReleaseByTag(ctx, me.OrgName(), me.RepoName(), rel)
	if err != nil {
		return err
	}

	filehash, err := ComputeSHA256Hash(file)
	if err != nil {
		return err
	}

	fle, err := os.Open(file)
	if err != nil {
		return err
	}

	defer fle.Close()

	for _, asset := range rele.Assets {
		if asset.GetName() == filepath.Base(fle.Name()) {
			if asset.GetLabel() == filehash {
				return nil
			} else {
				zerolog.Ctx(ctx).Info().Str("local", filehash).Str("release", asset.GetLabel()).Msgf("file hash missmatch, deleting asset %s", asset.GetName())
				_, err = me.client.Repositories.DeleteReleaseAsset(ctx, me.OrgName(), me.RepoName(), asset.GetID())
				if err != nil {
					return err
				}
			}
		}
	}

	_, _, err = me.client.Repositories.UploadReleaseAsset(ctx, me.OrgName(), me.RepoName(), rele.GetID(), &github.UploadOptions{
		Name:  filepath.Base(fle.Name()),
		Label: filehash,
	}, fle)
	if err != nil {
		return err
	}

	return nil
}

func (me *GithubClient) Finalize(ctx context.Context) (*semver.Version, error) {

	rel, brc, err := GetCurrentRunTags(ctx)
	if err != nil {
		return nil, err
	}

	if rel == "" {
		return nil, fmt.Errorf("no release found")
	}

	vers, err := semver.NewVersion(rel)
	if err != nil {
		return nil, err
	}

	if brc != "" {
		return nil, fmt.Errorf("buildrc tag found, not finalizing")
	}

	rele, _, err := me.client.Repositories.GetReleaseByTag(ctx, me.OrgName(), me.RepoName(), rel)
	if err != nil {
		return nil, err
	}

	// update release to not be a draft
	_, _, err = me.client.Repositories.EditRelease(ctx, me.OrgName(), me.RepoName(), rele.GetID(), &github.RepositoryRelease{
		Draft: github.Bool(false),
	})

	if err != nil {
		return nil, err
	}

	brct, err := GetNameForThisBuildrcCommitTag()
	if err != nil {
		return nil, err
	}

	err = me.TagCommit(ctx, brct)
	if err != nil {
		return nil, err
	}

	return vers, nil

}