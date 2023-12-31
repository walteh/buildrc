package root

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/buildrc/cmd/root/binary_download"
	"github.com/walteh/buildrc/cmd/root/binary_install"
	"github.com/walteh/buildrc/cmd/root/diff"

	"github.com/walteh/buildrc/cmd/root/full"
	"github.com/walteh/buildrc/cmd/root/next_version"
	"github.com/walteh/buildrc/cmd/root/revision"
	"github.com/walteh/buildrc/pkg/git"

	myversion "github.com/walteh/buildrc/version"
	"github.com/walteh/snake"
)

type Root struct {
	Quiet   bool
	Debug   bool
	Version bool
	File    string
	GitDir  string
}

var _ snake.Snakeable = (*Root)(nil)

func (me *Root) BuildCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buildrc",
		Short: "buildrc is a tool to help with building releases",
	}

	cmd.PersistentFlags().BoolVarP(&me.Quiet, "quiet", "q", false, "Do not print any output")
	cmd.PersistentFlags().BoolVarP(&me.Debug, "debug", "d", false, "Print debug output")
	cmd.PersistentFlags().BoolVarP(&me.Version, "version", "v", false, "Print version and exit")
	cmd.PersistentFlags().StringVar(&me.GitDir, "git-dir", ".", "The git directory to use")

	snake.MustNewCommand(ctx, cmd, "next-version", &next_version.Handler{})
	snake.MustNewCommand(ctx, cmd, "revision", &revision.Handler{})
	snake.MustNewCommand(ctx, cmd, "full", &full.Handler{})
	snake.MustNewCommand(ctx, cmd, "binary-install", &binary_install.Handler{})
	snake.MustNewCommand(ctx, cmd, "diff", &diff.Handler{})
	snake.MustNewCommand(ctx, cmd, "binary-download", &binary_download.Handler{})

	cmd.SetOutput(os.Stdout)

	cmd.SilenceUsage = true

	cmd.SetHelpTemplate(cmd.UsageTemplate())
	cmd.SetUsageTemplate("Usage:  {{.UseLine}}\n")

	return cmd
}

func (me *Root) ParseArguments(ctx context.Context, cmd *cobra.Command, _ []string) error {

	var level zerolog.Level
	if me.Debug {
		level = zerolog.TraceLevel
	} else if me.Quiet {
		level = zerolog.NoLevel
	} else {
		level = zerolog.InfoLevel
	}

	ctx = zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Caller().Logger().Level(level).WithContext(ctx)

	if me.Version {
		cmd.Printf("%s %s %s\n", myversion.Package, myversion.Version, myversion.Revision)
		os.Exit(0)
	}

	root := afero.NewOsFs()

	gpv, err := git.NewGitGoGitProvider(afero.NewOsFs(), me.GitDir)
	if err != nil {
		return err
	}

	ctx = snake.Bind(ctx, (*git.GitProvider)(nil), gpv)
	ctx = snake.Bind(ctx, (*afero.Fs)(nil), root)

	cmd.SetContext(ctx)

	return nil
}
