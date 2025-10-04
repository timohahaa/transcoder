package composer

import (
	"context"
	"time"

	retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	pb "github.com/timohahaa/transcoder/proto/composer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	skipTask = "SKIP_TASK"
	noTasks  = "NO_TASKS"
)

type (
	Task              = pb.Task
	GetTaskRequest    = pb.GetTaskRequest
	FinishTaskRequest = pb.FinishTaskRequest
	UpdateProgressReq = pb.UpdateProgressReq
)

func NewClient(addrs []string) (*Client, error) {
	register(addrs)

	var (
		opts = []retry.CallOption{
			retry.WithMax(3),
			retry.WithBackoff(retry.BackoffExponential(100 * time.Millisecond)),
		}
	)

	conn, err := grpc.NewClient("transcoder:///composer", []grpc.DialOption{
		grpc.WithUnaryInterceptor(retry.UnaryClientInterceptor(opts...)),
		grpc.WithStreamInterceptor(retry.StreamClientInterceptor(opts...)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [ { "round_robin": {} } ]}`),
	}...)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: pb.NewComposerClient(conn),
		conn:   conn,
	}, nil
}

type Client struct {
	client pb.ComposerClient
	conn   *grpc.ClientConn
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) GetTask(ctx context.Context, req *GetTaskRequest) (*Task, error) {
	return c.client.GetTask(ctx, req)
}

func (c *Client) FinishTask(ctx context.Context, req *FinishTaskRequest) error {
	_, err := c.client.FinishTask(ctx, req)
	return err
}

func (c *Client) UpdateProgress(ctx context.Context, req *UpdateProgressReq) error {
	_, err := c.client.UpdateProgress(ctx, req)
	return err
}

func IsSkipErr(err error) bool {
	if e, ok := status.FromError(err); ok {
		return e.Message() == noTasks
	}
	return false
}

func IsNoTasksErr(err error) bool {
	if e, ok := status.FromError(err); ok {
		return e.Message() == skipTask
	}
	return false
}

func IsUnavailableErr(err error) bool {
	if e, ok := status.FromError(err); ok {
		return e.Code() == codes.Unavailable
	}
	return false
}
