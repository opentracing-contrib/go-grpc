package interceptor_test

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	otgrpc "github.com/opentracing-contrib/go-grpc"
	testpb "github.com/opentracing-contrib/go-grpc/test/otgrpc_testing"
	"github.com/opentracing/opentracing-go/mocktracer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	streamLength = 5
)

type testServer struct{}

func (s *testServer) UnaryCall(_ context.Context, in *testpb.SimpleRequest) (*testpb.SimpleResponse, error) {
	return &testpb.SimpleResponse{Payload: in.Payload}, nil
}

func (s *testServer) StreamingOutputCall(in *testpb.SimpleRequest, stream testpb.TestService_StreamingOutputCallServer) error {
	for range streamLength {
		if err := stream.Send(&testpb.SimpleResponse{Payload: in.Payload}); err != nil {
			return err
		}
	}
	return nil
}

func (s *testServer) StreamingInputCall(stream testpb.TestService_StreamingInputCallServer) error {
	sum := int32(0)
	for {
		in, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		sum += in.Payload
	}
	return stream.SendAndClose(&testpb.SimpleResponse{Payload: sum})
}

func (s *testServer) StreamingBidirectionalCall(stream testpb.TestService_StreamingBidirectionalCallServer) error {
	for {
		in, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if err = stream.Send(&testpb.SimpleResponse{Payload: in.Payload}); err != nil {
			return err
		}
	}
}

type env struct {
	unaryClientInt  grpc.UnaryClientInterceptor
	streamClientInt grpc.StreamClientInterceptor
	unaryServerInt  grpc.UnaryServerInterceptor
	streamServerInt grpc.StreamServerInterceptor
}

type test struct {
	t   *testing.T
	e   env
	srv *grpc.Server
	cc  *grpc.ClientConn
	c   testpb.TestServiceClient
}

func newTest(t *testing.T, e env) *test {
	t.Helper()
	te := &test{
		t: t,
		e: e,
	}

	// Set up the server.
	sOpts := []grpc.ServerOption{}
	if e.unaryServerInt != nil {
		sOpts = append(sOpts, grpc.UnaryInterceptor(e.unaryServerInt))
	}
	if e.streamServerInt != nil {
		sOpts = append(sOpts, grpc.StreamInterceptor(e.streamServerInt))
	}
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		te.t.Fatalf("Failed to listen: %v", err)
	}
	te.srv = grpc.NewServer(sOpts...)
	testpb.RegisterTestServiceServer(te.srv, &testServer{})
	errChan := make(chan error, 1)
	go func() {
		errChan <- te.srv.Serve(lis)
	}()

	// Check for immediate server startup errors
	select {
	case errServe := <-errChan:
		if errServe != nil {
			te.t.Fatalf("Failed to serve: %v", errServe)
		}
	default:
	}

	// Set up a connection to the server.
	cOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if e.unaryClientInt != nil {
		cOpts = append(cOpts, grpc.WithUnaryInterceptor(e.unaryClientInt))
	}
	if e.streamClientInt != nil {
		cOpts = append(cOpts, grpc.WithStreamInterceptor(e.streamClientInt))
	}
	_, port, err := net.SplitHostPort(lis.Addr().String())
	if err != nil {
		te.t.Fatalf("Failed to parse listener address: %v", err)
	}
	srvAddr := "localhost:" + port
	te.cc, err = grpc.NewClient(srvAddr, cOpts...)
	if err != nil {
		te.t.Fatalf("Dial(%q) = %v", srvAddr, err)
	}
	te.c = testpb.NewTestServiceClient(te.cc)
	return te
}

func (te *test) tearDown() {
	te.cc.Close()
}

func assertChildParentSpans(t *testing.T, tracer *mocktracer.MockTracer) {
	t.Helper()
	spans := tracer.FinishedSpans()
	assert.Equal(t, 2, len(spans))
	if len(spans) != 2 {
		t.Fatalf("Incorrect span length")
	}
	parent := spans[1]
	child := spans[0]
	parentContext, ok := parent.Context().(mocktracer.MockSpanContext)
	if !ok {
		t.Fatalf("Failed to assert parent context as mocktracer.MockSpanContext")
	}
	assert.Equal(t, child.ParentID, parentContext.SpanID)
}

func TestUnaryOpenTracing(t *testing.T) {
	t.Parallel()
	tracer := mocktracer.New()
	e := env{
		unaryClientInt: otgrpc.OpenTracingClientInterceptor(tracer),
		unaryServerInt: otgrpc.OpenTracingServerInterceptor(tracer),
	}
	te := newTest(t, e)
	defer te.tearDown()

	payload := int32(0)
	resp, err := te.c.UnaryCall(context.Background(), &testpb.SimpleRequest{Payload: payload})
	if err != nil {
		t.Fatalf("Failed UnaryCall: %v", err)
	}
	assert.Equal(t, payload, resp.Payload)
	assertChildParentSpans(t, tracer)
}

