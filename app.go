package bshark

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/spf13/pflag"

	"github.com/kkkbird/bshark/debugserver"
	"github.com/kkkbird/qlog"
)

var log = qlog.WithField("bshark", "application")

func getFuncName(f interface{}) string {
	fv := reflect.ValueOf(f)

	if fv.Kind() != reflect.Func {
		return ""
	}
	return runtime.FuncForPC(fv.Pointer()).Name()
}

// DaemonFunc for bshark app daemon modules
type DaemonFunc func(ctx context.Context) error

// InitFunc for bshark app init modules
type InitFunc func(ctx context.Context) (CleanFunc, error)

// ClearFunc for bshark app, clean init module
type CleanFunc func(ctx context.Context)

//InitStage is executed with add sequence, InitFunc in one init stage will be called concurrently
type InitStage struct {
	name       string
	funcs      []InitFunc
	cleanFuncs []CleanFunc
}

// Run run a InitStage
func (s *InitStage) Run(ctx context.Context, a *Application) error {
	var wg sync.WaitGroup

	wg.Add(len(s.funcs))

	for _, fc := range s.funcs {
		go func(_fc InitFunc) {
			defer wg.Done()

			funcName := getFuncName(_fc)

			defer func() {
				if r := recover(); r != nil {
					a.initErrChan <- fmt.Errorf("%s() panic:%s", funcName, r)
				}
			}()

			log.Tracef("  %s() start...", funcName)

			cleanFunc, err := _fc(ctx)
			if err != nil {
				a.initErrChan <- fmt.Errorf("%s():%s", funcName, err)
				return
			}
			// no need to add clean func if err != nil
			if cleanFunc != nil {
				if s.cleanFuncs == nil {
					s.cleanFuncs = make([]CleanFunc, 0)
				}
				s.cleanFuncs = append(s.cleanFuncs, cleanFunc)
			}

			log.Tracef("  %s() done!", funcName)
		}(fc)
	}

	wg.Wait()

	return nil
}

// Clean the InitStage
func (s *InitStage) Clean(ctx context.Context, a *Application) error {
	var wg sync.WaitGroup

	if len(s.cleanFuncs) == 0 {
		log.Trace("  nothing to clean")
		return nil
	}

	wg.Add(len(s.cleanFuncs))

	for _, fc := range s.cleanFuncs {
		go func(_fc CleanFunc) {
			defer wg.Done()

			funcName := getFuncName(_fc)

			defer func() {
				if r := recover(); r != nil {
					log.Errorf("clean %s() panic:%s", funcName, r)
				}
			}()

			log.Tracef("  %s() cleaning...", funcName)

			_fc(ctx)

			log.Tracef("  %s() done!", funcName)
		}(fc)
	}

	wg.Wait()

	return nil
}

func newInitStage(name string, funcs []InitFunc) *InitStage {
	return &InitStage{
		name:  name,
		funcs: funcs,
	}
}

// Application is a bshark app
type Application struct {
	initTimeout             time.Duration
	initErrChan             chan error
	cleanTimeout            time.Duration // default 1s
	daemonForceCloseTimeout time.Duration // default 3s

	preInit             func()
	onConfigFileChanged func()
	cmdline             *pflag.FlagSet
	name                string
	initStages          []*InitStage
	daemons             []DaemonFunc
}

// AppOpts is setters for application options
type AppOpts func(a *Application)

// func WithName(name string) AppOpts {
// 	return func(a *Application) {
// 		a.name = name
// 	}
// }

// WithInitTimeout set init with a timeout
func WithInitTimeout(timeout time.Duration) AppOpts {
	return func(a *Application) {
		a.initTimeout = timeout
	}
}

// WithCleanTimeout set init force close timeout
func WithCleanTimeout(timeout time.Duration) AppOpts {
	return func(a *Application) {
		a.cleanTimeout = timeout
	}
}

// WithDaemonForceCloseTimeout set daemon force close timeout
func WithDaemonForceCloseTimeout(timeout time.Duration) AppOpts {
	return func(a *Application) {
		a.daemonForceCloseTimeout = timeout
	}
}

// WithLogger set logger of application
// func WithLogger(logger Logger) AppOpts {
// 	return func(a *Application) {
// 		a.logger = logger
// 	}
// }

// New create a bshark app object
func New(name string, opts ...AppOpts) *Application {
	app := &Application{
		initTimeout:             0, // no timeout
		cleanTimeout:            time.Second,
		daemonForceCloseTimeout: 3 * time.Second,
		cmdline:                 pflag.CommandLine,
		name:                    name,
		initStages:              make([]*InitStage, 0),
		daemons:                 make([]DaemonFunc, 0),

		initErrChan: make(chan error, 1),
	}

	for _, opt := range opts {
		opt(app)
	}

	app.AddInitStage("preInit", app.initParams).AddDaemons(debugserver.Run)

	return app
}

