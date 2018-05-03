package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"time"
)

var myClient = &http.Client{Timeout: 10 * time.Second}

type JenkinsMasterBranchResponse struct {
	Builds []struct {
		URL string `json:"url"`
	} `json:"builds"`
}

type JenkinsJobResponse struct {
	Result  string `json:"result"`
	Actions []struct {
		Class string `json:"_class,omitempty"`
	}
}

type TestResultsResponse struct {
	FailCount int `json:"failCount"`
	Suites    []struct {
		Cases []struct {
			// TestActions     []interface{} `json:"testActions"`
			// Age             int           `json:"age"`
			// ClassName       string        `json:"className"`
			// Duration        float64       `json:"duration"`
			// ErrorDetails    interface{}   `json:"errorDetails"`
			// ErrorStackTrace interface{}   `json:"errorStackTrace"`
			// FailedSince     int           `json:"failedSince"`
			Name string `json:"name"`
			// Skipped         bool          `json:"skipped"`
			// SkippedMessage  interface{}   `json:"skippedMessage"`
			Status string `json:"status"`
			// Stderr          interface{}   `json:"stderr"`
			// Stdout          interface{}   `json:"stdout"`
		} `json:"cases"`
	} `json:"suites"`
}

func getJson(url string, target interface{}) {
	r, err := myClient.Get(url)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	err = json.NewDecoder(r.Body).Decode(target)
	if err != nil {
		panic(err)
	}
}

func main() {
	if len(os.Args) < 2 {
		panic("Missing URL as a argument")
	}

	failureList := []string{}
	failureMap := make(map[string]int)

	jenkinsJob := os.Args[1]
	jenkinsJobAPI := jenkinsJob + "api/json"
	jenkinsMaster := JenkinsMasterBranchResponse{}
	log.Println("Checking master", jenkinsJobAPI)
	getJson(jenkinsJobAPI, &jenkinsMaster)
	for _, build := range jenkinsMaster.Builds {
		jobInfo := JenkinsJobResponse{}
		getJson(build.URL+"api/json", &jobInfo)
		hasTestResults := false
		for _, action := range jobInfo.Actions {
			if action.Class == "hudson.tasks.junit.TestResultAction" {
				hasTestResults = true
			}
		}
		if !hasTestResults {
			log.Println("Skipping as there is no test results for this build")
			continue
		}
		if jobInfo.Result != "FAILURE" {
			log.Println("Skipping as it was not a FAILURE")
			continue
		}
		testResultsURL := build.URL + "testReport/api/json"
		testResults := TestResultsResponse{}
		log.Println("Checking test results", testResultsURL)
		getJson(testResultsURL, &testResults)
		for _, testSuite := range testResults.Suites {
			for _, testCase := range testSuite.Cases {
				if testCase.Status == "FAILED" {
					failureList = append(failureList, testCase.Name)
				}
			}
		}
	}

	for _, failure := range failureList {
		failureMap[failure] = failureMap[failure] + 1
	}

	type kv struct {
		Key   string
		Value int
	}

	var ss []kv
	for k, v := range failureMap {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	fmt.Println("")
	fmt.Println("### Results (descending sort by most failures)")
	for _, kv := range ss {
		fmt.Printf("%s, %d\n", kv.Key, kv.Value)
	}
}
