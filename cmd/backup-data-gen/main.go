package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/nomad/api"
	"github.com/tcordingly-godaddy/plat-v2-tools/pkg/utils/appexec"
	"github.com/tcordingly-godaddy/plat-v2-tools/pkg/utils/datagen"
)

var (
	jobID                = flag.String("jobId", "", "Nomad job ID to execute commands on (optional)")
	accountID            = flag.String("accountId", "", "Account ID to find all jobs for (optional)")
	jobIDsFile           = flag.String("jobIdsFile", "", "File containing list of job IDs (one per line) (optional)")
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

	slog.Info("Command arguments", "jobID", *jobID, "accountID", *accountID, "jobIDsFile", *jobIDsFile, "customCmd", *customCmd, "sizeDistributionType", *sizeDistributionType, "baseRootDir", *baseRootDir, "concurrency", *concurrency, "maxFiles", *maxFiles, "logLevel", *logLevel)

	// Create Nomad client
	nomadClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatalf("Error creating Nomad client: %v", err)
	}

	// Create appExec that finds all the details for an app-unit to exec to run commands
	appExec := appexec.NewAppExec(nomadClient, *concurrency)

	// Determine the command to execute
	// With a jobID specified, we can just run a single command on the app
	if *jobID != "" {
		if *customCmd != "" {
			runSingleExec(appExec, *jobID, *customCmd)
		} else {
			backupsDataGen := datagen.NewBackupDataGen(*baseRootDir, *maxFiles, *sizeDistributionType)
			runSingleExec(appExec, *jobID, backupsDataGen.GenerateBackupDataOnApp())
		}
		return
	}

	// Else get job ids and run stream the command(s) to the app for each job
	var jobs []string
	if *jobIDsFile != "" {
		// Read job IDs from file
		slog.Info("Running backup data generation on jobs from file", "jobIDsFile", *jobIDsFile)
		jobs, err = readJobIDsFromFile(*jobIDsFile)
		if err != nil {
			log.Fatalf("Error reading job IDs from file: %v", err)
		}
	} else {
		// Get jobs from account ID
		jobs, err = appExec.GetAppJobs(*accountID)
		if err != nil {
			log.Fatalf("Error getting app jobs: %v", err)
		}
		if len(jobs) == 0 {
			log.Fatalf("No jobs found for account ID: %s", *accountID)
		}
	}

	//TODO add context for signal handling
	var dataGenFunc func() string
	if *customCmd != "" {
		dataGenFunc = func() string {
			return *customCmd
		}
	} else {
		backupsDataGen := datagen.NewBackupDataGen(*baseRootDir, *maxFiles, *sizeDistributionType)
		dataGenFunc = backupsDataGen.GenerateBackupDataOnApp
	}
	run(appExec, jobs, dataGenFunc)
	slog.Info(fmt.Sprintf("Completed data generation for %s type on %d jobs", *sizeDistributionType, len(jobs)))
	slog.Info(fmt.Sprintf("Total run time with concurrency of %d: %v", *concurrency, time.Since(start)))
}

// readJobIDsFromFile reads job IDs from a file, one per line
func readJobIDsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %v", filename, err)
	}
	defer file.Close()

	var jobIDs []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments (lines starting with #)
		if line != "" && !strings.HasPrefix(line, "#") {
			jobIDs = append(jobIDs, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filename, err)
	}

	if len(jobIDs) == 0 {
		return nil, fmt.Errorf("no valid job IDs found in file %s", filename)
	}

	return jobIDs, nil
}

func run(appExec *appexec.AppExec, jobs []string, dataGenFunc func() string) {
	slog.Info("Running data generation on jobs", "numJobs", len(jobs))

	wg := sync.WaitGroup{}
	currentJobCount := 0
	for _, job := range jobs {
		currentJobCount++
		slog.Info("Starting data generation on job", "jobID", job, "currentJobCount", currentJobCount, "totalJobs", len(jobs))
		cmds := dataGenFunc()
		appExec.WaitForAppExec()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer appExec.ReleaseAppExec()
			// blocking call to sync the service
			slog.Info("Starting exec to job", "jobID", job)
			runSingleExec(appExec, job, cmds)
			slog.Info("Finished exec to job", "jobID", job)
		}()
	}
	wg.Wait()
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
		slog.Info(fmt.Sprintf("Stdout for job %s: %s", jobID, resp.Stdout))
	}
	if resp.Stderr != "" {
		slog.Error(fmt.Sprintf("Stderr for job %s: %s", jobID, resp.Stderr))
	}
}
