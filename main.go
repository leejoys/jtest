package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"contrib.go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/trace"
)

func main() {
	// Initialize OpenCensus tracer
	exporter, err := jaeger.NewExporter(jaeger.Options{
		CollectorEndpoint: "http://jaeger:14268/api/traces",
		Process: jaeger.Process{
			ServiceName: "test-trace",
			Tags: []jaeger.Tag{
				jaeger.StringTag("env", "test env"),
			},
		},
		BufferMaxCount: 100,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer exporter.Flush()

	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <url>")
		os.Exit(1)
	}
	url := os.Args[1]

	// Create a span
	ctx, span := trace.StartSpan(context.Background(), "simple-trace")
	defer span.End()

	// Create an HTTP request with tracing context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Println(err)
		span.SetStatus(trace.Status{
			Code:    trace.StatusCodeUnknown,
			Message: "test error" + err.Error(),
		})

	}

	// Send the HTTP request and get the response
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		span.SetStatus(trace.Status{
			Code:    trace.StatusCodeUnknown,
			Message: "test error" + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		span.SetStatus(trace.Status{
			Code:    trace.StatusCodeUnknown,
			Message: "test error" + err.Error(),
		})
	}

	fmt.Println(resp.Proto, resp.Status)
	for k, v := range resp.Header {
		fmt.Println(k+":", v[0])
	}
	fmt.Println()
	fmt.Println(string(body))
}
