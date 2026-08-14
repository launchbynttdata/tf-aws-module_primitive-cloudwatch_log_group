package common

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/launchbynttdata/lcaf-component-terratest/types"
	"github.com/stretchr/testify/require"
)

func TestComposableComplete(t *testing.T, ctx types.TestContext) {
	groupName := terraform.OutputContext(t, context.Background(), ctx.TerratestTerraformOptions(), "log_group_name")
	groupArn := terraform.OutputContext(t, context.Background(), ctx.TerratestTerraformOptions(), "log_group_arn")
	kmsKeyID := terraform.OutputContext(t, context.Background(), ctx.TerratestTerraformOptions(), "log_group_kms_key_id")
	region := extractRegionFromArn(t, groupArn)
	cloudwatchClient := cloudwatchlogs.NewFromConfig(GetAWSConfig(t, region))

	var describeOutput *cloudwatchlogs.DescribeLogGroupsOutput
	var lastErr error
	const maxAttempts = 10
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, callErr := cloudwatchClient.DescribeLogGroups(context.TODO(), &cloudwatchlogs.DescribeLogGroupsInput{
			LogGroupNamePrefix: &groupName,
		})
		if callErr != nil {
			lastErr = callErr
			t.Logf("DescribeLogGroups attempt %d/%d failed: %v", attempt, maxAttempts, callErr)
			time.Sleep(6 * time.Second)
			continue
		}
		if len(resp.LogGroups) == 0 {
			lastErr = fmt.Errorf("log group %s not found yet", groupName)
			t.Logf("DescribeLogGroups attempt %d/%d: %v", attempt, maxAttempts, lastErr)
			time.Sleep(6 * time.Second)
			continue
		}
		describeOutput = resp
		lastErr = nil
		break
	}
	require.NotNilf(t, describeOutput, "unable to find log group %s after %d attempts: %v", groupName, maxAttempts, lastErr)

	// describeOutput is guaranteed non-nil when retries succeed
	output := describeOutput

	t.Run("TestDoesGroupExist", func(t *testing.T) {
		require.Equal(t, 1, len(output.LogGroups), "Expected 1 log group, got %d", len(output.LogGroups))
	})

	t.Run("TestGroupArn", func(t *testing.T) {
		require.Equal(t, groupArn, *output.LogGroups[0].LogGroupArn, "Expected ARN to be %s, got %s", groupArn, *output.LogGroups[0].LogGroupArn)
	})

	t.Run("TestKmsKeyAssociation", func(t *testing.T) {
		require.NotNil(t, output.LogGroups[0].KmsKeyId, "Expected log group to have a KMS key attached")
		require.Equal(t, kmsKeyID, aws.ToString(output.LogGroups[0].KmsKeyId), "Expected KMS key ID to be %s, got %s", kmsKeyID, aws.ToString(output.LogGroups[0].KmsKeyId))
	})
}

func GetAWSConfig(t *testing.T, region string) (cfg aws.Config) {
	loadOptions := []func(*config.LoadOptions) error{}
	if region != "" {
		loadOptions = append(loadOptions, config.WithRegion(region))
	} else if envRegion := os.Getenv("AWS_REGION"); envRegion != "" {
		loadOptions = append(loadOptions, config.WithRegion(envRegion))
	} else if envRegion := os.Getenv("AWS_DEFAULT_REGION"); envRegion != "" {
		loadOptions = append(loadOptions, config.WithRegion(envRegion))
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), loadOptions...)
	require.NoErrorf(t, err, "unable to load SDK config, %v", err)
	return cfg
}

func extractRegionFromArn(t *testing.T, arn string) string {
	parts := strings.Split(arn, ":")
	require.GreaterOrEqualf(t, len(parts), 4, "unexpected ARN format: %s", arn)
	return parts[3]
}
