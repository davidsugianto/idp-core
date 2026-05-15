package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/davidsugianto/go-pkgs/grace"
	"github.com/davidsugianto/go-pkgs/logs"
	cronHandler "github.com/davidsugianto/idp-core/internal/handler/cron"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/davidsugianto/idp-core/internal/pkg/redislock"
	"github.com/davidsugianto/idp-core/internal/pkg/webhook"
	budgetUC "github.com/davidsugianto/idp-core/internal/usecase/budget"
	costUC "github.com/davidsugianto/idp-core/internal/usecase/cost"
	rightsizingUC "github.com/davidsugianto/idp-core/internal/usecase/rightsizing"
	"github.com/robfig/cron/v3"
)

type Server struct {
	c         *cron.Cron
	schedules map[string]string
	port      int
	handler   *cronHandler.Handler
	config    *config.Config
	distlock  redislock.IMutex
	handlers  map[string]func(ctx context.Context) error
}

type Dependencies struct {
	Schedules         map[string]string
	Port              int
	CostUseCase       costUC.Usecase
	BudgetUseCase     budgetUC.Usecase
	RightsizingUseCase rightsizingUC.Usecase
	Config            *config.Config
	Distlock          redislock.IMutex
	WebhookValidator  *webhook.Validator
}

func New(deps Dependencies) *Server {
	return &Server{
		c:         cron.New(),
		schedules: deps.Schedules,
		port:      deps.Port,
		handler: cronHandler.New(cronHandler.Dependencies{
			CostUseCase:        deps.CostUseCase,
			BudgetUseCase:      deps.BudgetUseCase,
			RightsizingUseCase: deps.RightsizingUseCase,
			AuthConfig:         &deps.Config.Auth,
			WebhookValidator:   deps.WebhookValidator,
		}),
		config:   deps.Config,
		distlock: deps.Distlock,
		handlers: make(map[string]func(context.Context) error),
	}
}

func (s *Server) Run(ctx context.Context, graceTimeOut time.Duration) {
	// register cron job
	s.register(ctx, "ping", s.handler.Ping)
	s.register(ctx, "cost-sync", s.handler.FinopsSyncCosts)
	s.register(ctx, "budget-alert-check", s.handler.BudgetAlertCheck)
	s.register(ctx, "rightsizing-generate", s.handler.RightsizingGenerate)

	httpServer := s.httpServer(ctx)
	go func() {
		grace.ServeHTTP(fmt.Sprintf(":%d", s.port), httpServer.Handler)
	}()

	term := make(chan os.Signal, 1)

	s.c.Start()

	logs.Info("Cron Started...")

	// SIGHUP is for handling upstart reload
	signal.Notify(term, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	if <-term != nil {
		logs.Info("signal terminate detected")
		s.stop(graceTimeOut)
	}

	logs.Info("👋")
}

func (s *Server) httpServer(ctx context.Context) *http.Server {
	return &http.Server{
		Addr: ":" + strconv.Itoa(s.port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			defer func() {
				if r := recover(); r != nil {
					logs.Errorf("Recovered from panic in job %v: %v", path, r)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()

			if len(path) > 0 && path[0] == '/' {
				path = path[1:]
			}
			handler, ok := s.handlers[path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			go func() {
				err := handler(ctx)
				if err != nil {
					logs.Errorf("Error in handler %v: %v", path, err)
				}
			}()

			w.WriteHeader(http.StatusOK)
		}),
	}
}

func (s *Server) jobFunc(ctx context.Context, handler func(context.Context) error, jobName string) func() {
	mutex, err := s.distlock.NewMutexW(redislock.MutexOpt{
		Key:          jobName,
		LockTime:     120 * time.Second,
		AutoExtend:   true,
		CheckTTLTime: 1 * time.Second,
	})
	if err != nil {
		return func() {
			logs.Errorf("failed initiate mutex for job %s, %v", jobName, err)
		}
	}

	return func() {
		logs.Infof("running job %s", jobName)
		defer logs.Infof("job %s done", jobName)

		err := mutex.GetLock()
		if err != nil {
			logs.Errorf("failed acquire lock, %v", err)
			return
		}
		defer mutex.Release()

		err = handler(ctx)
		if err != nil {
			logs.Errorf("job %s failed, %v", jobName, err)
		}
	}
}

func (s *Server) register(ctx context.Context, scheduleName string, handler func(context.Context) error) {
	s.handlers[scheduleName] = handler
	if schedule, ok := s.schedules[scheduleName]; ok {
		entryID, err := s.c.AddFunc(schedule, s.jobFunc(ctx, handler, scheduleName))
		if err != nil {
			logs.Errorf(fmt.Sprintf("Task %v (%v) scheduled at %v fail, reason: ", scheduleName, entryID, schedule), err)
			return
		}
		logs.Infof("Task %v (%v) scheduled at %v", scheduleName, entryID, schedule)
	} else {
		logs.Infof("Task named %v schedule isn't available", scheduleName)
	}
}

func (s *Server) stop(timeout time.Duration) {
	ctx := s.c.Stop()

	logs.Infof("waiting job to be done. will terminated after %vs", timeout.Seconds())

	select {
	case <-time.After(timeout):
		logs.Errorf("wait timed out, unfinished job aborted")
		return
	case <-ctx.Done():
		logs.Infof("success shutdown gracefully")
		return
	}
}