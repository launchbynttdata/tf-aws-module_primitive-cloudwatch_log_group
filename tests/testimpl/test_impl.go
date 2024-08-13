package common

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/launchbynttdata/lcaf-component-terratest/types"
	"github.com/stretchr/testify/require"
)

func TestLogGroup(t *testing.T, ctx types.TestContext) {
	cloudwatchClient := cloudwatchlogs.NewFromConfig(GetAWSConfig(t))
	groupName := terraform.Output(t, ctx.TerratestTerraformOptions(), "log_group_name")
	groupArn := terraform.Output(t, ctx.TerratestTerraformOptions(), "log_group_arn")

	output, err := cloudwatchClient.DescribeLogGroups(context.TODO(), &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: &groupName,
	})
	if err != nil {
		t.Errorf("Unable to get log groups, %v", err)
	}

	t.Run("TestDoesGroupExist", func(t *testing.T) {
		require.Equal(t, 1, len(output.LogGroups), "Expected 1 log group, got %d", len(output.LogGroups))
	})

	t.Run("TestGroupArn", func(t *testing.T) {
		require.Equal(t, groupArn, *output.LogGroups[0].LogGroupArn, "Expected ARN to be %s, got %s", groupArn, *output.LogGroups[0].LogGroupArn)
	})
}

func GetAWSConfig(t *testing.T) (cfg aws.Config) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	require.NoErrorf(t, err, "unable to load SDK config, %v", err)
	return cfg
}
