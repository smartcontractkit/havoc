package k8schaos

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type BaseChaos struct {
	Description string
	DelayCreate time.Duration // Delay before creating the chaos object
	Status      ChaosStatus
	startTime   time.Time
	endTime     time.Time
	logger      *zerolog.Logger

	BaseChaosLifecycle
}

type BaseChaosOpts struct {
	Description string
	DelayCreate time.Duration
	Logger      *zerolog.Logger
}

func NewBaseChaos(opts BaseChaosOpts) (BaseChaos, error) {
	if opts.Logger == nil {
		return BaseChaos{}, errors.New("logger is required")
	}

	return BaseChaos{
		Description: opts.Description,
		DelayCreate: opts.DelayCreate,
		logger:      opts.Logger,
	}, nil
}

type BaseChaosLifecycle interface {
	CreateNow(ctx context.Context) error // Function to execute the actual creation
	OnChaosCreated()                     // Function to execute after creation is successful
	OnChaosCreationFailed(err error)     // Function to execute after creation failed
	OnChaosStarted()
}

// CreateAsync initiates the asynchronous creation of a chaos object, respecting context cancellation and deletion requests.
// It uses a timer based on `DelayCreate` and calls create function upon expiration unless preempted by deletion.
func (c *BaseChaos) CreateAsync(ctx context.Context) {
	done := make(chan struct{})

	// Create the timer with the delay to create the chaos object
	timer := time.NewTimer(c.DelayCreate)

	go func() {
		select {
		case <-ctx.Done():
			// If the context is canceled, stop the timer and exit
			if !timer.Stop() {
				<-timer.C // If the timer already expired, drain the channel
			}
			close(done) // Signal that the operation was canceled
		case <-timer.C:
			// Timer expired, check if deletion was not requested
			if c.Status != StatusDeleted {
				if err := c.CreateNow(ctx); err != nil {
					c.OnChaosCreationFailed(err)
				} else {
					c.OnChaosCreated()
				}
			}
			close(done) // Signal that the creation process is either done or skipped
		}
	}()
}

// CreateSync initiates the synchronous creation of a chaos object, respecting context cancellation and deletion requests.
// It blocks until the DelayCreate has elapsed and then proceeds to create the chaos object unless preempted by deletion.
func (c *BaseChaos) CreateSync(ctx context.Context) error {
	select {
	case <-time.After(c.DelayCreate):
		// Delay has elapsed, proceed with creation
		if c.Status != StatusDeleted {
			if err := c.CreateNow(ctx); err != nil {
				c.OnChaosCreationFailed(err)
				return err
			}
			c.OnChaosCreated()
			return nil
		}
		return nil // Returning nil because the operation was skipped due to status
	case <-ctx.Done():
		// Context was canceled before delay elapsed
		return ctx.Err()
	}
}

func (c *BaseChaos) GetChaosDescription() string {
	return c.Description
}

// GetStartTime returns the time when the chaos experiment started
func (c *BaseChaos) GetStartTime() time.Time {
	return c.startTime
}

// GetEndTime returns the time when the chaos experiment ended
func (c *BaseChaos) GetEndTime() time.Time {
	return c.endTime
}
