module github.com/mingcheng/obsync.go

go 1.12

replace github.com/mingcheng/pidfile => ../pidfile

require (
	github.com/mingcheng/pidfile v0.0.0-00010101000000-000000000000
	golang.org/x/sys v0.0.0-20190610200419-93c9922d18ae // indirect
)
