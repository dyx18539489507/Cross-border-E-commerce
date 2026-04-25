package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/logger"
)

type DistributionScheduler struct {
	service  *services.DistributionService
	interval time.Duration
	log      *logger.Logger

	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
	doneCh  chan struct{}
}

func NewDistributionScheduler(service *services.DistributionService, interval time.Duration, log *logger.Logger) *DistributionScheduler {
	if interval <= 0 {
		interval = 20 * time.Second
	}

	return &DistributionScheduler{
		service:  service,
		interval: interval,
		log:      log,
	}
}

func (s *DistributionScheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}

	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})
	s.running = true

	go s.loop()
}

func (s *DistributionScheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}

	stopCh := s.stopCh
	doneCh := s.doneCh
	s.running = false
	s.mu.Unlock()

	close(stopCh)
	<-doneCh
}

func (s *DistributionScheduler) loop() {
	defer close(s.doneCh)

	s.runOnce()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.runOnce()
		case <-s.stopCh:
			return
		}
	}
}

func (s *DistributionScheduler) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), s.interval)
	defer cancel()

	if err := s.service.ProcessPendingWork(ctx, 50); err != nil {
		s.log.Errorw("Distribution scheduler tick failed", "error", err)
	}
}
