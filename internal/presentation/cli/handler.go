package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/victor-devv/ec2-drift-detector/internal/common/errors"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/service"
	"github.com/victor-devv/ec2-drift-detector/internal/presentation/reporter"
)

// Handler handles CLI commands
type Handler struct {
	app          service.DriftDetectorProvider
	logger       *logging.Logger
	configLoader *config.ConfigLoader
	config       *config.Config
	errorHandler *errors.ErrorHandler
	rootCmd      *cobra.Command
	ctx          context.Context
}

// NewHandler creates a new CLI handler
func NewHandler(ctx context.Context, application service.DriftDetectorProvider, configLoader *config.ConfigLoader, cfg *config.Config, logger *logging.Logger) *Handler {
	logger = logger.WithField("component", "cli-handler")
	errorHandler := errors.NewErrorHandler(logger)

	h := &Handler{
		app:          application,
		logger:       logger,
		configLoader: configLoader,
		config:       cfg,
		errorHandler: errorHandler,
		ctx:          ctx,
	}

	h.initCommands()
	return h
}

// initCommands initializes CLI commands
func (h *Handler) initCommands() {
	rootCmd := &cobra.Command{
		Use:   "drift-detector",
		Short: "Terraform drift detector",
		Long:  "A tool to detect drift between AWS EC2 instances and Terraform configurations",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Update configuration from CLI flags
			cliOpts := make(map[string]interface{})

			// Get flags from all commands
			cmd.Flags().Visit(func(f *pflag.Flag) {
				cliOpts[f.Name] = f.Value.String()
			})

			// Update configuration
			if err := h.configLoader.UpdateConfig(h.config, cliOpts); err != nil {
				h.errorHandler.HandleWithExit(err)
			}

			if err := h.config.Validate(); err != nil {
				h.errorHandler.HandleWithExit(err)
			}

			// Update service configuration
			h.updateServiceConfig()
		},
	}

	// Add global flags
	rootCmd.PersistentFlags().String("log-level", "INFO", "Log level (DEBUG, INFO, WARN, ERROR)")
	rootCmd.PersistentFlags().StringP("state-file", "s", "", "Terraform state file path")
	rootCmd.PersistentFlags().String("hcl-dir", "", "Terraform HCL directory path")
	rootCmd.PersistentFlags().String("source-of-truth", "terraform", "Source of truth (aws or terraform)")
	rootCmd.PersistentFlags().StringSliceP("attributes", "a", nil, "Attributes to check for drift")
	rootCmd.PersistentFlags().IntP("parallel-checks", "p", 0, "Number of parallel checks to run")
	rootCmd.PersistentFlags().StringP("output", "o", "", "Output format (json, console, or both)")
	rootCmd.PersistentFlags().StringP("output-file", "f", "", "Output file for JSON (defaults to stdout)")
	rootCmd.PersistentFlags().String("schedule-expression", "", "Cron expression for scheduled drift checks")

	// Add commands
	h.addDetectCommand(rootCmd)
	h.addServerCommand(rootCmd)
	h.addConfigCommand(rootCmd)

	h.rootCmd = rootCmd
}

// addDetectCommand adds the detect command
func (h *Handler) addDetectCommand(rootCmd *cobra.Command) {
	detectCmd := &cobra.Command{
		Use:   "detect [instance-id]",
		Short: "Detect drift for a specific instance or all instances",
		Long:  "Detect drift between AWS EC2 instances and Terraform configurations",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(h.ctx, time.Duration(h.config.GetTimeout())*time.Second)
			defer cancel()

			if len(args) > 0 {
				// Detect drift for a specific instance
				instanceID := args[0]
				h.logger.Info(fmt.Sprintf("Detecting drift for instance %s", instanceID))
				return h.app.DetectAndReportDrift(ctx, instanceID, h.config.GetAttributes())
			}

			// Detect drift for all instances
			h.logger.Info("Detecting drift for all instances")
			return h.app.DetectAndReportDriftForAll(ctx, h.config.GetAttributes())
		},
	}

	rootCmd.AddCommand(detectCmd)
}

// addServerCommand adds the server command
func (h *Handler) addServerCommand(rootCmd *cobra.Command) {
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Run as a server with scheduled drift checks",
		Long:  "Run the drift detector as a server with scheduled drift checks",
		RunE: func(cmd *cobra.Command, args []string) error {
			h.logger.Info("Starting drift detector server")

			// Start the scheduler
			if err := h.app.StartScheduler(h.ctx); err != nil {
				return err
			}

			// Wait for signal to stop
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			h.logger.Info("Drift detector server started. Press Ctrl+C to stop")
			<-sigCh

			// Stop the scheduler
			h.app.StopScheduler()
			h.logger.Info("Drift detector server stopped")

			return nil
		},
	}

	rootCmd.AddCommand(serverCmd)
}

