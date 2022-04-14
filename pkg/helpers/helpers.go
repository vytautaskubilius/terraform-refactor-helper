package helpers

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"log"
	"strings"
)

func SetupTerraform(workingDir string, backendConfig string, workspace string) *tfexec.Terraform {
	execPath := "/usr/local/bin/terraform"
	log.Println("Initializing Terraform with the following configuration:")
	log.Printf("Working directory: %s", workingDir)
	log.Printf("Backend configuration: %s", backendConfig)
	log.Printf("Workspace: %s", workspace)

	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		log.Fatalf("Error running NewTerraform: %s", err)
	}

	err = tf.Init(context.Background(), tfexec.BackendConfig(backendConfig))
	if err != nil {
		log.Fatalf("Error inititalizing Terraform: %s", err)
	}

	err = tf.WorkspaceSelect(context.Background(), workspace)
	if err != nil {
		log.Fatalf("Error selecting workspace: %s", err)
	}

	return tf
}

func GetTerraformState(tf *tfexec.Terraform) *tfjson.State {
	workspace, err := tf.WorkspaceShow(context.Background())
	if err != nil {
		log.Fatalf("Error showing Terraform workspace: %s", err)
	}
	log.Printf("Fetching the Terraform state from workspace %s in %s...", workspace, tf.WorkingDir())
	state, err := tf.Show(context.Background())
	if err != nil {
		log.Fatalf("Error showing Terraform state: %s", err)
	}
	return state
}

func GetManagedResources(module tfjson.StateModule) map[string]string {
	resourceMap := make(map[string]string)
	for _, resource := range module.Resources {
		if resource.Mode == "managed" && !resource.Tainted {
			switch resource.Type {
			case "aws_ecs_cluster":
				value := fmt.Sprintf("%v", resource.AttributeValues["name"])
				resourceMap[resource.Address] = value
			case "aws_ecs_service":
				serviceName := fmt.Sprintf("%v", resource.AttributeValues["name"])
				clusterName := strings.Split(fmt.Sprintf("%v", resource.AttributeValues["cluster"]), "/")[1]
				value := fmt.Sprintf("%s/%s", clusterName, serviceName)
				resourceMap[resource.Address] = value
			case "aws_ecs_task_definition":
				arn := fmt.Sprintf("%v", resource.AttributeValues["arn"])
				resourceMap[resource.Address] = arn
			case "aws_route":
				routeTableId := fmt.Sprintf("%v", resource.AttributeValues["route_table_id"])
				destinationCidrBlock := fmt.Sprintf("%v", resource.AttributeValues["destination_cidr_block"])
				value := fmt.Sprintf("%s_%s", routeTableId, destinationCidrBlock)
				resourceMap[resource.Address] = value
			case "aws_cloudwatch_event_target":
				eventBusName := fmt.Sprintf("%v", resource.AttributeValues["event_bus_name"])
				ruleName := fmt.Sprintf("%v", resource.AttributeValues["rule"])
				targetId := fmt.Sprintf("%v", resource.AttributeValues["target_id"])
				value := fmt.Sprintf("%s/%s/%s", eventBusName, ruleName, targetId)
				resourceMap[resource.Address] = value
			default:
				value := fmt.Sprintf("%v", resource.AttributeValues["id"])
				resourceMap[resource.Address] = value
			}
		}
	}
	if module.ChildModules != nil {
		for _, childModule := range module.ChildModules {
			childModuleResourceMap := GetManagedResources(*childModule)
			for childModuleResourceAddress, childModuleResourceValue := range childModuleResourceMap {
				resourceMap[childModuleResourceAddress] = childModuleResourceValue
			}
		}
	}
	return resourceMap
}

func GetAllResources(module tfjson.StateModule) map[string]string {
	resourceMap := make(map[string]string)
	for _, resource := range module.Resources {
		value := fmt.Sprintf("%v", resource.AttributeValues["id"])
		resourceMap[resource.Address] = value
	}
	if module.ChildModules != nil {
		for _, childModule := range module.ChildModules {
			childModuleResourceMap := GetAllResources(*childModule)
			for childModuleResourceAddress, childModuleResourceValue := range childModuleResourceMap {
				resourceMap[childModuleResourceAddress] = childModuleResourceValue
			}
		}
	}
	return resourceMap
}

func GetFilteredResources(module tfjson.StateModule, filters []string, mode string) []map[string]string {
	log.Println("Fetching resources from the state...")
	var resources map[string]string
	switch mode {
	case "managed":
		resources = GetManagedResources(module)
	case "all":
		resources = GetAllResources(module)
	}
	filteredResourcesSlice := make([]map[string]string, 0)
	for _, filter := range filters {
		filteredResourcesSlice = append(filteredResourcesSlice, FilterResources(resources, filter))
	}
	return filteredResourcesSlice
}

func FilterResources(resourceMap map[string]string, filter string) map[string]string {
	filteredResourceMap := make(map[string]string)
	if !strings.HasSuffix(filter, ".") {
		filter = filter + "."
	}
	for address, id := range resourceMap {
		if strings.HasPrefix(address, filter) {
			filteredResourceMap[address] = id
		}
	}
	return filteredResourceMap
}

func ImportResources(tf *tfexec.Terraform, filteredResourcesSlice []map[string]string) {
	log.Println("Starting the import.")
	for _, filteredResources := range filteredResourcesSlice {
		for address, id := range filteredResources {
			log.Printf("Importing %s into %s\n", id, address)
			err := tf.Import(context.Background(), address, id)
			if err != nil {
				log.Printf("Error importing Terraform resource: %s", err)
			}
		}
	}
}

func RemoveResources(tf *tfexec.Terraform, filteredResourcesSlice []map[string]string) {
	log.Println("Starting resource removal.")
	for _, filteredResources := range filteredResourcesSlice {
		for address := range filteredResources {
			log.Printf("Removing %s\n", address)
			err := tf.StateRm(context.Background(), address)
			if err != nil {
				log.Printf("Error removing Terraform resource: %s", err)
			}
		}
	}
}
