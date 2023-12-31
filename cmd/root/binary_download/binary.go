package binary_download

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/buildrc/pkg/buildrc"
	"github.com/walteh/buildrc/pkg/install"
	"github.com/walteh/snake"
)

var _ snake.Snakeable = (*Handler)(nil)

type Handler struct {
	Organization string
	Repository   string
	Version      string
	Token        string
	Provider     string
	OutFile      string
	Platform     string
}

func (me *Handler) BuildCommand(_ context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Short: "install buildrc",
	}

	cmd.Args = cobra.ExactArgs(0)

	cmd.PersistentFlags().StringVar(&me.Provider, "provider", "github", "Provider to install from")
	cmd.PersistentFlags().StringVar(&me.Repository, "repository", "", "Repository to install from")
	cmd.PersistentFlags().StringVar(&me.Organization, "organization", "", "Organization to install from")
	cmd.PersistentFlags().StringVar(&me.Version, "version", "latest", "Version to install")
	cmd.PersistentFlags().StringVar(&me.OutFile, "outfile", "", "Output file")
	cmd.PersistentFlags().StringVar(&me.Platform, "platform", "runtime.GOOS/runtime.GOARCH", "Platform to install for")

	cmd.PersistentFlags().StringVar(&me.Token, "token", "", "Oauth2 token to use")

	return cmd
}

func (me *Handler) ParseArguments(_ context.Context, _ *cobra.Command, _ []string) error {

	if me.Repository == "" || me.Organization == "" {
		return errors.Errorf("Repository and organization must be specified")
	}

	return nil

}

func (me *Handler) Run(ctx context.Context) error {
	var fle afero.File
	var err error

	switch me.Provider {
	case "github":
		{

			var plat *buildrc.Platform
			if me.Platform == "runtime.GOOS/runtime.GOARCH" {
				plat = buildrc.GetGoPlatform(ctx)
			} else {

				plat, err = buildrc.NewPlatformFromFullString(me.Platform)
				if err != nil {
					return err
				}
			}

			fle, err = install.DownloadGithubReleaseWithOptions(ctx, afero.NewOsFs(), &install.DownloadGithubReleaseOptions{
				Org:      me.Organization,
				Name:     me.Repository,
				Version:  me.Version,
				Token:    me.Token,
				Platform: plat,
			})
			if err != nil {
				return err
			}
		}
	default:
		{
			return errors.Errorf("Unknown provider: %s", me.Provider)
		}
	}

	defer fle.Close()

	fls := afero.NewOsFs()

	err = afero.WriteReader(fls, me.OutFile, fle)
	if err != nil {
		return err
	}

	return fls.Chmod(me.OutFile, 0755)

}
