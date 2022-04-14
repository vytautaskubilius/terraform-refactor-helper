package main

import (
	"github.com/urfave/cli/v2"
	"github.com/vytautaskubilius/terraform-refactor-helper/pkg/helpers"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name:  "terraform-refactor-helper",
		Usage: "Tool for performing common Terraform refactoring tasks",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "module-prefix",
				Usage:    "Module prefix to import",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "backend-config",
				Usage:    "Backend configuration for the remote Terraform state file",
				Required: true,
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "migrate-state-resources",
				Usage: "Migrate resources with the selected prefix to a different state",
				Action: func(context *cli.Context) error {
					terraformSource := helpers.SetupTerraform(context.String("source-working-dir"), context.String("backend-config"), context.String("source-workspace"))
					terraformDestination := helpers.SetupTerraform(context.String("destination-working-dir"), context.String("backend-config"), context.String("destination-workspace"))
					state := helpers.GetTerraformState(terraformSource)
					filteredResourcesSlice := helpers.GetFilteredResources(*state.Values.RootModule, context.StringSlice("module-prefix"), "managed")
					helpers.ImportResources(terraformDestination, filteredResourcesSlice)
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "source-working-dir",
						Usage:    "Directory of the source Terraform module",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "source-workspace",
						Usage:    "Terraform workspace for the source state",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "destination-working-dir",
						Usage:    "Directory of the destination Terraform module",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "destination-workspace",
						Usage:    "Terraform workspace for the destination state (must exist prior to running this tool)",
						Required: true,
					},
				},
			},
			{
				Name:  "cleanup-state-resources",
				Usage: "Remove resources with the selected prefix from the state",
				Action: func(context *cli.Context) error {
					terraform := helpers.SetupTerraform(context.String("working-dir"), context.String("backend-config"), context.String("workspace"))
					state := helpers.GetTerraformState(terraform)
					filteredResourcesSlice := helpers.GetFilteredResources(*state.Values.RootModule, context.StringSlice("module-prefix"), "all")
					helpers.RemoveResources(terraform, filteredResourcesSlice)
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "working-dir",
						Usage:    "Directory of the Terraform module",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "workspace",
						Usage:    "Terraform workspace",
						Required: true,
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("Error running the application: %s", err)
	}
}
