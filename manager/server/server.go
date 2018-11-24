package server

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/logp"

	"github.com/pkg/errors"

	"github.com/catrossim/monbeat/manager/pb"
	"github.com/golang/protobuf/ptypes/empty"
)

type RemoteServer struct {
	WorkDir string
	logger  *logp.Logger
}

func NewServer(workdir string) (*RemoteServer, error) {
	return &RemoteServer{
		WorkDir: workdir,
		logger:  logp.NewLogger("manager server"),
	}, nil
}

func (bs *RemoteServer) Ping(context.Context, *empty.Empty) (*pb.Response, error) {
	return &pb.Response{
		Code:      pb.ResultCode_SUCCESS,
		Result:    "",
		Timestamp: time.Now().Unix(),
	}, nil
}

func (bs *RemoteServer) Execute(stream pb.Remote_ExecuteServer) error {
	// receive file
	blob := []byte{}
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				if chunk != nil && len(chunk.Content) > 0 {
					blob = append(blob, chunk.Content...)
				}
				log.Printf("successfully receive files, length: %d", len(blob))
				break
			}
			err = errors.Wrapf(err, "failed to read chunk of file")
			return err
		}
		blob = append(blob, chunk.Content...)
	}
	// write file

	path := fmt.Sprintf("%s/%x.sh", bs.WorkDir, md5.Sum(blob))
	err := ioutil.WriteFile(path, blob, 0644)
	if err != nil {
		err = errors.Wrapf(err, "failed to write file %s", path)
		return err
	}
	// execute file
	r, err := execCmd(fmt.Sprintf("bash %s", path))
	if err != nil {
		err = errors.Wrapf(err, "failed to execute file %s", path)
		return err
	}
	err = stream.SendAndClose(&pb.Response{
		Code:      pb.ResultCode_SUCCESS,
		Result:    fmt.Sprintf("%s\n", string(r)),
		Timestamp: time.Now().Unix(),
	})
	bs.logger.Infof("successfully execute file %s", path)
	return nil
}

func execCmd(command string) ([]byte, error) {
	tokens := strings.Split(command, " ")
	cmd := exec.Command(tokens[0], tokens[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// func main() {
// 	lis, err := net.Listen("tcp", ":30112")
// 	if err != nil {
// 		log.Fatalf("failed to listen: %v", err)
// 	}
// 	s := grpc.NewServer()
// 	pb.RegisterRemoteServer(s, &RemoteServer{})
// 	// Register reflection service on gRPC server.
// 	reflection.Register(s)
// 	if err := s.Serve(lis); err != nil {
// 		log.Fatalf("failed to serve: %v", err)
// 	}
// }
