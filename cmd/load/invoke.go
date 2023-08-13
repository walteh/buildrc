package load

import (
	"context"

	"github.com/nuggxyz/buildrc/internal/buildrc"
	"github.com/nuggxyz/buildrc/internal/common"
	"github.com/nuggxyz/buildrc/internal/git"
	"github.com/nuggxyz/buildrc/internal/pipeline"
)

const (
	CommandID = "load"
)

type Handler struct {
	File string `flag:"file" type:"file:" default:".buildrc"`
}

func NewHandler(file string) *Handler {
	return &Handler{File: file}
}

func (me *Handler) Run(ctx context.Context, prov common.Provider) (err error) {

	out, err := buildrc.Parse(ctx, me.File)
	if err != nil {
		return err
	}

	err = pipeline.SetupEnvDirs(ctx, prov.Pipeline(), prov.FileSystem())
	if err != nil {
		return err
	}

	targetSemver, err := git.CalculateNextPreReleaseTag(ctx, prov.Buildrc(), prov.Git(), prov.PR())
	if err != nil {
		return err
	}

	arr, err := out.PackagesArrayJSON()
	if err != nil {
		return err
	}

	sha256, err := pipeline.BuildrcArtifactsToReleaseAsSha256Dir.Path(ctx, prov.Pipeline())
	if err != nil {
		return err
	}

	targz, err := pipeline.BuildrcArtifactsToReleaseAsTarGZDir.Path(ctx, prov.Pipeline())
	if err != nil {
		return err
	}

	mapper, err := out.PackagesMapJSON()
	if err != nil {
		return err
	}

	outer, err := pipeline.ResolveRunsOnMapJSON(out, prov.Pipeline())
	if err != nil {
		return err
	}

	export := map[string]string{
		"BUILDRC_PACKAGES_NAME_ARRAY_JSON":  out.PackagesNamesArrayJSON(),
		"BUILDRC_PACKAGES_RUNS_ON_MAP_JSON": outer,
		"BUILDRC_TAG":                       targetSemver.String(),
		"BUILDRC_PACKAGES_ARRAY_JSON":       arr,
		"BUILDRC_SHA256":                    sha256,
		"BUILDRC_TARGZ":                     targz,
		"BUILDRC_PACKAGES_MAP_JSON":         mapper,
	}

	err = pipeline.AddContentToEnv(ctx, prov.Pipeline(), prov.FileSystem(), CommandID, export)

	if err != nil {
		return err
	}

	return nil
}
