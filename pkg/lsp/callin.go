package lspcore

type CallStackEntry struct {
	name string
}

type CallStack struct {
	data []CallStackEntry
}

func NewCallStack() *CallStack {
	ret := CallStack{}
	return &ret
}

type LspCallInRecord struct {
	name string
	data []CallStack
}
