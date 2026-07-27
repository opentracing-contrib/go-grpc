module github.com/opentracing-contrib/go-grpc/test

go 1.25.0

replace github.com/opentracing-contrib/go-grpc => ../

require (
	github.com/golang/protobuf v1.5.4
	github.com/opentracing-contrib/go-grpc v0.0.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/stretchr/testify v1.11.1
	google.golang.org/grpc v1.82.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260414002931-afd174a4e478 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