// addConfigCommand adds the config command
func (h *Handler) addConfigCommand(rootCmd *cobra.Command) {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration operations",
		Long:  "View and update configuration",
	}

	// Add show subcommand
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			h.logger.Info("Showing current configuration")

			fmt.Println("Current Configuration:")
			fmt.Println("======================")
			fmt.Printf("Source of Truth: %s\n", h.config.GetSourceOfTruth())
			fmt.Printf("Attributes: %s\n", strings.Join(h.config.GetAttributes(), ", "))
			fmt.Printf("Parallel Checks: %d\n", h.config.GetParallelChecks())
			fmt.Printf("Timeout: %d seconds\n", h.config.GetTimeout())
			reporterType := h.config.GetReporterType()
			fmt.Printf("Reporter Type: %s\n", reporterType)

			if reporterType == "json" || reporterType == "both" {
				fmt.Printf("Output File: %s\n", h.config.GetOutputFile())
				fmt.Printf("Pretty Print: %v\n", h.config.GetPrettyPrint())
			}

			if cronExpression := h.config.GetScheduleExpression(); cronExpression != "" {
				fmt.Printf("Schedule Expression: %s\n", cronExpression)
			}

			fmt.Printf("Log Level: %s\n", h.config.GetLogLevel())
			fmt.Printf("AWS Region: %s\n", h.config.GetAWSRegion())

			if h.config.GetUseHCL() {
				fmt.Printf("Terraform HCL Directory: %s\n", h.config.GetHCLDir())
			} else {
				fmt.Printf("Terraform State File: %s\n", h.config.GetStateFile())
			}

			return nil
		},
	}

	// Add reload subcommand
	reloadCmd := &cobra.Command{
		Use:   "reload",
		Short: "Reload configuration from file",
		RunE: func(cmd *cobra.Command, args []string) error {
			h.logger.Info("Reloading configuration")

			// Reload configuration
			config, err := h.configLoader.ReloadConfig()
			if err != nil {
				return err
			}

			// Update the reference
			h.config = config

			// Update service configuration
			h.updateServiceConfig()

			h.logger.Info("Configuration reloaded successfully")
			return nil
		},
	}

	configCmd.AddCommand(showCmd)
	configCmd.AddCommand(reloadCmd)
	rootCmd.AddCommand(configCmd)
}

// updateServiceConfig updates service configuration from the config object
func (h *Handler) updateServiceConfig() {
	// Update drift detector configuration
	detector := h.app

	sourceOfTruth := model.ResourceOrigin(h.config.GetSourceOfTruth())
	detector.SetSourceOfTruth(sourceOfTruth)
	detector.SetAttributePaths(h.config.GetAttributes())
	detector.SetParallelChecks(h.config.GetParallelChecks())
	detector.SetTimeout(time.Duration(h.config.GetTimeout()) * time.Second)
	detector.SetScheduleExpression(h.config.GetScheduleExpression())

	// Update reporters based on configuration
	var reporters []service.Reporter

	switch h.config.GetReporterType() {
	case "console":
		reporters = append(reporters, reporter.NewConsoleReporter(h.logger))
	case "json":
		reporters = append(reporters, reporter.NewJSONReporter(h.logger, h.config.GetOutputFile()))
	case "both":
		reporters = append(reporters, reporter.NewConsoleReporter(h.logger))
		reporters = append(reporters, reporter.NewJSONReporter(h.logger, h.config.GetOutputFile()))
	default:
		h.logger.Warn("Unknown reporter type: %s, using console reporter", h.config.GetReporterType())
		reporters = append(reporters, reporter.NewConsoleReporter(h.logger))
	}

	detector.SetReporters(reporters)
}

// Execute executes the root command
func (h *Handler) Execute(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		defer close(done)
		h.rootCmd.Execute()
	}()

	select {
	case <-ctx.Done():
		h.logger.Warn("Received interrupt signal, exiting...")
		return ctx.Err()
	case <-done:
		return nil
	}
}

// GetRootCommand returns the root command
func (h *Handler) GetRootCommand() *cobra.Command {
	return h.rootCmd
}
