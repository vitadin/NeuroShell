package services

import "neuroshell/internal/context"

// QueueService provides command queuing functionality for the state machine
type QueueService struct {
	initialized bool
}

// NewQueueService creates a new queue service instance
func NewQueueService() *QueueService {
	return &QueueService{
		initialized: false,
	}
}

// Name returns the service name for registry
func (qs *QueueService) Name() string {
	return "queue"
}

// Initialize initializes the queue service
func (qs *QueueService) Initialize() error {
	qs.initialized = true
	return nil
}

// QueueCommand adds a single command to the execution queue
func (qs *QueueService) QueueCommand(command string) {
	if !qs.initialized {
		return
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)
	ctx.QueueCommand(command)
}

// QueueCommands adds multiple commands to the execution queue
func (qs *QueueService) QueueCommands(commands []string) {
	if !qs.initialized {
		return
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)
	for _, cmd := range commands {
		ctx.QueueCommand(cmd)
	}
}

// GetQueueSize returns the number of commands in the execution queue
func (qs *QueueService) GetQueueSize() int {
	if !qs.initialized {
		return 0
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)
	return ctx.GetQueueSize()
}

// ClearQueue removes all commands from the execution queue
func (qs *QueueService) ClearQueue() {
	if !qs.initialized {
		return
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)
	ctx.ClearQueue()
}

// DequeueCommand removes and returns the next command from the queue
func (qs *QueueService) DequeueCommand() (string, bool) {
	if !qs.initialized {
		return "", false
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)
	return ctx.DequeueCommand()
}

// PeekQueue returns a copy of the execution queue without modifying it
func (qs *QueueService) PeekQueue() []string {
	if !qs.initialized {
		return []string{}
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)
	return ctx.PeekQueue()
}
