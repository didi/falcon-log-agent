package worker

import (
	"fmt"
	"testing"
	"time"
)

func TestCreatejobAndDeletejob(t *testing.T) {
	config := &ConfigInfo{
		ID:       1,
		FilePath: "/Users/anbaoyong/Project/test/aby.${%Y-%m-%d-%H}",
	}
	cache := make(chan string, 100)

	go func() {
		time.Sleep(2 * time.Second)
		deleteJob(config)
	}()
	if err := createJob(config, cache); err == nil {
		for line := range cache {
			fmt.Println(line)
		}
	} else {
		fmt.Println("create job failed : %v", err)
	}
}
