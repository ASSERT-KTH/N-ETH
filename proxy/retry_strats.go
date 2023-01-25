package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"sync"
)

type RetryStragety int

type Strategy interface {
	GetNext() int32
	Init()
	Success()
	Failure()
}

type FixedPriorityStrategy struct {
	index        int32
	priorityList []int32
}

func (s *FixedPriorityStrategy) Init() {
	s.index = 0
	s.priorityList = []int32{0, 1, 2}
}

func (s *FixedPriorityStrategy) GetNext() int32 {
	defer func() { s.index++ }()
	return s.index
}
func (s *FixedPriorityStrategy) Success() {}
func (s *FixedPriorityStrategy) Failure() {}

var rrIndex int32 = 0
var rrIndexMutex = sync.Mutex{}

type RoundRobinStrategy struct {
	index int32
	size  int32
}

func (s *RoundRobinStrategy) Init() {
	s.size = 3
	rrIndexMutex.Lock()
	s.index = rrIndex
	rrIndex = (rrIndex + 1) % s.size
	rrIndexMutex.Unlock()
}

func (s *RoundRobinStrategy) GetNext() int32 {
	defer func() { s.index = (s.index + 1) % s.size }()
	return s.index
}

func (s *RoundRobinStrategy) Success() {}
func (s *RoundRobinStrategy) Failure() {}

type RandomStrategy struct {
	index        int32
	priorityList []int32
}

func (s *RandomStrategy) Init() {
	s.priorityList = []int32{0, 1, 2}
	rand.Shuffle(
		len(s.priorityList),
		func(i int, j int) {
			s.priorityList[i], s.priorityList[j] = s.priorityList[j], s.priorityList[i]
		},
	)
	s.index = 0
}

func (s *RandomStrategy) GetNext() int32 {
	defer func() { s.index++ }()
	return s.index
}

func (s *RandomStrategy) Success() {}
func (s *RandomStrategy) Failure() {}

type AdaptiveScore struct {
	index     int32
	requests  int32
	successes int32
	score     float32
}

func (s *AdaptiveScore) updateScore() {
	s.score = float32(s.successes) / float32(s.requests)
}

type AdaptiveOrder []*AdaptiveScore

func (s *AdaptiveOrder) debug() {
	for _, v := range *s {
		fmt.Printf("index: %d, score: %f\n", v.index, v.score)
		fmt.Printf("index: %d, successes: %d\n", v.index, v.successes)
		fmt.Printf("index: %d, requests: %d\n", v.index, v.requests)
	}
}

func (s *AdaptiveOrder) get(i int32) (*AdaptiveScore, error) {
	for _, v := range *s {
		if v.index == i {
			return v, nil
		}
	}
	return nil, errors.New("index not found")
}

func (s *AdaptiveOrder) sort() {
	sort.Slice(*s, func(i, j int) bool {
		return (*s)[i].score < (*s)[j].score
	})
}

func (s *AdaptiveOrder) getIndexSlice() []int32 {
	r := make([]int32, 0)
	for _, v := range *s {
		r = append(r, v.index)
	}
	return r
}

func (s *AdaptiveOrder) addSuccess(index int32) {
	entry, _ := s.get(index)
	entry.requests++
	entry.successes++
	entry.updateScore()
	s.sort()
}

func (s *AdaptiveOrder) addFailure(index int32) {
	entry, _ := s.get(index)
	entry.requests++
	entry.updateScore()
	s.sort()
}

var adaptiveOrder = AdaptiveOrder{
	&AdaptiveScore{index: 0, requests: 1, successes: 1},
	&AdaptiveScore{index: 1, requests: 1, successes: 1},
	&AdaptiveScore{index: 2, requests: 1, successes: 1},
}
var adaptiveOrderMutex = sync.Mutex{}

type AdaptiveStrategy struct {
	index        int32
	priorityList []int32
}

func (s *AdaptiveStrategy) Init() {
	adaptiveOrderMutex.Lock()
	s.priorityList = adaptiveOrder.getIndexSlice()
	adaptiveOrderMutex.Unlock()
	s.index = 0
}

func (s *AdaptiveStrategy) GetNext() int32 {
	defer func() { s.index++ }()
	return s.priorityList[s.index]
}

func (s *AdaptiveStrategy) Success(index int32) int32 {
	adaptiveOrderMutex.Lock()
	adaptiveOrder.addSuccess(index)
	adaptiveOrderMutex.Unlock()
	adaptiveOrder.debug()
	return s.index
}

func (s *AdaptiveStrategy) Failure(index int32) int32 {
	adaptiveOrderMutex.Lock()
	adaptiveOrder.addFailure(index)
	adaptiveOrderMutex.Unlock()
	adaptiveOrder.debug()
	return s.index
}
