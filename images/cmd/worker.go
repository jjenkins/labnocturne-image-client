package cmd

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jjenkins/labnocturne/images/internal/service"
	"github.com/jjenkins/labnocturne/images/internal/store"
	"github.com/spf13/cobra"
)

var dryRun bool

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Run background cleanup jobs for expired files",
	Long: `Run cleanup jobs to delete expired test files (7+ days) and
permanently remove soft-deleted files (30+ days). This command runs
once and exits, suitable for cron scheduling.`,
	Run: runWorker,
}

func init() {
	rootCmd.AddCommand(workerCmd)
	workerCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Log what would be deleted without actually deleting")
}

func runWorker(cmd *cobra.Command, args []string) {
	startTime := time.Now()

	if dryRun {
		log.Printf("Starting cleanup worker in DRY RUN mode")
	} else {
		log.Printf("Starting cleanup worker")
	}

	// 1. Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)
		cancel()
	}()

	// 2. Initialize database connection
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := store.NewDB(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Printf("Successfully connected to database")

	// 3. Initialize AWS S3 client
	s3Bucket := os.Getenv("AWS_S3_BUCKET")
	if s3Bucket == "" {
		log.Fatal("AWS_S3_BUCKET environment variable is required")
	}

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		log.Fatal("AWS_REGION environment variable is required")
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(awsRegion),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	log.Printf("Successfully initialized S3 client for bucket: %s", s3Bucket)

	// 4. Run cleanup operations
	cleanupService := service.NewCleanupService(db, s3Client, s3Bucket, dryRun)

	log.Printf("Starting test file cleanup (files older than 7 days)...")
	testCount, testErr := cleanupService.CleanupExpiredTestFiles(ctx)

	log.Printf("Starting soft-deleted file cleanup (files deleted more than 30 days ago)...")
	softCount, softErr := cleanupService.CleanupExpiredSoftDeleted(ctx)

	// 5. Process bandwidth logs (if configured)
	var bandwidthCount int
	var bandwidthErr error

	cloudfrontLogBucket := os.Getenv("CLOUDFRONT_LOG_BUCKET")
	if cloudfrontLogBucket != "" {
		log.Printf("Starting bandwidth log processing...")

		// Process logs from 2 days ago (ensures CloudFront log delivery is complete)
		twoDaysAgo := time.Now().UTC().AddDate(0, 0, -2)

		fileStore := store.NewFileStore(db)
		bandwidthStore := store.NewBandwidthStore(db)
		bandwidthService := service.NewBandwidthService(
			s3Client,
			fileStore,
			bandwidthStore,
			cloudfrontLogBucket,
			"cloudfront/", // Log prefix
		)

		bandwidthCount, bandwidthErr = bandwidthService.ProcessLogFiles(ctx, twoDaysAgo)
		if bandwidthErr != nil {
			log.Printf("warning: bandwidth processing completed with errors: %v", bandwidthErr)
		}
	} else {
		log.Printf("Skipping bandwidth log processing (CLOUDFRONT_LOG_BUCKET not configured)")
	}

	// 6. Log summary
	duration := time.Since(startTime)
	log.Printf("Cleanup worker completed in %v", duration.Round(time.Millisecond))
	log.Printf("Summary: %d test files deleted, %d soft-deleted files permanently removed, %d users bandwidth processed", testCount, softCount, bandwidthCount)

	// 7. Exit with appropriate status
	// Note: Bandwidth errors are non-critical, so we only fail on cleanup errors
	if testErr != nil || softErr != nil {
		combinedErr := errors.Join(testErr, softErr)
		log.Printf("Worker completed with errors: %v", combinedErr)
		os.Exit(1)
	}

	log.Printf("Worker completed successfully")
}
