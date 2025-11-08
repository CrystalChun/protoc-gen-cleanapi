/*
Copyright (c) 2026 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
the License. You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
*/

package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/jhernand/protoc-gen-cleanapi/internal/generator"
)

// Config contains the configuration loaded from environment variables. Note that these settings are only used for
// debugging purposes, refrain from using them in production.
type Config struct {
	// ReadRequest is the file where the codge generator request should be read from for debugging purposes. If
	// not defined or empty, the request will not be read from a file, which is the default behavior.
	ReadRequest string `envconfig:"READ_REQUEST"`

	// WriteRequest is the file where the codge generator request should be written for debugging purposes. If
	// not defined or empty, the request will not be written.
	WriteRequest string `envconfig:"WRITE_REQUEST"`

	// WriteResponse is the file where the code generator response should be written for debugging purposes.
	// not defined or empty, the response will not be written.
	WriteResponse string `envconfig:"WRITE_RESPONSE"`

	// LogFile is the path to the log file for debugging purposes. Special values 'stdout' and 'stderr' will write
	// to standard output or standard error respectively. Note that writing to standard output will interfere with
	// protoc, so only use it when running the plugin directly for debugging purposes.
	LogFile string `envconfig:"LOG_FILE" default:"protoc-gen-cleanapi.log"`

	// LogLevel is the logging level for debugging purposes. Valid values are 'debug', 'info', 'warn', 'error'.
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	// WaitFile is the path to a file that will be used to wait for the debugger to attach. When this is set the
	// program will stop till that file exists. The user can then attach the debuger and create the file to
	// continue.
	//
	// The file is automatically removed so that next time the program will wait again.
	WaitFile string `envconfig:"WAIT_FILE"`
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration from environment variables:
	var config Config
	err := envconfig.Process("PROTOC_GEN_CLEANAPI", &config)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// This is intended to debugging the plugin. When the 'PROTOC_GEN_CLEANAPI_WAIT_FILE' environment variable is
	// set we will wait till that file exists before continuing. That gives the user time to attch the debugger and
	// set breakpoints. After that the user can create the file and the program will continue automatically.
	//
	// The file is automatically removed so that next time the program will wait again.
	if config.WaitFile != "" {
		err = os.Remove(config.WaitFile)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove wait file '%s': %w", config.WaitFile, err)
		}
		defer func() {
			err = os.Remove(config.WaitFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to remove wait file '%s': %v\n", config.WaitFile, err)
			}
		}()
		fmt.Fprintf(os.Stderr, "Attach the debugger and then create file '%s' to continue ...\n", config.WaitFile)
		for {
			_, err := os.Stat(config.WaitFile)
			if err == nil {
				break
			}
			if !os.IsNotExist(err) {
				return fmt.Errorf("failed to check if file '%s' exists: %w", config.WaitFile, err)
			}
			time.Sleep(time.Second)
		}
	}

	// Parse the log level:
	var logLevel slog.Level
	err = logLevel.UnmarshalText([]byte(config.LogLevel))
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	// Prepare the log file:
	var logWriter io.Writer
	switch {
	case strings.EqualFold(config.LogFile, "stdout"):
		logWriter = os.Stdout
	case strings.EqualFold(config.LogFile, "stderr"):
		logWriter = os.Stderr
	case config.LogFile != "":
		var logFile *os.File
		logFile, err = os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file '%s': %w", config.LogFile, err)
		}
		logWriter = logFile
		defer logFile.Close()
	default:
		logWriter = io.Discard
	}
	logHandler := slog.NewJSONHandler(logWriter, &slog.HandlerOptions{
		Level: logLevel,
	})
	logger := slog.New(logHandler)

	// Log the command line arguments:
	logger.Debug(
		"Command line arguments",
		slog.Any("args", os.Args),
	)

	// Read the request, either from a file (for debugging purposes) or from the standard input stream:
	var input []byte
	if config.ReadRequest != "" {
		input, err = os.ReadFile(config.ReadRequest)
		if err != nil {
			return fmt.Errorf("failed to read request from file '%s': %w", config.ReadRequest, err)
		}
	} else {
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read request from standard input: %w", err)
		}
	}

	// Write the request to a file if configured:
	if config.WriteRequest != "" {
		err = os.WriteFile(config.WriteRequest, input, 0644)
		if err != nil {
			logger.Warn(
				"Failed to dump request",
				slog.String("file", config.WriteRequest),
				slog.Any("error", err),
			)
		} else {
			logger.Info(
				"Wrote request to file",
				slog.String("file", config.WriteRequest),
			)
		}
	}

	// Unmarshal the request:
	var request pluginpb.CodeGeneratorRequest
	err = proto.Unmarshal(input, &request)
	if err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}

	// Process the request:
	g, err := generator.New().
		SetLogger(logger).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build generator: %w", err)
	}
	response, err := g.Generate(&request)
	if err != nil {
		return fmt.Errorf("failed to generate: %w", err)
	}

	// Marshal the response:
	output, err := proto.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Write the response to a file if configured:
	if config.WriteResponse != "" {
		err = os.WriteFile(config.WriteResponse, output, 0644)
		if err != nil {
			logger.Warn(
				"Failed to write response",
				slog.String("file", config.WriteResponse),
				slog.Any("error", err),
			)
		} else {
			logger.Info(
				"Wrote response to file",
				slog.String("file", config.WriteResponse),
			)
		}
	}

	// Write the response to the standard output stream, which is where protoc expects it:
	_, err = os.Stdout.Write(output)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}
