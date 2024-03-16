package workflow

import (
	"context"
	"fmt"
	"github.com/viant/afs"
	"github.com/viant/endly"
	"github.com/viant/endly/model"
	"github.com/viant/endly/model/location"
	"github.com/viant/toolbox/data"
	"github.com/viant/toolbox/storage"
	"strings"
)

var processesKey = (*model.Processes)(nil)

func processes(context *endly.Context) *model.Processes {
	var result *model.Processes
	if !context.Contains(processesKey) {
		result = model.NewProcesses()
		_ = context.Put(processesKey, result)
	} else {
		context.GetInto(processesKey, &result)
	}
	return result
}

// Push push process to context
func Push(context *endly.Context, process *model.Process) {
	var processes = processes(context)
	if process.Source != nil {
		context.Source = process.Source
	}
	processes.Push(process)
}

// Remove push process to context
func Pop(context *endly.Context) *model.Process {
	var processes = processes(context)
	var process = processes.Pop()
	if process != nil && process.Source != nil {
		context.Source = process.Source
	}
	return process
}

// Last returns last process
func Last(context *endly.Context) *model.Process {
	var processes = processes(context)
	return processes.Last()
}

// LastWorkflow returns last workflow
func LastWorkflow(context *endly.Context) *model.Process {
	var processes = processes(context)
	return processes.LastWorkflow()
}

// FirstWorkflow returns last workflow
func FirstWorkflow(context *endly.Context) *model.Process {
	var processes = processes(context)
	return processes.FirstWorkflow()
}

// GetResource returns workflow resource
func GetResource(ctx context.Context, state data.Map, URL string) *location.Resource {
	for _, candidate := range getURLs(URL) {
		resource := location.NewResource(candidate)
		storageService, err := storage.NewServiceForURL(resource.URL, "")
		if err != nil {
			return nil
		}
		exists, _ := storageService.Exists(resource.URL)
		if exists {
			return resource
		}
	}
	if strings.Contains(URL, ":/") || strings.HasPrefix(URL, "/") {
		return nil
	}
	//Lookup shared workflow
	fs := afs.New()
	for _, candidate := range getURLs(URL) {
		if ok, _ := fs.Exists(ctx, candidate); ok {
			return location.NewResource(candidate)
		}
	}
	return nil
}

func getURLs(URL string) []string {
	selector := model.WorkflowSelector(URL)
	workflowName := selector.Name()
	workflowFilename := fmt.Sprintf("%v.csv", workflowName)
	dedicatedFolderURL := strings.Replace(URL, workflowFilename, fmt.Sprintf("%v/%v", workflowName, workflowFilename), 1)
	return []string{
		URL,
		dedicatedFolderURL,
	}
}

func isWorkflowRunAction(action *model.Action) bool {
	return action.Action == "run" && action.Service == ServiceID
}

func runWithoutSelfIfNeeded(process *model.Process, action *model.Action, state data.Map, handler func() error) error {
	if !isWorkflowRunAction(action) {
		return handler()
	}
	state.Delete(selfStateKey)
	defer state.Put(selfStateKey, process.State)
	return handler()
}
