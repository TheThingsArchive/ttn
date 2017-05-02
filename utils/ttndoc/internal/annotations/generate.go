// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

//go:generate protoc -I=$GOPATH/src/github.com/googleapis/googleapis --gogo_out=Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:. $GOPATH/src/github.com/googleapis/googleapis/google/api/annotations.proto $GOPATH/src/github.com/googleapis/googleapis/google/api/http.proto
//go:generate mv ./google.golang.org/genproto/googleapis/api/annotations/annotations.pb.go ./
//go:generate mv ./google.golang.org/genproto/googleapis/api/annotations/http.pb.go ./
//go:generate rm -rf ./google.golang.org

package annotations
