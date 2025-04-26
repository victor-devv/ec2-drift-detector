package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/victor-devv/ec2-drift-detector/internal/common/errors"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/service"
	"github.com/victor-devv/ec2-drift-detector/pkg/comparator"
)

// DriftDetectorService implements the drift detection service
type DriftDetectorService struct {
	awsProvider        service.InstanceProvider
	terraformProvider  service.InstanceProvider
	repository         service.DriftRepository
	reporters          []service.Reporter
	logger             *logging.Logger
	comparator         *comparator.Comparator
	sourceOfTruth      model.ResourceOrigin
	attributePaths     []string
	parallelChecks     int
	timeout            time.Duration
	scheduleExpression string
	scheduler          *cron.Cron
}

// DriftDetectorConfig holds the configuration for the drift detector service
type DriftDetectorConfig struct {
	SourceOfTruth      model.ResourceOrigin
	AttributePaths     []string
	ParallelChecks     int
	Timeout            time.Duration
	ScheduleExpression string
}

// NewDriftDetectorService creates a new drift detector service
func NewDriftDetectorService(
	awsProvider service.InstanceProvider,
	terraformProvider service.InstanceProvider,
	repository service.DriftRepository,
	reporters []service.Reporter,
	config DriftDetectorConfig,
	logger *logging.Logger,
) *DriftDetectorService {
	logger = logger.WithField("component", "drift-detector")

	return &DriftDetectorService{
		awsProvider:        awsProvider,
		terraformProvider:  terraformProvider,
		repository:         repository,
		reporters:          reporters,
		logger:             logger,
		comparator:         comparator.NewComparator(),
		sourceOfTruth:      config.SourceOfTruth,
		attributePaths:     config.AttributePaths,
		parallelChecks:     config.ParallelChecks,
		timeout:            config.Timeout,
		scheduleExpression: config.ScheduleExpression,
		scheduler:          cron.New(),
	}
}

// DetectAndReportDrift detects and reports drift for a single instance
func (s *DriftDetectorService) DetectAndReportDrift(ctx context.Context, instanceID string, attributePaths []string) error {
	s.logger.Info(fmt.Sprintf("Detecting and reporting drift for instance %s", instanceID))

	// Use specified attributes or default to configured ones
	attrs := attributePaths
	if len(attrs) == 0 {
		attrs = s.attributePaths
	}

	// Detect drift
	result, err := s.DetectDriftByID(ctx, instanceID, attrs)
	if err != nil {
		return err
	}

	// Report drift
	return s.reportDrift(result)
}

// DetectAndReportDriftForAll detects and reports drift for all instances
func (s *DriftDetectorService) DetectAndReportDriftForAll(ctx context.Context, attributePaths []string) error {
	s.logger.Info("Detecting and reporting drift for all instances")

	// Use specified attributes or default to configured ones
	attrs := attributePaths
	if len(attrs) == 0 {
		attrs = s.attributePaths
	}

	// Detect drift
	results, err := s.DetectDriftForAll(ctx, attrs)
	if err != nil {
		return err
	}

	// Report drift
	return s.reportMultipleDrifts(results)
}

// DetectDrift detects drift between two instances for specified attributes
func (s *DriftDetectorService) DetectDrift(ctx context.Context, source, target *model.Instance, attributePaths []string) (*model.DriftResult, error) {
	s.logger.Info(fmt.Sprintf("Detecting drift for instance %s", source.ID))

	// Create a drift result
	result := model.NewDriftResult(source.ID, source.Origin)

	// Compare attributes
	drifts := model.CompareAttributes(source, target, attributePaths)
	if len(drifts) > 0 {
		result.SetDriftedAttributes(drifts)
		s.logger.Info(fmt.Sprintf("Detected %d drifted attributes for instance %s", len(drifts), source.ID))
	}

	// Store the result
	if err := s.repository.SaveDriftResult(ctx, result); err != nil {
		return nil, errors.NewOperationalError(fmt.Sprintf("Failed to save drift result for instance %s", source.ID), err)
	}

	return result, nil
}

