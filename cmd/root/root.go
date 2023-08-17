package root

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/walteh/buildrc/cmd/root/next"
	"github.com/walteh/buildrc/internal/buildrc"
	"github.com/walteh/buildrc/version"
	"github.com/walteh/snake"
)

type Root struct {
	Quiet   bool
	Debug   bool
	Version bool
	File    string
}

var _ snake.Snakeable = (*Root)(nil)

func (me *Root) BuildCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buildrc",
		Short: "buildrc brings tabs to terraform",
	}

	cmd.PersistentFlags().BoolVarP(&me.Quiet, "quiet", "q", false, "Do not print any output")
	cmd.PersistentFlags().BoolVarP(&me.Debug, "debug", "d", false, "Print debug output")
	cmd.PersistentFlags().BoolVarP(&me.Version, "version", "v", false, "Print version and exit")
	cmd.PersistentFlags().StringVarP(&me.File, "file", "f", ".buildrc", "The buildrc file to use")

	snake.MustNewCommand(ctx, cmd, &next.Handler{})

	return cmd
}

func (me *Root) ParseArguments(ctx context.Context, cmd *cobra.Command, args []string) error {

	var level zerolog.Level
	if me.Debug {
		level = zerolog.TraceLevel
	} else if me.Quiet {
		level = zerolog.NoLevel
	} else {
		level = zerolog.InfoLevel
	}

	ctx = zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger().Level(level).WithContext(ctx)

	if me.Version {
		cmd.Printf("%s %s %s\n", version.Package, version.Version, version.Revision)
		os.Exit(0)
	}

	zerolog.Ctx(ctx).Debug().Msg("parsing buildrc file")

	abc, err := buildrc.Parse(ctx, me.File)
	if err != nil {
		return err
	}

	ctx = snake.Bind(ctx, &buildrc.Buildrc{}, abc)

	cmd.SetContext(ctx)

	return nil
}