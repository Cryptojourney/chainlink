package v2

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink/v2/core/services/workflows/types"
)

func Test_classifyActivationError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want ActivationRetryPolicy
	}{
		{
			name: "artifact fetch is retryable",
			err:  &types.ArtifactFetchError{ArtifactType: "binary", URL: "http://example", Err: errors.New("503")},
			want: ActivationRetryable,
		},
		{
			name: "wrapped artifact fetch is retryable",
			err:  fmt.Errorf("create spec: %w", &types.ArtifactFetchError{ArtifactType: "config", URL: "http://example", Err: errors.New("timeout")}),
			want: ActivationRetryable,
		},
		{
			name: "global limit is non-retryable",
			err:  types.ErrGlobalWorkflowCountLimitReached,
			want: ActivationNonRetryable,
		},
		{
			name: "per owner limit is non-retryable",
			err:  types.ErrPerOwnerWorkflowCountLimitReached,
			want: ActivationNonRetryable,
		},
		{
			name: "explicit non-retryable wrapper",
			err:  nonRetryable(errors.New("workflowID mismatch")),
			want: ActivationNonRetryable,
		},
		{
			name: "workflow id mismatch message is non-retryable",
			err:  fmt.Errorf("engine initialization failed: %w", errors.New("workflowID mismatch: abc != def")),
			want: ActivationNonRetryable,
		},
		{
			name: "invalid cron schedule is non-retryable",
			err:  fmt.Errorf("engine initialization failed: %w", errors.New("invalid cron schedule 'bad'")),
			want: ActivationNonRetryable,
		},
		{
			name: "interval exceeded is non-retryable",
			err:  fmt.Errorf("engine initialization failed: %w", errors.New("cron trigger interval exceeded")),
			want: ActivationNonRetryable,
		},
		{
			// Verbatim from the cron capability, wrapped the way
			// v2.Engine wraps a trigger registration failure. None of the
			// pre-existing cron substrings match gocron's wording, so this
			// activation was retried up to defaultMaxActivationRetries even
			// though the schedule can never parse.
			name: "gocron crontab parse failure is non-retryable",
			err: fmt.Errorf("failed to register trigger %s: %w", "trigger_0",
				errors.New("[3]InvalidArgument: failed to initialize job: gocron: CronJob: crontab parse failure\nprovided bad location moon: unknown time zone moon")),
			want: ActivationNonRetryable,
		},
		{
			name: "unknown error is retryable",
			err:  errors.New("unexpected engine failure"),
			want: ActivationRetryable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, classifyActivationError(tt.err))
		})
	}
}

func Test_activationRetryPolicyForEvent(t *testing.T) {
	t.Parallel()

	t.Run("non-activation events are retryable", func(t *testing.T) {
		t.Parallel()
		policy := activationRetryPolicyForEvent(WorkflowDeleted, &types.ArtifactFetchError{})
		require.Equal(t, ActivationRetryable, policy)
	})

	t.Run("activation artifact fetch is retryable", func(t *testing.T) {
		t.Parallel()
		policy := activationRetryPolicyForEvent(WorkflowActivated, &types.ArtifactFetchError{})
		require.Equal(t, ActivationRetryable, policy)
	})
}