// DetectDriftByID detects drift for an instance by ID
func (s *DriftDetectorService) DetectDriftByID(ctx context.Context, instanceID string, attributePaths []string) (*model.DriftResult, error) {
	s.logger.Info(fmt.Sprintf("Detecting drift for instance %s", instanceID))

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Get the instance from both providers
	var awsInstance, terraformInstance *model.Instance
	var awsErr, terraformErr error

	// Get instances concurrently
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		awsInstance, awsErr = s.awsProvider.GetInstance(ctx, instanceID)
		if awsErr != nil {
			s.logger.Error(fmt.Sprintf("Failed to get AWS instance %s: %v", instanceID, awsErr))
		}
	}()

	go func() {
		defer wg.Done()
		terraformInstance, terraformErr = s.terraformProvider.GetInstance(ctx, instanceID)
		if terraformErr != nil {
			s.logger.Error(fmt.Sprintf("Failed to get Terraform instance %s: %v", instanceID, terraformErr))
		}
	}()

	wg.Wait()

	// Check for errors
	if awsErr != nil && terraformErr != nil {
		return nil, errors.NewOperationalError(fmt.Sprintf("Failed to get instance %s from both providers", instanceID), nil)
	}

	if awsErr != nil {
		return nil, errors.NewOperationalError(fmt.Sprintf("Failed to get AWS instance %s", instanceID), awsErr)
	}

	if terraformErr != nil {
		return nil, errors.NewOperationalError(fmt.Sprintf("Failed to get Terraform instance %s", instanceID), terraformErr)
	}

	// Determine source and target based on source of truth
	var source, target *model.Instance
	if s.sourceOfTruth == model.OriginAWS {
		source = awsInstance
		target = terraformInstance
	} else {
		source = terraformInstance
		target = awsInstance
	}

	// Detect drift
	return s.DetectDrift(ctx, source, target, attributePaths)
}

// DetectDriftForAll detects drift for all instances
func (s *DriftDetectorService) DetectDriftForAll(ctx context.Context, attributePaths []string) ([]*model.DriftResult, error) {
	s.logger.Info("Detecting drift for all instances")

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Get all instances from both providers
	var awsInstances, terraformInstances []*model.Instance
	var awsErr, terraformErr error

	// Get instances concurrently
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		awsInstances, awsErr = s.awsProvider.ListInstances(ctx)
		if awsErr != nil {
			s.logger.Error(fmt.Sprintf("Failed to list AWS instances: %v", awsErr))
		}
	}()

	go func() {
		defer wg.Done()
		terraformInstances, terraformErr = s.terraformProvider.ListInstances(ctx)
		if terraformErr != nil {
			s.logger.Error(fmt.Sprintf("Failed to list Terraform instances: %v", terraformErr))
		}
	}()

	wg.Wait()

	// Check for errors
	if awsErr != nil && terraformErr != nil {
		return nil, errors.NewOperationalError("Failed to list instances from both providers", nil)
	}

	if awsErr != nil {
		return nil, errors.NewOperationalError("Failed to list AWS instances", awsErr)
	}

	if terraformErr != nil {
		return nil, errors.NewOperationalError("Failed to list Terraform instances", terraformErr)
	}

	// Map instances by ID for easier lookup
	awsInstanceMap := make(map[string]*model.Instance)
	terraformInstanceMap := make(map[string]*model.Instance)

	for _, instance := range awsInstances {
		awsInstanceMap[instance.ID] = instance
	}

	for _, instance := range terraformInstances {
		terraformInstanceMap[instance.ID] = instance
	}

	// Get the union of all instance IDs
	instanceIDs := make(map[string]bool)
	for id := range awsInstanceMap {
		instanceIDs[id] = true
	}
	for id := range terraformInstanceMap {
		instanceIDs[id] = true
	}

	// Detect drift for each instance
	var results []*model.DriftResult
	var resultsMutex sync.Mutex
	var errs []error
	var errorsMutex sync.Mutex

	// Use a semaphore to limit concurrent operations
	sem := make(chan struct{}, s.parallelChecks)
	var wgDrift sync.WaitGroup

	for id := range instanceIDs {
		wgDrift.Add(1)
		go func(instanceID string) {
			defer wgDrift.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Get instances from both providers
			awsInstance := awsInstanceMap[instanceID]
			terraformInstance := terraformInstanceMap[instanceID]

			// Skip if an instance doesn't exist in one of the providers
			if awsInstance == nil || terraformInstance == nil {
				// Create a result indicating the instance only exists in one provider
				result := model.NewDriftResult(instanceID, s.sourceOfTruth)
				if awsInstance == nil {
					result.AddDriftedAttribute("exists", false, true)
					s.logger.Warn(fmt.Sprintf("Instance %s exists in Terraform but not in AWS", instanceID))
				} else {
					result.AddDriftedAttribute("exists", true, false)
					s.logger.Warn(fmt.Sprintf("Instance %s exists in AWS but not in Terraform", instanceID))
				}

				// Save the result
				resultsMutex.Lock()
				results = append(results, result)
				resultsMutex.Unlock()

				// Store the result
				if err := s.repository.SaveDriftResult(ctx, result); err != nil {
					errorsMutex.Lock()
					errs = append(errs, err)
					errorsMutex.Unlock()
				}

				return
			}

			// Determine source and target based on source of truth
			var source, target *model.Instance
			if s.sourceOfTruth == model.OriginAWS {
				source = awsInstance
				target = terraformInstance
			} else {
				source = terraformInstance
				target = awsInstance
			}

			// Detect drift
			result, err := s.DetectDrift(ctx, source, target, attributePaths)
			if err != nil {
				errorsMutex.Lock()
				errs = append(errs, err)
				errorsMutex.Unlock()
				return
			}

			resultsMutex.Lock()
			results = append(results, result)
			resultsMutex.Unlock()
		}(id)
	}

	wgDrift.Wait()

	// Check for errors
	if len(errs) > 0 {
		return results, errors.NewOperationalError(fmt.Sprintf("Failed to detect drift for %d instances", len(errs)), nil)
	}

	return results, nil
}

