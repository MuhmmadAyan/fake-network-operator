package deviceplugin

import (
	"context"
	"net"
	"os"
	"path/filepath"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// Register connects to kubelet's registration gRPC and registers the device plugin.
func Register(socketPath string, resourceName string, pluginSocketPath string) error {
	conn, err := grpc.Dial(
		socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return net.Dial("unix", addr)
		}),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	req := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     filepath.Base(pluginSocketPath),
		ResourceName: resourceName,
	}

	_, err = client.Register(context.Background(), req)
	return err
}

// Serve creates a gRPC server, registers the DevicePluginServer, and listens on the unix socket.
func Serve(server pluginapi.DevicePluginServer, socketPath string) error {
	_ = os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	pluginapi.RegisterDevicePluginServer(grpcServer, server)

	return grpcServer.Serve(listener)
}
