// +build e2e

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestEnvironment holds the test environment with all running services.
type TestEnvironment struct {
	// Service addresses
	DeviceManagerAddr      string
	UserServiceAddr        string
	APIGatewayAddr         string
	TelemetryCollectorAddr string
	MQTTBrokerAddr         string

	// Running processes
	deviceManagerCmd      *exec.Cmd
	userServiceCmd        *exec.Cmd
	apiGatewayCmd         *exec.Cmd
	telemetryCollectorCmd *exec.Cmd

	// Output buffers for debugging
	deviceManagerLog      *bytes.Buffer
	userServiceLog        *bytes.Buffer
	apiGatewayLog         *bytes.Buffer
	telemetryCollectorLog *bytes.Buffer

	// Cleanup function
	cleanup func()
}

// SetupE2EEnvironment starts all services and returns a configured test environment.
func SetupE2EEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()

	// Get project root
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	// Use different ports for E2E tests to avoid conflicts
	// Note: MQTT broker runs via docker-compose on standard port 1883
	env := &TestEnvironment{
		DeviceManagerAddr:      "localhost:18081",
		UserServiceAddr:        "localhost:18083",
		APIGatewayAddr:         "localhost:18080",
		TelemetryCollectorAddr: "localhost:18084",
		MQTTBrokerAddr:         "localhost:1883",
		deviceManagerLog:       &bytes.Buffer{},
		userServiceLog:         &bytes.Buffer{},
		apiGatewayLog:          &bytes.Buffer{},
		telemetryCollectorLog:  &bytes.Buffer{},
	}

	// Clean database before starting
	cleanDatabase(t)

	// Build all services
	t.Log("Building services...")
	buildServices(t, projectRoot)

	// Start Device Manager
	t.Log("Starting Device Manager...")
	env.deviceManagerCmd = startDeviceManager(t, projectRoot, env)

	// Start User Service
	t.Log("Starting User Service...")
	env.userServiceCmd = startUserService(t, projectRoot, env)

	// Start API Gateway
	t.Log("Starting API Gateway...")
	env.apiGatewayCmd = startAPIGateway(t, projectRoot, env)

	// Start Telemetry Collector
	t.Log("Starting Telemetry Collector...")
	env.telemetryCollectorCmd = startTelemetryCollector(t, projectRoot, env)

	// Wait for all services to be ready
	t.Log("Waiting for services to be ready...")
	waitForServices(t, env)

	// Setup cleanup
	env.cleanup = func() {
		t.Log("Cleaning up services...")
		if env.telemetryCollectorCmd != nil && env.telemetryCollectorCmd.Process != nil {
			env.telemetryCollectorCmd.Process.Kill()
		}
		if env.apiGatewayCmd != nil && env.apiGatewayCmd.Process != nil {
			env.apiGatewayCmd.Process.Kill()
		}
		if env.userServiceCmd != nil && env.userServiceCmd.Process != nil {
			env.userServiceCmd.Process.Kill()
		}
		if env.deviceManagerCmd != nil && env.deviceManagerCmd.Process != nil {
			env.deviceManagerCmd.Process.Kill()
		}

		// Print logs if test failed
		if t.Failed() {
			t.Log("=== Device Manager Logs ===")
			t.Log(env.deviceManagerLog.String())
			t.Log("=== User Service Logs ===")
			t.Log(env.userServiceLog.String())
			t.Log("=== API Gateway Logs ===")
			t.Log(env.apiGatewayLog.String())
			t.Log("=== Telemetry Collector Logs ===")
			t.Log(env.telemetryCollectorLog.String())
		}
	}

	t.Cleanup(env.cleanup)

	t.Log("E2E environment ready!")
	return env
}

// buildServices builds all service binaries.
func buildServices(t *testing.T, projectRoot string) {
	t.Helper()

	services := []struct {
		name string
		path string
	}{
		{"device-manager", "services/device-manager"},
		{"user-service", "services/user-service"},
		{"api-gateway", "services/api-gateway"},
		{"telemetry-collector", "services/telemetry-collector"},
	}

	for _, svc := range services {
		cmd := exec.Command("go", "build", "-o", filepath.Join(projectRoot, "bin", svc.name), ".")
		cmd.Dir = filepath.Join(projectRoot, svc.path)
		cmd.Env = append(os.Environ(), "CGO_ENABLED=0")

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to build %s: %v\n%s", svc.name, err, output)
		}
		t.Logf("Built %s", svc.name)
	}
}

// startDeviceManager starts the Device Manager service.
func startDeviceManager(t *testing.T, projectRoot string, env *TestEnvironment) *exec.Cmd {
	t.Helper()

	cmd := exec.Command(filepath.Join(projectRoot, "bin", "device-manager"))
	cmd.Env = append(os.Environ(),
		"DEVICE_MANAGER_PORT=18081",
		"STORAGE_TYPE=postgres",
		"DB_HOST=localhost",
		"DB_PORT=5432",
		"DB_NAME=iot_platform",
		"DB_USER=iot_user",
		"DB_PASSWORD=iot_password",
		"DB_SSLMODE=disable",
	)

	// Capture output
	cmd.Stdout = io.MultiWriter(env.deviceManagerLog, testLogWriter{t, "device-manager"})
	cmd.Stderr = io.MultiWriter(env.deviceManagerLog, testLogWriter{t, "device-manager"})

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start Device Manager: %v", err)
	}

	return cmd
}

