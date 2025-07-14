package mem

import "go.uber.org/fx"

// ModuleMemory provides a memory for messages
var ModuleMemory = fx.Module("memory",
	fx.Provide(NewSimpleMemory))
