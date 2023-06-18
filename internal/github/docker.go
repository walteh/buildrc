package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func (me *GithubClient) BuildXTagString(ctx context.Context, tag string) (string, error) {
	ismain := strings.Contains(me.RepoName(), "main")

	tagnov := strings.TrimPrefix(tag, "v")
	str := ""
	str += "type=ref,event=branch\n"
	str += "type=ref,event=pr\n"
	str += "type=schedule\n"
	str += fmt.Sprintf("type=semver,pattern=v{{version}},value=%s\n", tagnov)
	str += "type=sha\n"
	str += fmt.Sprintf("type=raw,value=latest,enable=%v\n", ismain)
	str += fmt.Sprintf("type=semver,pattern=v{{major}}.{{minor}},value=%s,enable=%v\n", tagnov, ismain)
	str += fmt.Sprintf("type=semver,pattern=v{{major}},value=%s,enable=%v", tagnov, ismain)

	res, err := json.Marshal(str)
	if err != nil {
		return "", err
	}

	return string(res), nil
}

//                                         # labels: "org.opencontainers.image.title=${{ matrix.package }}\norg.opencontainers.image.source=${{ github.event.organization.avatar_url }}",

func (me *GithubClient) BuildxLabelString(ctx context.Context, name string, tag string) (string, error) {

	// Fetch repository
	repo, _, err := me.client.Repositories.Get(ctx, me.OrgName(), me.RepoName())
	if err != nil {
		return "", err
	}

	// Use fetched repository to set labels
	str := ""
	str += fmt.Sprintf("org.opencontainers.image.title=%s/%s/%s\n", me.OrgName(), me.RepoName(), name)
	str += fmt.Sprintf("org.opencontainers.image.source=%s\n", repo.GetHTMLURL())
	str += fmt.Sprintf("org.opencontainers.image.url=%s\n", repo.GetHTMLURL())
	str += fmt.Sprintf("org.opencontainers.image.documentation=%s\n", repo.GetHTMLURL()+"/wiki")
	str += fmt.Sprintf("org.opencontainers.image.version=%s\n", tag)
	str += fmt.Sprintf("org.opencontainers.image.revision=%s\n", tag)
	str += fmt.Sprintf("org.opencontainers.image.vendor=%s\n", repo.GetOwner().GetLogin())
	str += fmt.Sprintf("org.opencontainers.image.licenses=%s\n", repo.GetLicense().GetSPDXID())
	str += fmt.Sprintf("org.opencontainers.image.created=%s\n", time.Now().Format(time.RFC3339))

	res, err := json.Marshal(str)
	if err != nil {
		return "", err
	}

	return string(res), nil
}

// type=ref,event=branch
//   type=semver,pattern=v{{version}},value=
//   type=sha
//   type=raw,value=latest,enable=true
//   type=semver,pattern=v{{major}}.{{minor}},value=,enable=true
//   type=semver,pattern=v{{major}},value=,enable=true

//   type=ref,event=branch
//   type=semver,pattern=v{{version}},value=0.1.0
//   type=sha
//   type=raw,value=latest,enable=false
//   type=semver,pattern=v{{major}}.{{minor}},value=0.1.0,enable=false
//   type=semver,pattern=v{{major}},value=0.1.0,enable=false