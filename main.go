package main

import (
	"fmt"
	"sync"
	"time"
)

const runWorkersInterval = time.Second * 10

type Result struct {
	jobId     string
	jobStatus string
}

func process(jobId string, resultChan chan<- Result, cancelChan <-chan struct{}) {
	//fmt.Println("Job started", jobId)

	//To simulate the behaviour of some jobs may take long time time
	if time.Now().Local().Nanosecond()%2 == 0 {
		time.Sleep(10 * time.Second)
	}
	//fmt.Println("Job Completed", jobId)
	result := Result{jobId: jobId, jobStatus: "Completed"}
	resultChan <- result
	return
}

func worker(jobsChan <-chan string, resultChan chan<- Result, cancelChan <-chan struct{}) {
	for {
		select {
		case <-cancelChan:
			fmt.Println("Signal received")
			return
		case jobId := <-jobsChan:
			if jobId != "" {
				process(jobId, resultChan, cancelChan)
			}
		}
	}
}

func CollectResults(jobs []string) {
	timeStart := time.Now()
	var ResultList []Result

	var workerCount = 5
	jobsChan := make(chan string, len(jobs))
	resultChan := make(chan Result, len(jobs))

	//create a cancel channel
	cancelChan := make(chan struct{})

	// This is to wait for goroutines before exiting
	var wg sync.WaitGroup
	wg.Add(workerCount)

	// This starts up workerCount workers, initially blocked because there are no jobs yet.
	for w := 1; w <= workerCount; w++ {
		go func() {
			defer wg.Done()
			worker(jobsChan, resultChan, cancelChan)
		}()
	}

	// Here we send all jobs to jobsAgentConfigChan and then close that channel to indicate that’s all the work we have.
	for _, job := range jobs {
		jobsChan <- job
	}
	close(jobsChan)

	//Loop will timeout after at max 7 seconds, since new iteration will anyways start after 10 seconds
	//We dont want it to be blocked forever in case if some jobs took long time
	timeout := time.After(7 * time.Second)
	var count int
	for count = 0; count < len(jobs); count++ {
		select {
		case Result := <-resultChan:
			ResultList = append(ResultList, Result)
		case <-timeout:
			fmt.Println(fmt.Sprintf("Could not get result of all jobs, some jobs might have taken long, Duration: %s", time.Since(timeStart)))
			goto FinalResult
		}
	}

FinalResult:
	//cancel the ongoing goroutine(s) if any, close the cancel channel
	close(cancelChan)
	wg.Wait()
	fmt.Println(fmt.Sprintf("Completed %d jobs, Duration %s, Jobs Result %+v", count, time.Since(timeStart), ResultList))
}

func main() {
	Jobs := []string{"job-1", "job-2", "job-3", "job-4", "job-5", "job-6", "job-7", "job-8", "job-9", "job-10"}
	go func() {
		ticker := time.NewTicker(runWorkersInterval)
		defer ticker.Stop()

		for {
			CollectResults(Jobs)
			<-ticker.C
		}
	}()
	time.Sleep(60 * time.Second)
}
