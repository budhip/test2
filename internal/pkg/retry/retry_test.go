package retry_test

import (
	"context"
	"testing"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/retry"

	mockKafka "bitbucket.org/Amartha/go-megatron/internal/pkg/kafka/mock"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func init() {
	xlog.InitForTest()
}

type retryTestHelper struct {
	mockCtrl      *gomock.Controller
	mockPublisher *mockKafka.MockPublisher

	retryerSUT retry.Retryer
}

func newRetryTestHelper(t *testing.T, ebCfg *config.ExponentialBackOffConfig) retryTestHelper {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockPublisher := mockKafka.NewMockPublisher(mockCtrl)

	return retryTestHelper{
		mockCtrl:      mockCtrl,
		mockPublisher: mockPublisher,
		retryerSUT:    retry.NewExponentialBackOff(ebCfg),
	}
}

func Test_Retry_ExponentialBackoff(t *testing.T) {
	t.Run("failed - DLQ Operation called and return err", func(t *testing.T) {
		var dlqCallbackCalled int
		testHelper := newRetryTestHelper(t, &config.ExponentialBackOffConfig{MaxRetries: 1})

		testHelper.mockPublisher.EXPECT().
			PublishSyncWithKey(gomock.AssignableToTypeOf(context.Background()), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(assert.AnError).AnyTimes()

		err := testHelper.retryerSUT.Retry(
			context.Background(),
			func() error {
				return testHelper.mockPublisher.PublishSyncWithKey(context.Background(), "", "", "")
			},
			func() error {
				dlqCallbackCalled = dlqCallbackCalled + 1
				return assert.AnError
			},
		)
		assert.NotNil(t, err)
		assert.Equal(t, dlqCallbackCalled, 1)
	})

	t.Run("success - DLQ Operation not called", func(t *testing.T) {
		var dlqCallbackCalled int
		testHelper := newRetryTestHelper(t, &config.ExponentialBackOffConfig{})

		testHelper.mockPublisher.EXPECT().
			PublishSyncWithKey(gomock.AssignableToTypeOf(context.Background()), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(assert.AnError).AnyTimes()

		err := testHelper.retryerSUT.Retry(
			context.Background(),
			func() error {
				return testHelper.mockPublisher.PublishSyncWithKey(context.Background(), "", "", "")
			},
			func() error {
				dlqCallbackCalled = dlqCallbackCalled + 1
				return nil
			},
		)
		assert.NotNil(t, err)
		assert.Equal(t, dlqCallbackCalled, 1)
	})

	t.Run("success - force stop retrying", func(t *testing.T) {
		var dlqCallbackCalled int
		var processCount int
		testHelper := newRetryTestHelper(t, &config.ExponentialBackOffConfig{MaxRetries: 5})

		testHelper.mockPublisher.EXPECT().
			PublishSyncWithKey(gomock.AssignableToTypeOf(context.Background()), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(assert.AnError).AnyTimes()

		err := testHelper.retryerSUT.Retry(
			context.Background(),
			func() error {
				processCount = processCount + 1

				err := testHelper.mockPublisher.PublishSyncWithKey(context.Background(), "", "", "")

				// force stop retrying
				return testHelper.retryerSUT.StopRetryWithErr(err)
			},
			func() error {
				dlqCallbackCalled = dlqCallbackCalled + 1
				return nil
			},
		)

		assert.NotNil(t, err)
		assert.Equal(t, processCount, 1)
		assert.Equal(t, dlqCallbackCalled, 1)
	})
}
