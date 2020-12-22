package cli

import (
	"context"
	"sort"

	"github.com/dustin/go-humanize"
	"github.com/golang/protobuf/ptypes"
	"github.com/posener/complete"

	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	clientpkg "github.com/hashicorp/waypoint/internal/client"
	"github.com/hashicorp/waypoint/internal/clierrors"
	"github.com/hashicorp/waypoint/internal/pkg/flag"
	pb "github.com/hashicorp/waypoint/internal/server/gen"
	serversort "github.com/hashicorp/waypoint/internal/server/sort"
)

type BuildListCommand struct {
	*baseCommand

	flagWorkspaceAll bool
	flagId           idFormat
}

func (c *BuildListCommand) Run(args []string) int {
	// Initialize. If we fail, we just exit since Init handles the UI.
	if err := c.Init(
		WithArgs(args),
		WithFlags(c.Flags()),
		WithSingleApp(),
	); err != nil {
		return 1
	}

	// Get our API client
	client := c.project.Client()

	err := c.DoApp(c.Ctx, func(ctx context.Context, app *clientpkg.App) error {
		var wsRef *pb.Ref_Workspace
		if !c.flagWorkspaceAll {
			wsRef = c.project.WorkspaceRef()
		}

		// List builds
		resp, err := client.ListBuilds(c.Ctx, &pb.ListBuildsRequest{
			Application: app.Ref(),
			Workspace:   wsRef,
		})
		if err != nil {
			c.project.UI.Output(clierrors.Humanize(err), terminal.WithErrorStyle())
			return ErrSentinel
		}
		sort.Sort(serversort.BuildStartDesc(resp.Builds))

		const bullet = "●"

		table := terminal.NewTable("", "ID", "Workspace", "Builder", "Started", "Completed")
		for _, b := range resp.Builds {
			// Determine our bullet
			status := ""
			statusColor := ""
			switch b.Status.State {
			case pb.Status_RUNNING:
				status = bullet
				statusColor = terminal.Yellow

			case pb.Status_SUCCESS:
				status = "✔"
				statusColor = terminal.Green

			case pb.Status_ERROR:
				status = "✖"
				statusColor = terminal.Red
			}

			// Parse our times
			var startTime, completeTime string
			if t, err := ptypes.Timestamp(b.Status.StartTime); err == nil {
				startTime = humanize.Time(t)
			}
			if t, err := ptypes.Timestamp(b.Status.CompleteTime); err == nil {
				completeTime = humanize.Time(t)
			}

			table.Rich([]string{
				status,
				c.flagId.FormatId(b.Sequence, b.Id),
				b.Workspace.Workspace,
				b.Component.Name,
				startTime,
				completeTime,
			}, []string{
				statusColor,
			})
		}

		c.ui.Table(table)

		return nil
	})
	if err != nil {
		return 1
	}

	return 0
}

func (c *BuildListCommand) Flags() *flag.Sets {
	return c.flagSet(0, func(set *flag.Sets) {
		f := set.NewSet("Command Options")
		f.BoolVar(&flag.BoolVar{
			Name:   "workspace-all",
			Target: &c.flagWorkspaceAll,
			Usage:  "List builds in all workspaces for this project and application.",
		})

		initIdFormat(f, &c.flagId)
	})
}

func (c *BuildListCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *BuildListCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *BuildListCommand) Synopsis() string {
	return "List builds."
}

func (c *BuildListCommand) Help() string {
	return formatHelp(`
Usage: waypoint artifact list-builds [options]
  List builds.

` + c.Flags().Help())
}
