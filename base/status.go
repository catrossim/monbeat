package base

type Scode uint32

const (
	StatusEnabled Scode = 1 << iota
	StatusDisabled
	StatusStandby
	StatusRunning
	StatusStopped
)
