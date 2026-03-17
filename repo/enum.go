package repo

type FrameFlags uint8

const (
	FlagNone FrameFlags = 0
	FlagGzip FrameFlags = 1 << 0
	FlagTLS  FrameFlags = 1 << 1
)