// startUserService starts the User Service.
func startUserService(t *testing.T, projectRoot string, env *TestEnvironment) *exec.Cmd {
	t.Helper()

	cmd := exec.Command(filepath.Join(projectRoot, "bin", "user-service"))
	cmd.Env = append(os.Environ(),
		"USER_SERVICE_PORT=18083",
		"STORAGE_TYPE=postgres",
		"DB_HOST=localhost",
		"DB_PORT=5432",
		"DB_NAME=iot_platform",
		"DB_USER=iot_user",
		"DB_PASSWORD=iot_password",
		"DB_SSLMODE=disable",
	)

	cmd.Stdout = io.MultiWriter(env.userServiceLog, testLogWriter{t, "user-service"})
	cmd.Stderr = io.MultiWriter(env.userServiceLog, testLogWriter{t, "user-service"})

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start User Service: %v", err)
	}

	return cmd
}

// startAPIGateway starts the API Gateway.
func startAPIGateway(t *testing.T, projectRoot string, env *TestEnvironment) *exec.Cmd {
	t.Helper()

	cmd := exec.Command(filepath.Join(projectRoot, "bin", "api-gateway"))
	cmd.Env = append(os.Environ(),
		"PORT=18080",
		"DEVICE_MANAGER_ADDR=localhost:18081",
		"USER_SERVICE_ADDR=localhost:18083",
		"TELEMETRY_SERVICE_ADDR=localhost:18084",
		"JWT_SECRET=e2e-test-secret-key-for-testing-only",
	)

	cmd.Stdout = io.MultiWriter(env.apiGatewayLog, testLogWriter{t, "api-gateway"})
	cmd.Stderr = io.MultiWriter(env.apiGatewayLog, testLogWriter{t, "api-gateway"})

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start API Gateway: %v", err)
	}

	return cmd
}

// startTelemetryCollector starts the Telemetry Collector service.
func startTelemetryCollector(t *testing.T, projectRoot string, env *TestEnvironment) *exec.Cmd {
	t.Helper()

	cmd := exec.Command(filepath.Join(projectRoot, "bin", "telemetry-collector"))
	cmd.Env = append(os.Environ(),
		"TELEMETRY_GRPC_PORT=18084",
		"MQTT_BROKER=tcp://localhost:1883",
		"MQTT_CLIENT_ID=telemetry-collector-e2e",
		"MQTT_TOPIC=devices/+/telemetry",
		"DB_HOST=localhost",
		"DB_PORT=5432",
		"DB_NAME=iot_platform",
		"DB_USER=iot_user",
		"DB_PASSWORD=iot_password",
		"DB_SSLMODE=disable",
	)

	cmd.Stdout = io.MultiWriter(env.telemetryCollectorLog, testLogWriter{t, "telemetry-collector"})
	cmd.Stderr = io.MultiWriter(env.telemetryCollectorLog, testLogWriter{t, "telemetry-collector"})

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start Telemetry Collector: %v", err)
	}

	return cmd
}

// waitForServices waits for all services to be ready by checking health endpoints.
func waitForServices(t *testing.T, env *TestEnvironment) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	services := []struct {
		name string
		url  string
	}{
		{"API Gateway", fmt.Sprintf("http://%s/health", env.APIGatewayAddr)},
	}

	for _, svc := range services {
		if err := waitForHTTP(ctx, svc.url); err != nil {
			t.Fatalf("%s not ready: %v", svc.name, err)
		}
		t.Logf("%s is ready", svc.name)
	}

	// Additional wait for gRPC services (they don't have HTTP health endpoints)
	// We just wait a bit for them to start
	time.Sleep(2 * time.Second)
	t.Log("gRPC services should be ready")
}

// waitForHTTP waits for an HTTP endpoint to become available.
func waitForHTTP(ctx context.Context, url string) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for %s", url)
		case <-ticker.C:
			resp, err := http.Get(url)
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				return nil
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
	}
}

// cleanDatabase cleans the test database before running tests.
func cleanDatabase(t *testing.T) {
	t.Helper()

	// Clean each table separately to handle missing tables gracefully
	// Order matters due to foreign key constraints - clean dependent tables first
	tables := []string{"device_telemetry_latest", "device_telemetry", "devices", "users"}

	for _, table := range tables {
		cmd := exec.Command("docker-compose", "exec", "-T", "postgres",
			"psql", "-U", "iot_user", "-d", "iot_platform",
			"-c", fmt.Sprintf("TRUNCATE TABLE %s CASCADE;", table),
		)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Warning: Failed to clean table %s: %v\nOutput: %s", table, err, string(output))
		} else {
			t.Logf("Cleaned table: %s\nOutput: %s", table, string(output))
		}
	}
}

// testLogWriter wraps test logger for real-time output.
type testLogWriter struct {
	t       *testing.T
	service string
}

func (w testLogWriter) Write(p []byte) (n int, err error) {
	// Only log if verbose mode or test fails
	if testing.Verbose() {
		w.t.Logf("[%s] %s", w.service, string(p))
	}
	return len(p), nil
}
