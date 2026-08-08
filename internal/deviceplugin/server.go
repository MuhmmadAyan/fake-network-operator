package deviceplugin

import (
	"context"
	"fmt"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// Server is a reusable gRPC device plugin server.
type Server struct {
	resourceName string
	deviceIDs    []string
	envVarPrefix string
}

// NewServer creates a new Server instance.
func NewServer(resourceName string, deviceIDs []string, envVarPrefix string) *Server {
	return &Server{
		resourceName: resourceName,
		deviceIDs:    deviceIDs,
		envVarPrefix: envVarPrefix,
	}
}

// ListAndWatch sends devices and blocks.
func (s *Server) ListAndWatch(empty *pluginapi.Empty, stream pluginapi.DevicePlugin_ListAndWatchServer) error {
	devices := make([]*pluginapi.Device, len(s.deviceIDs))
	for i, id := range s.deviceIDs {
		devices[i] = &pluginapi.Device{
			ID:     id,
			Health: pluginapi.Healthy,
		}
	}

	err := stream.Send(&pluginapi.ListAndWatchResponse{Devices: devices})
	if err != nil {
		return err
	}

	// Block forever
	<-stream.Context().Done()
	return nil
}

// Allocate returns ContainerAllocateResponse with env vars.
func (s *Server) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	resp := &pluginapi.AllocateResponse{}

	for _, req := range reqs.ContainerRequests {
		cresp := &pluginapi.ContainerAllocateResponse{
			Envs: make(map[string]string),
		}
		if s.envVarPrefix != "" {
			var ids string
			for i, id := range req.DevicesIDs {
				if i > 0 {
					ids += ","
				}
				ids += id
			}
			cresp.Envs[fmt.Sprintf("%s_DEVICES", s.envVarPrefix)] = ids
		}
		resp.ContainerResponses = append(resp.ContainerResponses, cresp)
	}

	return resp, nil
}

// GetDevicePluginOptions is a no-op.
func (s *Server) GetDevicePluginOptions(ctx context.Context, empty *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

// PreStartContainer is a no-op.
func (s *Server) PreStartContainer(ctx context.Context, req *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

// GetPreferredAllocation returns empty.
func (s *Server) GetPreferredAllocation(ctx context.Context, req *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return &pluginapi.PreferredAllocationResponse{}, nil
}
