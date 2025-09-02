package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/hashicorp/nomad/api"
	"github.com/tcordingly-godaddy/plat-v2-tools/pkg/utils/appexec"
	"github.com/tcordingly-godaddy/plat-v2-tools/pkg/utils/datagen"
)

var (
	jobID                = flag.String("jobId", "", "Nomad job ID to execute commands on (optional)")
	accountID            = flag.String("accountId", "", "Account ID to find all jobs for (optional)")
	customCmd            = flag.String("cmd", "", "Custom command to run on the app (optional)")
	sizeDistributionType = flag.String("size", "medium", "Size distribution for backup generation: medium or large (default: medium)")
	baseRootDir          = flag.String("rootDir", "./wp-content/mwp-perf-data", "Base root directory for backup generation (default: ./wp-content/mwp-perf-data)")
	concurrency          = flag.Int("concurrency", appexec.ExecConcurrency, "Number of concurrent execs (default: 5)")
	maxFiles             = flag.Int("maxFiles", 30, "Maximum files per directory (default: 30)")
	logLevel             = flag.String("logLevel", "info", "Log level: debug or info")
)

func main() {
	flag.Parse()

	start := time.Now()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	slog.SetLogLoggerLevel(slog.LevelInfo)

	if *logLevel == "debug" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	slog.Info("Command arguments", "jobID", *jobID, "accountID", *accountID, "customCmd", *customCmd, "sizeDistributionType", *sizeDistributionType, "baseRootDir", *baseRootDir, "concurrency", *concurrency, "maxFiles", *maxFiles, "logLevel", *logLevel)

	// Create Nomad client
	nomadClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatalf("Error creating Nomad client: %v", err)
	}

	// Create appExec that finds all the details for an app-unit to exec to run commands
	appExec := appexec.NewAppExec(nomadClient, *concurrency)

	// Determine the command to execute
	// With a jobID and customCmd, we can just run a single command on the app
	if *jobID != "" && *customCmd != "" {
		runSingleExec(appExec, *jobID, *customCmd)
		return
	}

	// With a jobID, we can generate the backup data and run it on the app
	backupsDataGen := datagen.NewBackupDataGen(*baseRootDir, *maxFiles, *sizeDistributionType)
	if *jobID != "" {
		runMultipleExec(appExec, *jobID, backupsDataGen.GenerateBackupDataOnApp())
		return
	}

	// Else Run full backup data for all jobs, with optional filter by accountID
	jobs, err := appExec.GetAppJobs(*accountID)
	if err != nil {
		log.Fatalf("Error getting app jobs: %v", err)
	}
	if len(jobs) == 0 {
		log.Fatalf("No jobs found for account ID: %s", *accountID)
	}

	//TODO add context for signal handling
	run(appExec, jobs, backupsDataGen)
	slog.Info(fmt.Sprintf("Completed data generation for %s type on %d jobs\n", *sizeDistributionType, len(jobs)))
	slog.Info(fmt.Sprintf("Total run time with concurrency of %d: %v\n", *concurrency, time.Since(start)))
}

func run(appExec *appexec.AppExec, jobs []string, backupsDataGen *datagen.BackupDataGen) {
	slog.Info("Running data generation on jobs", "numJobs", len(jobs))
	currentJobCount := 0
	for _, job := range jobs {
		currentJobCount++
		slog.Info("Starting data generation on job", "jobID", job, "currentJobCount", currentJobCount, "totalJobs", len(jobs))
		cmds := backupsDataGen.GenerateBackupDataOnApp()
		runMultipleExec(appExec, job, cmds)
	}
}

func runMultipleExec(appExec *appexec.AppExec, jobID string, cmds []string) {
	slog.Info("Running data generation commands", "jobID", jobID, "numCommands", len(cmds))
	cmdCount := 0
	for _, cmd := range cmds {
		appExec.WaitForAppExec()
		cmdCount++
		if cmdCount%10 == 0 {
			slog.Info("Progress on job", "jobID", jobID, "commandsCompleted", cmdCount, "totalCommands", len(cmds))
		}
		go func() {
			defer appExec.ReleaseAppExec()
			// blocking call to sync the service
			runSingleExec(appExec, jobID, cmd)
		}()
	}
	slog.Info("Finished generating data on job", "jobID", jobID)
}

func runSingleExec(appExec *appexec.AppExec, jobID string, command string) {
	slog.Debug(fmt.Sprintf("Executing command on job %s", jobID))
	resp, err := appExec.ExecuteCommandOnApp(context.Background(), jobID, command)
	if err != nil {
		slog.Warn(fmt.Sprintf("Error executing command on job %s", jobID), "error", err)
		return
	}
	slog.Debug("Command executed successfully on job", "jobID", jobID, "exitCode", resp.ExitCode)
	if resp.Stdout != "" {
		slog.Debug(fmt.Sprintf("Stdout for job %s: %s", jobID, resp.Stdout))
	}
	if resp.Stderr != "" {
		slog.Debug(fmt.Sprintf("Stderr for job %s: %s", jobID, resp.Stderr))
	}
}