// RunScheduledDriftCheck runs a scheduled drift check
func (s *DriftDetectorService) RunScheduledDriftCheck(ctx context.Context) error {
	s.logger.Info("Running scheduled drift check")
	return s.DetectAndReportDriftForAll(ctx, nil)
}

// reportDrift reports a single drift detection result
func (s *DriftDetectorService) reportDrift(result *model.DriftResult) error {
	s.logger.Info(fmt.Sprintf("Reporting drift for instance %s", result.ResourceID))

	// Report drift using all configured reporters
	for _, reporter := range s.reporters {
		if err := reporter.ReportDrift(result); err != nil {
			s.logger.Error(fmt.Sprintf("Failed to report drift for instance %s: %v", result.ResourceID, err))
			return errors.NewOperationalError(fmt.Sprintf("Failed to report drift for instance %s", result.ResourceID), err)
		}
	}

	return nil
}

// reportMultipleDrifts reports multiple drift detection results
func (s *DriftDetectorService) reportMultipleDrifts(results []*model.DriftResult) error {
	s.logger.Info(fmt.Sprintf("Reporting drift for %d instances", len(results)))

	// Report drift using all configured reporters
	for _, reporter := range s.reporters {
		if err := reporter.ReportMultipleDrifts(results); err != nil {
			s.logger.Error(fmt.Sprintf("Failed to report drift for multiple instances: %v", err))
			return errors.NewOperationalError("Failed to report drift for multiple instances", err)
		}
	}

	return nil
}

// StartScheduler starts the scheduler
func (s *DriftDetectorService) StartScheduler(ctx context.Context) error {
	s.logger.Info(fmt.Sprintf("Starting scheduler with expression: %s", s.scheduleExpression))

	if s.scheduleExpression == "" {
		return errors.NewValidationError("Schedule expression cannot be empty")
	}

	// Create a new scheduler
	s.scheduler = cron.New()

	// Add the scheduled drift check
	_, err := s.scheduler.AddFunc(s.scheduleExpression, func() {
		ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
		defer cancel()

		if err := s.RunScheduledDriftCheck(ctx); err != nil {
			s.logger.Error(fmt.Sprintf("Scheduled drift check failed: %v", err))
		}
	})

	if err != nil {
		return errors.NewOperationalError("Failed to add scheduled drift check", err)
	}

	// Start the scheduler
	s.scheduler.Start()

	return nil
}

// StopScheduler stops the scheduler
func (s *DriftDetectorService) StopScheduler() {
	s.logger.Info("Stopping scheduler")

	if s.scheduler != nil {
		s.scheduler.Stop()
	}
}

// SetSourceOfTruth sets the source of truth
func (s *DriftDetectorService) SetSourceOfTruth(sourceOfTruth model.ResourceOrigin) {
	s.sourceOfTruth = sourceOfTruth
}

// SetAttributePaths sets the attribute paths to check
func (s *DriftDetectorService) SetAttributePaths(attributePaths []string) {
	s.attributePaths = attributePaths
}

// SetParallelChecks sets the number of parallel checks
func (s *DriftDetectorService) SetParallelChecks(parallelChecks int) {
	s.parallelChecks = parallelChecks
}

// SetTimeout sets the timeout for drift detection operations
func (s *DriftDetectorService) SetTimeout(timeout time.Duration) {
	s.timeout = timeout
}

// SetScheduleExpression sets the schedule expression
func (s *DriftDetectorService) SetScheduleExpression(expression string) {
	s.scheduleExpression = expression
}

// GetAttributePaths returns the attribute paths to check
func (s *DriftDetectorService) GetAttributePaths() []string {
	return s.attributePaths
}

// GetSourceOfTruth returns the source of truth
func (s *DriftDetectorService) GetSourceOfTruth() model.ResourceOrigin {
	return s.sourceOfTruth
}

// GetParallelChecks returns the number of parallel checks
func (s *DriftDetectorService) GetParallelChecks() int {
	return s.parallelChecks
}

// GetTimeout returns the timeout for drift detection operations
func (s *DriftDetectorService) GetTimeout() time.Duration {
	return s.timeout
}

// GetScheduleExpression returns the schedule expression
func (s *DriftDetectorService) GetScheduleExpression() string {
	return s.scheduleExpression
}

// SetReporters updates the reporters based on the reporter type
func (s *DriftDetectorService) SetReporters(reporters []service.Reporter) {
	s.logger.Info("Updating reporters")
	s.reporters = reporters
}
