package services

import "neuroshell/internal/context"

// QueueService provides command queuing functionality for the state machine
type QueueService struct {
	context *context.NeuroContext
}

// NewQueueService creates a new queue service instance
func NewQueueService(ctx *context.NeuroContext) *QueueService {
	return &QueueService{
		context: ctx,
	}
}

// Name returns the service name for registry
func (qs *QueueService) Name() string {
	return "queue"
}

// Initialize initializes the queue service
func (qs *QueueService) Initialize() error {
	// Queue service is stateless and uses context directly
	return nil
}

// QueueCommand adds a single command to the execution queue
func (qs *QueueService) QueueCommand(command string) {
	qs.context.QueueCommand(command)
}

// QueueCommands adds multiple commands to the execution queue
func (qs *QueueService) QueueCommands(commands []string) {
	for _, cmd := range commands {
		qs.context.QueueCommand(cmd)
	}
}

// GetQueueSize returns the number of commands in the execution queue
func (qs *QueueService) GetQueueSize() int {
	return qs.context.GetQueueSize()
}

// ClearQueue removes all commands from the execution queue
func (qs *QueueService) ClearQueue() {
	qs.context.ClearQueue()
}

// DequeueCommand removes and returns the next command from the queue
func (qs *QueueService) DequeueCommand() (string, bool) {
	return qs.context.DequeueCommand()
}

// PeekQueue returns a copy of the execution queue without modifying it
func (qs *QueueService) PeekQueue() []string {
	return qs.context.PeekQueue()
}