func (a *Application) initParams(ctx context.Context) (CleanFunc, error) {
	a.handleFlagsAndEnv()
	return nil, nil
}

// func (a *Application) printf(format string, args ...interface{}) {
// 	if a.logger == nil {
// 		log.Infof(format, args...)
// 		return
// 	}

// 	a.logger.Printf(format, args...)
// }

// AddInitStage add a stage for bshark app
func (a *Application) AddInitStage(name string, funcs ...InitFunc) *Application {
	a.initStages = append(a.initStages, newInitStage(name, funcs))
	return a
}

// AddDaemons add a daemon for bshark app
func (a *Application) AddDaemons(funcs ...DaemonFunc) *Application {
	a.daemons = append(a.daemons, funcs...)
	return a
}

func (a *Application) runInitStages() error {
	var (
		ctx    = context.Background()
		cancel context.CancelFunc
		err    error
	)

	if a.initTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, a.initTimeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	// run init stages
	for i, s := range a.initStages {
		cErr := make(chan error, 1)

		go func() {
			log.Infof("Init stage %d-%s", i, s.name)
			cErr <- s.Run(ctx, a)
		}()

		select {
		case err = <-cErr:
			if err != nil {
				panic("shoud not run to this line")
				return err
			}
		case err = <-a.initErrChan:
			cancel()
			log.WithError(err).Errorf("!!Init err, exit in 1s")
			select { // wait the init stage done or cleanTimeout duration
			case <-cErr:
			case <-time.After(time.Second):
			}

			return err
		case <-ctx.Done():
			log.Errorf("!!Init timeount, exit in 1s")
			select { // wait the init stage done or cleanTimeout duration
			case <-cErr:
			case <-time.After(time.Second):
			}

			return ctx.Err()
		}
	}

	return nil
}

func (a *Application) runCleanStage() {
	var (
		ctx    = context.Background()
		cancel context.CancelFunc
	)

	if a.cleanTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, a.cleanTimeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	// run clean stage in reverse order
	for i := len(a.initStages) - 1; i >= 0; i-- {
		s := a.initStages[i]
		cErr := make(chan error, 1)

		go func() {
			log.Infof("Clean stage %d-%s", i, s.name)
			cErr <- s.Clean(ctx, a)
		}()

		select {
		case <-cErr: //ingore err, just continue clean
		case <-ctx.Done():
			log.Warn("!!Clean timeount")
			return
		}
	}
}

func (a *Application) runDaemons() error {

	var (
		ctx    context.Context
		cancel context.CancelFunc
		cErr   = make(chan error, len(a.daemons))
		cDone  = make(chan interface{}, 1)
	)

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// run daemon funcs
	go func() {
		wg.Add(len(a.daemons))

		for _, d := range a.daemons {
			go func(_d DaemonFunc) {
				defer wg.Done()

				funcName := getFuncName(_d)

				defer func() {
					if r := recover(); r != nil {
						cErr <- fmt.Errorf("%s() panic:%s", funcName, r)
					}
				}()

				log.Tracef("  %s() ... running", funcName)
				if err := _d(ctx); err != nil {
					cErr <- fmt.Errorf("%s():%s", funcName, err)
					return
				}
				log.Tracef("  %s() ... done", funcName)
			}(d)
		}

		wg.Wait()

		close(cDone)
	}()

	var err error
	var isCanceled = false

__daemon_loop:
	for {
		var closeTimer <-chan time.Time
		if isCanceled {
			closeTimer = time.After(a.daemonForceCloseTimeout)
		} else {
			closeTimer = nil
		}

		select {
		case err = <-cErr:
			log.WithError(err).Errorf("!!Daemon err, exit in %s ...", a.daemonForceCloseTimeout.String())
			cancel()
			isCanceled = true
			cErr = nil // set cErr to nil to ignore other daemon fail
		case <-closeTimer:
			log.Infof("!!Daemon exit after %s", a.daemonForceCloseTimeout.String())
			break __daemon_loop
		case <-cDone:
			break __daemon_loop
		}
	}
	return err
}

// Run run bshark app, it should be called at last
func (a *Application) Run() {
	var err error
	log.Infof("Application %s starting...", a.name)

	if err = a.runInitStages(); err != nil {
		a.runCleanStage()
		log.WithError(err).Panic("Application fail to init!")
	}

	log.Infof("All init stage done, starting daemons...")

	if err = a.runDaemons(); err != nil {
		a.runCleanStage()
		log.WithError(err).Panic("Application fail to run daemon!")
	}
	log.Infof("App %s done", a.name)
}
