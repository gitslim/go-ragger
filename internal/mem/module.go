package mem

import "go.uber.org/fx"

var ModuleMemory = fx.Module("memory",
	fx.Provide(NewSimpleMemory))
