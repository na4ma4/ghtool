package runnerlist

import (
	"sync"

	"github.com/google/go-github/v70/github"
)

type Runners struct {
	lock    sync.Mutex
	list    map[int64]*github.Runner
	fresh   map[int64]bool
	outChan chan *github.Runner
}

func NewRunners() *Runners {
	return &Runners{
		list:    make(map[int64]*github.Runner),
		fresh:   make(map[int64]bool),
		outChan: make(chan *github.Runner),
	}
}

func (r *Runners) FreshnessReset() {
	r.lock.Lock()
	defer r.lock.Unlock()

	for idx := range r.fresh {
		r.fresh[idx] = false
	}
}

func (r *Runners) Add(runner *github.Runner) bool {
	if runner.ID != nil {
		r.fresh[runner.GetID()] = true
		if _, ok := r.list[runner.GetID()]; ok {
			return r.Update(runner)
		}
	}

	r.list[runner.GetID()] = runner
	r.outChan <- runner

	return true
}

func (r *Runners) PushUnfresh() {
	r.lock.Lock()
	defer r.lock.Unlock()

	for idx := range r.fresh {
		if !r.fresh[idx] {
			if _, ok := r.list[idx]; ok {
				shutdown := "shutdown"
				r.list[idx].Status = &shutdown
				r.outChan <- r.list[idx]
				delete(r.list, idx)
				delete(r.fresh, idx)
			}
		}
	}
}

func (r *Runners) Update(runner *github.Runner) bool {
	if v, ok := r.list[runner.GetID()]; ok {
		if compareRunners(v, runner) {
			return false
		}

		r.list[runner.GetID()] = runner
		r.outChan <- runner

		return true
	}

	return r.Add(runner)
}

func (r *Runners) Close() {
	close(r.outChan)
}

func (r *Runners) Channel() chan *github.Runner {
	return r.outChan
}

//nolint:varnamelen // a/b makes sense for comparisons.
func compareRunners(a, b *github.Runner) bool {
	if a.GetID() != b.GetID() {
		return false
	}

	if a.GetName() != b.GetName() {
		return false
	}

	if a.GetBusy() != b.GetBusy() {
		return false
	}

	if a.GetOS() != b.GetOS() {
		return false
	}

	if a.GetStatus() != b.GetStatus() {
		return false
	}

	if len(a.Labels) != len(b.Labels) {
		return false
	}

	if !compareLabels(a.Labels, b.Labels) {
		return false
	}

	if !compareLabels(b.Labels, a.Labels) {
		return false
	}

	return true
}

func compareLabels(a, b []*github.RunnerLabels) bool {
	for _, al := range a {
		found := false

		for _, bl := range b {
			if *al.ID == *bl.ID {
				found = true

				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}
