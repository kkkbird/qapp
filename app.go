package bshark

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sync"
)

type InitFunc func(ctx context.Context) (context.Context, error)
type DaemonFunc func(ctx context.Context) error

type InitStage struct {
	name  string
	funcs []InitFunc
}

func getFuncName(f interface{}) string {
	fv := reflect.ValueOf(f)

	if fv.Kind() != reflect.Func {
		return ""
	}
	return runtime.FuncForPC(fv.Pointer()).Name()
}

func newInitStage(name string, funcs []InitFunc) *InitStage {
	return &InitStage{
		name:  name,
		funcs: funcs,
	}
}

type Application struct {
	logger     Logger
	name       string
	ctx        context.Context
	initStages []*InitStage
	daemons    []DaemonFunc
}

func New(name string, ctx context.Context) *Application {
	if ctx == nil {
		ctx = context.TODO()
	}

	return &Application{
		name:       name,
		ctx:        ctx,
		initStages: make([]*InitStage, 0),
		daemons:    make([]DaemonFunc, 0),
	}
}

func (a *Application) printf(format string, args ...interface{}) {
	if a.logger == nil {
		return
	}

	a.logger.Printf(format, args...)
}

func (a *Application) SetLogger(logger Logger) *Application {
	a.logger = logger
	return a
}

func (a *Application) AddInitStage(name string, funcs ...InitFunc) *Application {
	a.initStages = append(a.initStages, newInitStage(name, funcs))
	return a
}

func (a *Application) AddDaemons(funcs ...DaemonFunc) *Application {
	a.daemons = append(a.daemons, funcs...)
	return a
}

func (a *Application) Run() {
	var err error
	a.printf("App %s start", a.name)
	var wg sync.WaitGroup

	// run init stages
	for i, s := range a.initStages {
		a.printf("Init stage %d-%s", i+1, s.name)

		wg.Add(len(s.funcs))

		for _, fc := range s.funcs {
			go func(_fc InitFunc) {
				defer wg.Done()

				funcName := getFuncName(_fc)

				// TODO: add recover for call init func
				if a.ctx, err = _fc(a.ctx); err != nil {
					panic(fmt.Sprintf("%s() fail: %s", funcName, err))
				}

				a.printf("  %s() ... done", funcName)

			}(fc)
		}
		wg.Wait()
	}

	// run daemon funcs
	a.printf("All init stage done, start daemons")

	wg.Add(len(a.daemons))

	for _, d := range a.daemons {
		go func(_d DaemonFunc) {
			defer wg.Done()

			funcName := getFuncName(_d)

			// TODO: add recover for call daemon func
			a.printf("  %s() ... running", funcName)
			if err := _d(a.ctx); err != nil {
				panic(fmt.Sprintf("%s() fail: %s", funcName, err))
			}
			a.printf("  %s() ... done", funcName)
		}(d)
	}

	wg.Wait()

	a.printf("App %s done", a.name)

}
