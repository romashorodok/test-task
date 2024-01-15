package analytic

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync/atomic"
)

const LARGER_THEN_9_KEY = ">9"

type AnalyticsConsumer struct {
	state  map[string]*atomic.Uint64
	total  atomic.Uint64
	doneCh chan struct{}
}

func counterInit() *atomic.Uint64 {
	var counter atomic.Uint64
	counter.Add(1)
	return &counter
}

func (a *AnalyticsConsumer) Consume(event chan int) {
	for wordLen := range event {
		switch {
		case wordLen <= 9 && wordLen > 0:
			key := fmt.Sprint(wordLen)
			counter, exist := a.state[key]
			if !exist {
				a.state[key] = counterInit()
				a.total.Add(1)
				continue
			}
			counter.Add(1)
			a.total.Add(1)

		case wordLen > 9:
			counter, exist := a.state[LARGER_THEN_9_KEY]
			if !exist {
				a.state[LARGER_THEN_9_KEY] = counterInit()
				a.total.Add(1)
				continue
			}
			counter.Add(1)
			a.total.Add(1)

		default:
			log.Println("Consumed zero or negative value.")
		}
	}
	close(a.doneCh)
}

func (a *AnalyticsConsumer) ShowAnalytics(filename string) {
	result := fmt.Sprintf("%s:  ", filename)

	for i := 0; i <= 9; i++ {
		counter, exists := a.state[fmt.Sprint(i)]
		if !exists {
			continue
		}
		result += fmt.Sprintf("[%s] = %s, ", fmt.Sprint(i), strconv.FormatUint(counter.Load(), 10))
	}

	counter, exists := a.state[LARGER_THEN_9_KEY]
	if exists {
		result += fmt.Sprintf("[%s] = %s", LARGER_THEN_9_KEY, strconv.FormatUint(counter.Load(), 10))
	} else {
		result = strings.TrimRight(result, ", ")
	}

	log.Println(result)
	log.Printf("%s:  total words: %s", filename, strconv.FormatUint(a.total.Load(), 10))
}

func (a *AnalyticsConsumer) Done() chan struct{} {
	return a.doneCh
}

func NewAnalyticsConsumer() *AnalyticsConsumer {
	return &AnalyticsConsumer{
		state:  make(map[string]*atomic.Uint64),
		doneCh: make(chan struct{}),
	}
}
