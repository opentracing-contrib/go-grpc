module github.com/opentracing-contrib/go-grpc/test

go 1.22.7

replace github.com/opentracing-contrib/go-grpc => ../

require (
	github.com/golang/protobuf v1.5.4
	github.com/opentracing-contrib/go-grpc v0.0.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/stretchr/testify v1.10.0
	google.golang.org/grpc v1.70.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241202173237-19429a94021a // indirect
	google.golang.org/protobuf v1.35.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
