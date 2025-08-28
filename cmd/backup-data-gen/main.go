package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/nomad/api"
	"github.com/tcordingly-godaddy/plat-v2-tools/pkg/utils/appexec"
	"github.com/tcordingly-godaddy/plat-v2-tools/pkg/utils/datagen"
)

var (
	jobID                = flag.String("jobId", "", "Nomad job ID to execute commands on (optional)")
	accountID            = flag.String("accountId", "", "Account ID to find all jobs for (optional)")
	customCmd            = flag.String("cmd", "", "Custom command to run on the app (optional)")
	sizeDistributionType = flag.String("size", "medium", "Size distribution for backup generation: medium or large (default: medium)")
	baseRootDir          = flag.String("rootDir", "./wp-content/backup-gen", "Base root directory for backup generation (default: ./wp-content/backup-gen)")
	maxFiles             = flag.Int("maxFiles", 30, "Maximum files per directory (default: 30)")
)

func main() {
	flag.Parse()

	// Create Nomad client
	nomadClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatalf("Error creating Nomad client: %v", err)
	}

	// Create appExec that finds all the details for an app-unit to exec to run commands
	appExec := appexec.NewAppExec(nomadClient, appexec.ExecConcurrency)

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
}

func run(appExec *appexec.AppExec, jobs []string, backupsDataGen *datagen.BackupDataGen) {
	for _, job := range jobs {
		cmds := backupsDataGen.GenerateBackupDataOnApp()
		runMultipleExec(appExec, job, cmds)
	}
}

func runMultipleExec(appExec *appexec.AppExec, jobID string, cmds []string) {
	for _, cmd := range cmds {
		appExec.WaitForAppExec()
		go func() {
			defer appExec.ReleaseAppExec()
			// blocking call to sync the service
			runSingleExec(appExec, jobID, cmd)
		}()
	}
}

func runSingleExec(appExec *appexec.AppExec, jobID string, command string) {
	log.Printf("Executing command on job %s", jobID)
	resp, err := appExec.ExecuteCommandOnApp(context.Background(), jobID, command)
	if err != nil {
		log.Printf("Error executing command on job %s: %v", jobID, err)
		return
	}
	log.Printf("Command executed successfully on job %s. Exit code: %d", jobID, resp.ExitCode)
	if resp.Stdout != "" {
		log.Printf("Stdout for job %s: %s", jobID, resp.Stdout)
	}
	if resp.Stderr != "" {
		log.Printf("Stderr for job %s: %s", jobID, resp.Stderr)
	}
}
