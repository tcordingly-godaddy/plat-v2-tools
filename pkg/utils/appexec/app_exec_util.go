package appexec

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/nomad/api"
)

const (
	AppUnitTaskName = "app-unit"
	CustomerUser    = "customer"
	ExecConcurrency = 5 // max concurrent execs
)

// AppExec is a utility for executing commands on app containers to run commands as customer user
type AppExec struct {
	NomadClient   *api.Client
	execSemaphore chan struct{} // buffer chan to act as a semaphore for concurrent execs
}

func NewAppExec(nomadClient *api.Client, execConcurrency int) *AppExec {

	return &AppExec{
		NomadClient:   nomadClient,
		execSemaphore: make(chan struct{}, execConcurrency),
	}
}

func (ae *AppExec) WaitForAppExec() {
	ae.execSemaphore <- struct{}{}
}

func (ae *AppExec) ReleaseAppExec() {
	<-ae.execSemaphore
}

// GetAppJobs finds all jobs for a specific account ID
func (ae *AppExec) GetAppJobs(accountId string) ([]string, error) {
	jobStubs, _, err := ae.NomadClient.Jobs().List(&api.QueryOptions{
		Namespace:  "sites",
		AllowStale: true,
	})
	if err != nil {
		return nil, err
	}

	namePattern := regexp.MustCompile(`^app-[0-9]+$`)
	var jobIDs []string
	for _, job := range jobStubs {
		if namePattern.MatchString(job.Name) {
			jobIDs = append(jobIDs, job.ID)
		}
	}

	// Filter by account ID if provided, needs to call job info for each job to get the meta
	if accountId != "" {
		jobIDs = ae.filterJobIDsByAccountId(jobIDs, accountId)
	}

	return jobIDs, nil
}

// filterJobIDsByAccountId filters job IDs by account ID
func (ae *AppExec) filterJobIDsByAccountId(jobIDs []string, accountId string) []string {
	filteredJobIDs := []string{}
	for _, jobID := range jobIDs {
		job, _, err := ae.NomadClient.Jobs().Info(jobID, &api.QueryOptions{
			Namespace:  "sites",
			AllowStale: true,
		})
		if err != nil {
			// TODO Retry?
			log.Printf("error getting job info for job %s: %v", jobID, err)
			continue
		}
		if job.Meta != nil && job.Meta["account_id"] == accountId {
			filteredJobIDs = append(filteredJobIDs, jobID)
		}
	}
	return filteredJobIDs
}

// GetAppUnitAllocId gets the allocation ID for a given job ID and task name
func (ae *AppExec) GetAppUnitAllocId(jobID string) (string, error) {
	allocs, _, err := ae.NomadClient.Jobs().Allocations(jobID, false, &api.QueryOptions{
		Namespace:  "sites",
		AllowStale: true,
	})
	if err != nil {
		return "", err
	}

	for _, alloc := range allocs {
		// Verify app task exists and is running
		taskState, ok := alloc.TaskStates[AppUnitTaskName]
		if !ok {
			continue
		}
		if taskState.State != "running" {
			continue
		}
		return alloc.ID, nil
	}
	return "", fmt.Errorf("no running allocation found for job %s", jobID)
}

// ExecuteCommandOnApp executes a command on a job's running allocation as a specific user
func (ae *AppExec) ExecuteCommandOnApp(ctx context.Context, jobID, command string) (*ExecResponse, error) {
	// Find the running allocation for this job and task
	allocID, err := ae.GetAppUnitAllocId(jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get allocation ID: %w", err)
	}
	// Build the command to run as the specified user
	execCommand := buildExecAsUserCommand(command)

	// Execute the command on the allocation
	return ae.ExecCommandOnAllocation(ctx, allocID, execCommand)
}

// ExecCommandOnAllocation executes a command on a Nomad allocation
func (ae *AppExec) ExecCommandOnAllocation(ctx context.Context, allocID string, command []string) (*ExecResponse, error) {
	// Get allocation info to verify it's running
	alloc, _, err := ae.NomadClient.Allocations().Info(allocID, &api.QueryOptions{
		Namespace:  "sites",
		AllowStale: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get allocation info: %w", err)
	}
	if alloc == nil {
		return nil, fmt.Errorf("allocation not found: %s", allocID)
	}

	// Execute the command
	stdin := bytes.NewBufferString("")
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	exitCode, err := ae.NomadClient.Allocations().Exec(
		ctx,
		alloc,
		AppUnitTaskName,
		false, // allocate pty
		command,
		stdin,
		stdout,
		stderr,
		nil,
		&api.QueryOptions{
			Namespace:  "sites",
			AllowStale: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("exec failed: %w", err)
	}

	return &ExecResponse{
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}, nil
}

// ExecResponse represents the result of an exec command
type ExecResponse struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// buildExecAsUserCommand creates the command array to execute as a specific user
func buildExecAsUserCommand(cmd string) []string {

	execCmd := []string{
		"su",
		"-l",
		CustomerUser,
		"/bin/bash",
		"-c",
	}

	execCmd = append(execCmd, cmd)
	log.Printf("execCmd: %v", execCmd)

	return execCmd
}
