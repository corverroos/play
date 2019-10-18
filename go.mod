module github.com/corverroos/play

go 1.12

replace github.com/corverroos/unsure => ../unsure

require (
	github.com/corverroos/unsure v0.0.0-00010101000000-000000000000
	github.com/golang/protobuf v1.3.2
	github.com/luno/fate v0.0.0-20190906093333-f60ec39889bc
	github.com/luno/jettison v0.0.0-20191014084106-b0501ece4f1c
	github.com/luno/reflex v0.0.0-20191010085905-159383ec8c22
	github.com/luno/shift v0.0.0-20190912102423-a69494119072
	golang.org/x/net v0.0.0-20191009170851-d66e71096ffb
	google.golang.org/grpc v1.24.0
)
