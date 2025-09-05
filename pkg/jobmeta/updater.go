package jobmeta

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/hashicorp/nomad/api"
	"golang.org/x/time/rate"
)

// Updater handles updating meta tags in Nomad jobs
type Updater struct {
	client  *api.Client
	limiter *rate.Limiter
}

// NewUpdater creates a new JobMetaUpdater instance
func NewUpdater(limiter *rate.Limiter) (*Updater, error) {
	config := api.DefaultConfig()
	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Nomad client: %w", err)
	}

	return &Updater{
		client:  client,
		limiter: limiter,
	}, nil
}

// UpdateJobMeta updates the meta tags of a specific job and submits the update
func (u *Updater) UpdateJobMeta(ctx context.Context, jobID, namespace string, metaUpdates map[string]string) {
	// Get the current job
	job, _, err := u.client.Jobs().Info(jobID, &api.QueryOptions{
		Namespace: namespace,
	})
	if err != nil {
		log.Printf("failed to get job %s: %v", jobID, err)
		return
	}

	// Initialize meta map if it doesn't exist
	if job.Meta == nil {
		job.Meta = make(map[string]string)
	}

	// Update the meta tags
	for key, value := range metaUpdates {
		job.Meta[key] = value
	}

	// Submit the job update
	writeOpts := &api.WriteOptions{
		Namespace: namespace,
	}

	response, _, err := u.client.Jobs().Register(job, writeOpts)
	if err != nil {
		log.Printf("failed to register job update: %v", err)
		return
	}

	log.Printf("Job %s update submitted successfully. Evaluation ID: %s", jobID, response.EvalID)
}

func (u *Updater) GetJobs(ctx context.Context, namespace string, jobPattern string) ([]*api.JobListStub, error) {
	jobs, _, err := u.client.Jobs().List(&api.QueryOptions{
		Namespace: namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	matchingJobs := make([]*api.JobListStub, 0)
	for _, jobStub := range jobs {
		// Check if job matches pattern (simple prefix matching for now)
		if strings.HasPrefix(jobStub.ID, jobPattern) {
			matchingJobs = append(matchingJobs, jobStub)
		}
	}
	return matchingJobs, nil
}

// UpdateMultipleJobs updates meta tags for multiple jobs based on a pattern or list
func (u *Updater) UpdateMultipleJobs(ctx context.Context, namespace string, jobPattern string, metaUpdates map[string]string) error {
	// List jobs in the namespace
	jobs, err := u.GetJobs(ctx, namespace, jobPattern)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	log.Printf("Found %d jobs to update", len(jobs))
	wg := sync.WaitGroup{}
	for _, jobStub := range jobs {

		wg.Add(1)
		go func() {
			defer wg.Done()
			if u.limiter != nil {
				if err := u.limiter.Wait(ctx); err != nil {
					log.Printf("failed to wait for permission: %v", err)
					return
				}
			}
			u.UpdateJobMeta(ctx, jobStub.ID, namespace, metaUpdates)
		}()
	}

	log.Print("Waiting for jobs to finish")
	wg.Wait()
	log.Printf("Finish sending requests %d jobs", len(jobs))
	return nil
}
