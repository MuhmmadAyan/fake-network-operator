package main

import (
	"context"
	"flag"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	hostPath string
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	flag.StringVar(&hostPath, "host-path", "/opt/fake-network-operator/bin/", "Path on the host to inject binaries.")
}

func main() {
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog.Info("starting cli-injector", "hostPath", hostPath)

	if err := os.MkdirAll(hostPath, 0755); err != nil {
		setupLog.Error(err, "failed to create host path")
		os.Exit(1)
	}

	binPath := filepath.Join(hostPath, "fake-cli")
	
	sourcePath := "/usr/local/bin/fake-cli"
	if err := copyFile(sourcePath, binPath); err != nil {
		setupLog.Error(err, "failed to copy fake-cli")
	}
	
	os.Chmod(binPath, 0755)

	commands := []string{"ibstat", "ibv_devinfo", "ibv_devices", "rdma"}
	for _, cmd := range commands {
		target := filepath.Join(hostPath, cmd)
		os.Remove(target)
		if err := os.Symlink("fake-cli", target); err != nil {
			setupLog.Error(err, "failed to create symlink", "cmd", cmd)
		}
	}

	setupLog.Info("injection complete, sleeping forever")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	<-ctx.Done()
	setupLog.Info("shutting down cli-injector")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Sync()
}
