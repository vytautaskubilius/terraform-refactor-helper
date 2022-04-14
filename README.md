# Terraform refactor helper

This is a tool that simplifies various repetitive tasks that need to be performed when refactoring Terraform modules.

# Installing the tool

You will need [Go](https://go.dev/doc/install) to be able to build and install the tool.

1. Clone the repo
2. Install the executable with `go install`.

# Usage

Refer to the output of `terraform-refactor-helper --help` for usage instructions.

## Features

### Automate resource import

The tool allows the user to provide a list of resource address prefixes that will be used as filters to extract resources
from a source state and import them to a destination state.

### Automate resouce cleanup

The tool allows the user to provider a list of resource prefixes that will be used as filters to remove resources from
a Terraform state. This is a highly destructive operation, so the Terraform state should be [backed up locally](https://www.terraform.io/cli/commands/state/pull)
before the operation.

# Links

- [Download and install Go](https://go.dev/doc/install).
- [terraform-exec](https://pkg.go.dev/github.com/hashicorp/terraform-exec).
- [terraform-json](https://pkg.go.dev/github.com/hashicorp/terraform-json).

# TODO
- Add tests.
- Add a CI/CD workflow.
