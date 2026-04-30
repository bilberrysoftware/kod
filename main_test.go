package main

import (
	"context"
	"fmt"
	"kod/internal"
	"kod/types"
	"testing"

	"github.com/cucumber/godog"
)

/* Tagging struct for GoDogs */
type ContainerListResultKey struct{}

type ContainerListResultType struct {
	ProcessingResult types.HelmChartProcessingResult
	Containers       types.ContainerImageList
}

/* State struct fetched via ContainerListResultKey for HelmChartProcessingResult and ContainerImageList tests */
var ContainerListResult = ContainerListResultType{}

func thereIsANewHelmChartProcessingResult(ctx context.Context) (context.Context, error) {
	ContainerListResult.ProcessingResult.Children = []types.HelmChartProcessingResult{}
	ContainerListResult.ProcessingResult.Containers = []types.ContainerImage{}
	return context.WithValue(ctx, ContainerListResultKey{}, &ContainerListResult), nil
}

func aHelmChartProcessingResultContainerWith(ctx context.Context, registry, repository, tag, digest string) (context.Context, error) {
	ContainerListResult.ProcessingResult.Containers = append(ContainerListResult.ProcessingResult.Containers, types.ContainerImage{
		Registry:   registry,
		Repository: repository,
		Tag:        tag,
		Digest:     digest,
	})
	return context.WithValue(ctx, ContainerListResultKey{}, &ContainerListResult), nil
}

func iProcessTheHelmChartProcessingResultToGetAContainerList(ctx context.Context) (context.Context, error) {
	helmResult, ok := ctx.Value(ContainerListResultKey{}).(*ContainerListResultType)
	if !ok {
		return ctx, fmt.Errorf("could not get a container list result key from the context")
	}
	internal.PopulateContainerList(&ContainerListResult.Containers, &helmResult.ProcessingResult)
	return context.WithValue(ctx, ContainerListResultKey{}, &ContainerListResult), nil
}

func thereShouldBeXContainersInTheContainerList(ctx context.Context, expected int) error {
	helmResult, ok := ctx.Value(ContainerListResultKey{}).(*ContainerListResultType)
	if !ok {
		return fmt.Errorf("could not get a container list result key from the context")
	}
	available := len(helmResult.Containers.Containers)
	if available != expected {
		return fmt.Errorf("expected %d containers in the list, but there is %d", expected, available)
	}
	return nil
}

func oneOfTheseContainersWith(ctx context.Context, registry, repository, tag, digest string) error {
	helmResult, ok := ctx.Value(ContainerListResultKey{}).(*ContainerListResultType)
	if !ok {
		return fmt.Errorf("could not get a container list result key from the context")
	}
	found := false
	for _, ctr := range helmResult.Containers.Containers {
		found = found || ctr.Registry == registry && ctr.Repository == repository && ctr.Tag == tag && ctr.Digest == digest
	}
	if !found {
		return fmt.Errorf("could not find a container with registry: '%s', repository: '%s', tag: '%s', digest:'%s' in the list", registry, repository, tag, digest)
	}
	return nil
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t, // Testing instance that will run subtests.
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^there is a new Helm Chart Processing Result$`, thereIsANewHelmChartProcessingResult)
	ctx.Step(`^a Helm Chart Processing Result container with Registry \'(.*)\' and Repository \'(.*)\' and Tag \'(.*)\' and Digest \'(.*)\'$`, aHelmChartProcessingResultContainerWith)
	ctx.Step(`^I process the Helm Chart Processing Result to get a Container List$`, iProcessTheHelmChartProcessingResultToGetAContainerList)
	ctx.Step(`^one of these containers with Registry \'(.*)\' and Repository \'(.*)' and Tag \'(.*)\' and Digest \'(.*)\'$`, oneOfTheseContainersWith)
	ctx.Step(`^there should be (\d+) containers in the container list$`, thereShouldBeXContainersInTheContainerList)
}
