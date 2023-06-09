package runnerlist

import (
	"github.com/google/go-github/v53/github"
)

type Runners struct {
	list    map[int64]*github.Runner
	outChan chan *github.Runner
}

func NewRunners() *Runners {
	return &Runners{
		list:    make(map[int64]*github.Runner),
		outChan: make(chan *github.Runner),
	}
}

func (r *Runners) Add(runner *github.Runner) bool {
	if runner.ID != nil {
		if _, ok := r.list[runner.GetID()]; ok {
			return r.Update(runner)
		}
	}

	r.list[runner.GetID()] = runner
	r.outChan <- runner

	return true
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