func TestStreamingOutputCallOpenTracing(t *testing.T) {
	t.Parallel()
	tracer := mocktracer.New()
	e := env{
		streamClientInt: otgrpc.OpenTracingStreamClientInterceptor(tracer),
		streamServerInt: otgrpc.OpenTracingStreamServerInterceptor(tracer),
	}
	te := newTest(t, e)
	defer te.tearDown()

	payload := int32(0)
	stream, err := te.c.StreamingOutputCall(context.Background(), &testpb.SimpleRequest{Payload: payload})
	if err != nil {
		t.Fatalf("Failed StreamingOutputCall: %v", err)
	}
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("Failed StreamingOutputCall: %v", err)
		}
		assert.Equal(t, payload, resp.Payload)
	}
	assertChildParentSpans(t, tracer)
}

func TestStreamingInputCallOpenTracing(t *testing.T) {
	t.Parallel()
	tracer := mocktracer.New()
	e := env{
		streamClientInt: otgrpc.OpenTracingStreamClientInterceptor(tracer),
		streamServerInt: otgrpc.OpenTracingStreamServerInterceptor(tracer),
	}
	te := newTest(t, e)
	defer te.tearDown()

	payload := int32(1)
	stream, err := te.c.StreamingInputCall(context.Background())
	if err != nil {
		t.Fatalf("Failed StreamingInputCall: %v", err)
	}
	for range streamLength {
		if err = stream.Send(&testpb.SimpleRequest{Payload: payload}); err != nil {
			t.Fatalf("Failed StreamingInputCall: %v", err)
		}
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatalf("Failed StreamingInputCall: %v", err)
	}
	assert.Equal(t, streamLength*payload, resp.Payload)
	assertChildParentSpans(t, tracer)
}

func TestStreamingBidirectionalCallOpenTracing(t *testing.T) {
	t.Parallel()
	tracer := mocktracer.New()
	e := env{
		streamClientInt: otgrpc.OpenTracingStreamClientInterceptor(tracer),
		streamServerInt: otgrpc.OpenTracingStreamServerInterceptor(tracer),
	}
	te := newTest(t, e)
	defer te.tearDown()

	payload := int32(0)
	stream, err := te.c.StreamingBidirectionalCall(context.Background())
	if err != nil {
		t.Fatalf("Failed StreamingInputCall: %v", err)
	}
	errChan := make(chan error, 1)
	go func() {
		for range streamLength {
			if err := stream.Send(&testpb.SimpleRequest{Payload: payload}); err != nil {
				errChan <- err
				return
			}
		}
		if err := stream.CloseSend(); err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	if err := <-errChan; err != nil {
		t.Fatalf("Failed StreamingInputCall: %v", err)
	}
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("Failed StreamingOutputCall: %v", err)
		}
		assert.Equal(t, payload, resp.Payload)
	}
	assertChildParentSpans(t, tracer)
}

func TestStreamingContextCancellationOpenTracing(t *testing.T) {
	t.Parallel()
	tracer := mocktracer.New()
	e := env{
		streamClientInt: otgrpc.OpenTracingStreamClientInterceptor(tracer),
		streamServerInt: otgrpc.OpenTracingStreamServerInterceptor(tracer),
	}
	te := newTest(t, e)
	defer te.tearDown()

	payload := int32(0)
	ctx, cancel := context.WithCancel(context.Background())
	_, err := te.c.StreamingOutputCall(ctx, &testpb.SimpleRequest{Payload: payload})
	if err != nil {
		t.Fatalf("Failed StreamingOutputCall: %v", err)
	}
	cancel()
	time.Sleep(100 * time.Millisecond)
	spans := tracer.FinishedSpans()
	assert.Equal(t, 2, len(spans))
	if len(spans) != 2 {
		t.Fatalf("Incorrect span length")
	}
	parent := spans[0]
	child := spans[1]
	parentContext, ok := parent.Context().(mocktracer.MockSpanContext)
	if !ok {
		t.Fatalf("Failed to assert parent context as mocktracer.MockSpanContext")
	}
	assert.Equal(t, child.ParentID, parentContext.SpanID)
	errorTag, ok := parent.Tag("error").(bool)
	if !ok {
		t.Fatalf("Failed to assert error tag as bool")
	}
	assert.True(t, errorTag)
}
