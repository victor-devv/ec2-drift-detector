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
	"github.com/victor-devv/ec2-drift-detector/internal/app"
	"github.com/victor-devv/ec2-drift-detector/internal/common/errors"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/service"
	"github.com/victor-devv/ec2-drift-detector/internal/presentation/reporter"
)

// Handler handles CLI commands
type Handler struct {
	app          *app.Application
	logger       *logging.Logger
	configLoader *config.ConfigLoader
	config       *config.Config
	errorHandler *errors.ErrorHandler
	rootCmd      *cobra.Command
}

// NewHandler creates a new CLI handler
func NewHandler(application *app.Application, configLoader *config.ConfigLoader, cfg *config.Config, logger *logging.Logger) *Handler {
	logger = logger.WithField("component", "cli-handler")
	errorHandler := errors.NewErrorHandler(logger)

	h := &Handler{
		app:          application,
		logger:       logger,
		configLoader: configLoader,
		config:       cfg,
		errorHandler: errorHandler,
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
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(h.config.Detector.TimeoutSeconds)*time.Second)
			defer cancel()

			if len(args) > 0 {
				// Detect drift for a specific instance
				instanceID := args[0]
				h.logger.Info(fmt.Sprintf("Detecting drift for instance %s", instanceID))
				return h.app.DriftDetector.DetectAndReportDrift(ctx, instanceID, h.config.Detector.Attributes)
			}

			// Detect drift for all instances
			h.logger.Info("Detecting drift for all instances")
			return h.app.DriftDetector.DetectAndReportDriftForAll(ctx, h.config.Detector.Attributes)
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
			ctx := context.Background()
			if err := h.app.DriftDetector.StartScheduler(ctx); err != nil {
				return err
			}

			// Wait for signal to stop
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			h.logger.Info("Drift detector server started. Press Ctrl+C to stop")
			<-sigCh

			// Stop the scheduler
			h.app.DriftDetector.StopScheduler()
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
			fmt.Printf("Source of Truth: %s\n", h.config.Detector.SourceOfTruth)
			fmt.Printf("Attributes: %s\n", strings.Join(h.config.Detector.Attributes, ", "))
			fmt.Printf("Parallel Checks: %d\n", h.config.Detector.ParallelChecks)
			fmt.Printf("Timeout: %d seconds\n", h.config.Detector.TimeoutSeconds)
			fmt.Printf("Reporter Type: %s\n", h.config.Reporter.Type)

			if h.config.Reporter.Type == "json" || h.config.Reporter.Type == "both" {
				fmt.Printf("Output File: %s\n", h.config.Reporter.OutputFile)
				fmt.Printf("Pretty Print: %v\n", h.config.Reporter.PrettyPrint)
			}

			if h.config.App.ScheduleExpression != "" {
				fmt.Printf("Schedule Expression: %s\n", h.config.App.ScheduleExpression)
			}

			fmt.Printf("Log Level: %s\n", h.config.App.LogLevel)
			fmt.Printf("AWS Region: %s\n", h.config.AWS.Region)

			if h.config.Terraform.UseHCL {
				fmt.Printf("Terraform HCL Directory: %s\n", h.config.Terraform.HCLDir)
			} else {
				fmt.Printf("Terraform State File: %s\n", h.config.Terraform.StateFile)
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
	detector := h.app.DriftDetector

	sourceOfTruth := model.ResourceOrigin(h.config.Detector.SourceOfTruth)
	detector.SetSourceOfTruth(sourceOfTruth)
	detector.SetAttributePaths(h.config.Detector.Attributes)
	detector.SetParallelChecks(h.config.Detector.ParallelChecks)
	detector.SetTimeout(time.Duration(h.config.Detector.TimeoutSeconds) * time.Second)
	detector.SetScheduleExpression(h.config.App.ScheduleExpression)

	// Update reporters based on configuration
	var reporters []service.Reporter

	switch h.config.Reporter.Type {
	case "console":
		reporters = append(reporters, reporter.NewConsoleReporter(h.logger))
	case "json":
		reporters = append(reporters, reporter.NewJSONReporter(h.logger, h.config.Reporter.OutputFile))
	case "both":
		reporters = append(reporters, reporter.NewConsoleReporter(h.logger))
		reporters = append(reporters, reporter.NewJSONReporter(h.logger, h.config.Reporter.OutputFile))
	default:
		h.logger.Warn("Unknown reporter type: %s, using console reporter", h.config.Reporter.Type)
		reporters = append(reporters, reporter.NewConsoleReporter(h.logger))
	}

	detector.SetReporters(reporters)
}

// Execute executes the root command
func (h *Handler) Execute() error {
	return h.rootCmd.Execute()
}

// GetRootCommand returns the root command
func (h *Handler) GetRootCommand() *cobra.Command {
	return h.rootCmd
}
