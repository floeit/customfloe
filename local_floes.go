package customfloe

import (
	"floe/tasks"
	f "floe/workflow/flow"
	"time"
)

// checkout workspace and build it
func mBuildWorkspace(w *f.Workflow, start *f.TaskNode, repo, branch string) (begin *f.TaskNode, end *f.TaskNode) {

	updateWs := w.MakeTaskNode("update workspace", tasks.MakeDelayTask(1*time.Second))

	update := w.MakeTaskNode("update", tasks.MakeDelayTask(1*time.Second))

	clean := w.MakeTaskNode("clean", tasks.MakeDelayTask(1*time.Second))

	clone := w.MakeTaskNode("clone workspace", tasks.MakeExecTask("git", "clone --progress git@github.com:floeit/floeit.github.io.git", ""))

	// clone := w.MakeTaskNode("clone workspace", tasks.MakeDelayTask(1*time.Second))

	create := w.MakeTaskNode("create repos", tasks.MakeDelayTask(1*time.Second))

	switchBranch := w.MakeTaskNode("switch", tasks.MakeDelayTask(1*time.Second))

	build := w.MakeTaskNode("build", tasks.MakeDelayTask(1*time.Second))

	if start != nil {
		start.AddNext(0, updateWs)
	}

	updateWs.AddNext(1, update)
	updateWs.AddNext(0, clean)
	clean.AddNext(0, clone)
	clean.AddNext(1, clone) // if clean failed it was already deleted so carry on regardless
	clone.AddNext(0, create)
	create.AddNext(0, switchBranch)

	switchBranch.AddNext(0, build)
	update.AddNext(0, build)
	update.AddNext(1, clean)

	return updateWs, build

}

func mPlaybook(w *f.Workflow, start *f.TaskNode, repo, branch string) (begin *f.TaskNode, end *f.TaskNode) {

	updatePb := w.MakeTaskNode("update play", tasks.MakeDelayTask(1*time.Second))

	clean := w.MakeTaskNode("clean play", tasks.MakeDelayTask(1*time.Second))

	clone := w.MakeTaskNode("clone play", tasks.MakeDelayTask(1*time.Second))

	pause := w.MakeTaskNode("play done", tasks.MakeDelayTask(0))

	if start != nil {
		start.AddNext(0, updatePb)
	}

	updatePb.AddNext(1, clean)
	updatePb.AddNext(0, pause)
	clean.AddNext(0, clone)
	clone.AddNext(0, pause)

	begin = updatePb
	end = pause
	return
}

func mWorkspace(w *f.Workflow, start *f.TaskNode, repo, branch string) (begin *f.TaskNode, end *f.TaskNode) {

	updatePb := w.MakeTaskNode("update wrk ", tasks.MakeDelayTask(1*time.Second))

	clean := w.MakeTaskNode("clean wrk", tasks.MakeDelayTask(1*time.Second))

	clone := w.MakeTaskNode("clone wrk", tasks.MakeDelayTask(1*time.Second))

	pause := w.MakeTaskNode("wrk done", tasks.MakeDelayTask(0))

	if start != nil {
		start.AddNext(0, updatePb)
	}

	updatePb.AddNext(1, clean)
	updatePb.AddNext(0, pause)
	clean.AddNext(0, clone)
	clone.AddNext(0, pause)

	begin = updatePb
	end = pause
	return
}

type workSpaceFlow struct {
	f.BaseLaunchable
	repo   string
	branch string
}

func (l *workSpaceFlow) GetProps() *f.Props {
	p := l.DefaultProps()
	(*p)[f.KEY_TIDY_DESK] = "keep" // this to not trash the workspace
	return p
}

func (l *workSpaceFlow) FlowFunc(threadId int) *f.Workflow {
	w := f.MakeWorkflow()

	start := w.MakeTaskNode("start c", tasks.MakeDelayTask(0))

	_, pEnd := mPlaybook(w, start, l.repo, l.branch)
	_, wEnd := mWorkspace(w, start, l.repo, l.branch)

	mn := w.MakeMergeNode("wait ws")

	mn.AddTrigger(pEnd)
	mn.AddTrigger(wEnd)

	vagrantUp := w.MakeTaskNode("vagrant up", tasks.MakeDelayTask(2))
	mn.SetNext(vagrantUp)

	w.SetStart(start)
	w.SetEnd(vagrantUp)

	return w
}

type localWorkspaceFlow struct {
	f.BaseLaunchable
	repo   string
	branch string
}

func (l *localWorkspaceFlow) GetProps() *f.Props {
	p := l.DefaultProps()
	(*p)[f.KEY_TIDY_DESK] = "keep" // this to not trash the workspace
	return p
}

func (l *localWorkspaceFlow) FlowFunc(threadId int) *f.Workflow {

	w := f.MakeWorkflow()

	updateWs, build := mBuildWorkspace(w, nil, l.repo, l.branch)
	// _, build := mBuildWorkspace(w, nil, l.repo, l.branch)

	kill := w.MakeTaskNode("killprocs", tasks.MakeDelayTask(1*time.Second))

	build.AddNext(0, kill)

	unit := w.MakeTaskNode("unit", tasks.MakeDelayTask(1*time.Second))

	launch := w.MakeTaskNode("launch", tasks.MakeDelayTask(1*time.Second))

	kill.AddNext(0, unit)
	kill.AddNext(1, unit) // may well fail if no m-be processes are running

	unit.AddNext(0, launch)

	pause := w.MakeTaskNode("settle", tasks.MakeDelayTask(2*time.Second))

	// export GOPATH=${PWD}; go install ./src/...
	buildConTest := w.MakeTaskNode("build con test", tasks.MakeDelayTask(1*time.Second))
	runConTest := w.MakeTaskNode("run con test", tasks.MakeDelayTask(1*time.Second))

	buildIdTest := w.MakeTaskNode("build id test", tasks.MakeDelayTask(1*time.Second))
	runIdTest := w.MakeTaskNode("run id test", tasks.MakeDelayTask(1*time.Second))

	buildApiTest := w.MakeTaskNode("build api test", tasks.MakeDelayTask(1*time.Second))
	runApiTest := w.MakeTaskNode("run api test", tasks.MakeDelayTask(1*time.Second))

	launch.AddNext(0, pause)
	pause.AddNext(0, buildConTest)
	pause.AddNext(0, buildIdTest)
	pause.AddNext(0, buildApiTest)

	buildConTest.AddNext(0, runConTest)
	buildIdTest.AddNext(0, runIdTest)
	buildApiTest.AddNext(0, runApiTest)

	mn := w.MakeMergeNode("wait tests")

	mn.AddTrigger(runConTest)
	mn.AddTrigger(runIdTest)
	mn.AddTrigger(runApiTest)

	done := w.MakeTaskNode("done", tasks.MakeDelayTask(0))

	mn.SetNext(done)

	w.SetStart(updateWs)
	// w.SetStart(kill)
	w.SetEnd(done)

	return w
}

type localWorkspaceTest struct {
	f.BaseLaunchable
	repo   string
	branch string
}

func (l *localWorkspaceTest) GetProps() *f.Props {
	p := l.DefaultProps()
	(*p)[f.KEY_TIDY_DESK] = "keep" // this to not trash the workspace
	return p
}

func (l *localWorkspaceTest) FlowFunc(threadId int) *f.Workflow {

	w := f.MakeWorkflow()

	updateWs := w.MakeTaskNode("update workspace", tasks.MakeDelayTask(1*time.Second))

	w.SetStart(updateWs)

	return w
}
