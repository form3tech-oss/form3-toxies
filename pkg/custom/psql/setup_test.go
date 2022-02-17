package psql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Shopify/toxiproxy/v2"
	"github.com/Shopify/toxiproxy/v2/toxics"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

const postgresContainerName = "toxiproxy_postgres"
const postgresPort = 5432

var toxiProxyPort int

func TestMain(m *testing.M) {
	toxics.Register("psql", new(PostgresToxic))

	server := toxiproxy.NewServer()
	proxyPort, err := getFreePort()
	if err != nil {
		panic(err)
	}

	go server.Listen("localhost", strconv.Itoa(proxyPort))

	psqlPort := runPostgresContainer()
	waitForPostgres(psqlPort)

	toxiProxyPort = proxyPort
	m.Run()
}

func waitForPostgres(psqlPort int) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"localhost", psqlPort, "postgres", "postgres", "postgres")

	for i := 0; i < 30; i++ {
		db, err := sql.Open("postgres", psqlInfo)
		if err == nil {
			err = db.Ping()
			db.Close()
			if err == nil {
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func getHostPort(container types.Container, containerPort int) int {
	for _, p := range container.Ports {
		if p.PrivatePort == uint16(containerPort) {
			return int(p.PublicPort)
		}
	}
	return -1
}

func getContainerHostPort(containerName string, containerPort int) int {
	docker, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv)
	if err != nil {
		panic(err)
	}

	containers, err := docker.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}
	for _, c := range containers {
		if strings.HasSuffix(c.Names[0], containerName) {
			return getHostPort(c, containerPort)
		}
	}
	return -1
}

func runPostgresContainer() int {
	docker, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv)
	if err != nil {
		panic(err)
	}

	if port := getContainerHostPort(postgresContainerName, postgresPort); port > 0 {
		return port
	}

	port, err := getFreePort()
	if err != nil {
		panic(err)
	}

	hostBinding := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: strconv.Itoa(port),
	}
	containerPort, err := nat.NewPort("tcp", strconv.Itoa(postgresPort))
	if err != nil {
		panic("Unable to get the port")
	}

	cont, err := docker.ContainerCreate(
		context.Background(),
		&container.Config{
			Image: "postgres",
			Env: []string{
				"POSTGRES_PASSWORD=postgres",
			},
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{containerPort: []nat.PortBinding{hostBinding}},
		},
		nil,
		nil,
		postgresContainerName)

	if err != nil {
		panic(err)
	}

	err = docker.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}

	return port
}
