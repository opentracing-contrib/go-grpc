module github.com/opentracing-contrib/go-grpc/test

go 1.25.0

replace github.com/opentracing-contrib/go-grpc => ../

require (
	github.com/golang/protobuf v1.5.4
	github.com/opentracing-contrib/go-grpc v0.0.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/stretchr/testify v1.12.1
	google.golang.org/grpc v1.84.0
)

require (
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260706201446-f0a921348800 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
