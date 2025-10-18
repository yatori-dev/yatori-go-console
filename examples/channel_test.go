package examples

import (
	"log"
	"testing"
	"time"
)

func TestChannel(t *testing.T) {
	//var wg sync.WaitGroup
	queue := make(chan int, 3)
	for i := 0; i < 3; i++ {
		queue <- i
	}
	for {
		//wg.Add(1)
		idx := <-queue
		go func(idx int) {
			//defer wg.Done()
			defer func() { queue <- idx }()
			log.Println(idx)
			time.Sleep(5 * time.Second)
		}(idx)
	}
}
