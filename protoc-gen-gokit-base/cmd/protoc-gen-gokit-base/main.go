package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	generator "github.com/TuneLab/gob/protoc-gen-gokit-base/generator"
	"github.com/gengo/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	_ "github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

// parseReq reads io.Reader r into memory and attempts to marshal
// that input into a protobuf plugin CodeGeneratorRequest
func parseReq(r io.Reader) (*plugin.CodeGeneratorRequest, error) {
	glog.V(1).Info("Parsing code generator request")
	input, err := ioutil.ReadAll(r)
	if err != nil {
		glog.Errorf("Failed to read code generator request: %v", err)
		return nil, err
	}
	req := new(plugin.CodeGeneratorRequest)
	if err = proto.Unmarshal(input, req); err != nil {
		glog.Errorf("Failed to unmarshal code generator request: %v", err)
		return nil, err
	}
	glog.V(1).Info("Parsed code generator request")
	return req, nil
}

var (
	response = string("")
)

// Leland Batey's log to os.Stderr
func logf(format string, args ...interface{}) {
	response += fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, format, args...)
}

func main() {
	defer glog.Flush()
	glog.V(1).Info("Processing code generator request")

	registry := descriptor.NewRegistry()
	request, err := parseReq(os.Stdin)
	if err != nil {
		glog.Fatal(err)
	}

	g := generator.New(registry)

	if err := registry.Load(request); err != nil {
		return
	}

	var targets []*descriptor.File
	for _, target := range request.FileToGenerate {
		logf("file to be processed: %v\n", target)
		f, err := registry.LookupFile(target)
		if err != nil {
			glog.Fatal(err)
		}
		targets = append(targets, f)
	}

	//logf("%v\n", len(targets))
	glog.V(1).Info("Building Output")

	// Get working directory, trim off GOPATH, add generate.
	// This should be the absolute path for the relative package dependencies
	wd, _ := os.Getwd()
	goPath := os.Getenv("GOPATH")
	logf("working directory:%s\n$GOPATH:%s\n", wd, goPath)
	importPath := strings.TrimPrefix(wd, goPath+"/src/")
	importPath = importPath + "/generate/"
	logf("%s\n", importPath)

	codeGenFiles, _ := g.GenerateResponseFiles(targets)

	output := &plugin.CodeGeneratorResponse{
		File: codeGenFiles,
	}

	buf, err := proto.Marshal(output)
	if err != nil {
		glog.Fatal(err)
	}

	if _, err := os.Stdout.Write(buf); err != nil {
		glog.Fatal(err)
	}
}
