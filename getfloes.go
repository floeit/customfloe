package customfloe

import (
	"floe/tasks"
	triggers "floe/triggers"
	f "floe/workflow/flow"
	"time"
)

func getLocalFlows(p *f.Project) {
	p.AddTriggerFlow(p.MakeTriggerLauncher("triggered flow", FirstFlowFunc))

	spaceFlow := &workSpaceFlow{
		repo:   "danmux",
		branch: "master",
	}

	spaceFlow.Init("couch riak etcd")
	spaceFlowLauncher := f.MakeFlowLauncher(spaceFlow, 1, nil, nil)
	p.AddFlow(spaceFlowLauncher)

	localWorkspaceFlow := &localWorkspaceFlow{
		repo:   "danmux",
		branch: "development",
	}

	localWorkspaceFlow.Init("local workspace")
	// p.AddFlow(f.MakeFlowLauncher(localWorkspaceFlow, 1, spaceFlowLauncher, triggerFlowLauncher))
	p.AddFlow(f.MakeFlowLauncher(localWorkspaceFlow, 1, spaceFlowLauncher, nil))
}

func FirstFlowFunc(threadId int) *f.Workflow {
	w := f.MakeWorkflow()

	t1 := w.MakeTriggerNode("wait 14", triggers.MakeDelayTrigger(14*time.Second))
	t2 := w.MakeTriggerNode("wait 12", triggers.MakeDelayTrigger(12*time.Second))

	hip_start := w.MakeTaskNode("ping hipchat", tasks.MakeDelayTask(5*time.Second))

	co := w.MakeTaskNode("git checkout", tasks.MakeDelayTask(15*time.Second))

	last := w.MakeTaskNode("finish", tasks.MakeDelayTask(5*time.Second))

	tpush := w.MakeTriggerNode("push floeit pages", triggers.MakeGitPushTrigger("git@github.com:floeit/floeit.github.io.git", "", 10))

	hip_start.AddNext(0, co)
	co.AddNext(0, last)

	t1.AddNext(0, co)
	t2.AddNext(0, hip_start)

	tpush.AddNext(0, co)

	w.SetEnd(last)

	return w
}

func GetFlows(env string) *f.Project {

	p := f.MakeProject("V3")

	if env == "local" {
		getLocalFlows(p)
	}

	return p
}
