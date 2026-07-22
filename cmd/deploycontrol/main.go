package main

import (
	"context"
	"deploy-platform-go/internal/dockerimage"
	"deploy-platform-go/internal/project"
	"encoding/json"
	"flag"
	"github.com/moby/moby/api/types/build"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/jsonstream"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/moby/moby/client/pkg/jsonmessage"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// https://docs.docker.com/reference/api/engine/sdk/
func main() {
	// Get flag and validate path
	pathArg := flag.String("path", ".", "path to project")
	flag.Parse()

	absPath, err := project.Resolve(*pathArg)

	// Check if there is a valid dockerfile in it
	_, err = project.FindDockerfile(absPath)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// Files are there, go and start sdk now

	// Start sdk
	ctx := context.Background()
	apiClient, err := client.New(client.FromEnv)
	if err != nil {
		panic(err)
	}
	defer apiClient.Close()

	// Now build the image

	buildCtx, err := dockerimage.Context(absPath)
	buildResp, err := apiClient.ImageBuild(ctx, buildCtx, client.ImageBuildOptions{
		Dockerfile: "Dockerfile",
	})
	if err != nil {
		log.Fatalf("dockerfile invalid or build failed: %v", err)
	}
	defer buildResp.Body.Close()

	// Get imageID
	var imageID string
	err = jsonmessage.DisplayJSONMessagesStream(buildResp.Body, os.Stdout, os.Stdout.Fd(), false,
		func(msg jsonstream.Message) {
			if msg.Aux == nil {
				return
			}
			var result build.Result
			if err := json.Unmarshal(*msg.Aux, &result); err == nil && result.ID != "" {
				imageID = result.ID
			}
		},
	)
	if err != nil {
		log.Fatalf("error reading build output: %v", err)
	}
	if imageID == "" {
		log.Fatal("build succeeded but no image ID was returned")
	}
	log.Printf("built image=%s", imageID)

	// Get port from env and
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("port=%s", port)

	hostPort, err := network.ParsePort(port + "/tcp")
	if err != nil {
		log.Fatalf("could not parse port %q: %v", port, err)
	}

	resp, err := apiClient.ContainerCreate(ctx, client.ContainerCreateOptions{
		Image: imageID, // exact image id
		Config: &container.Config{
			Env:          []string{"PORT=" + port}, // pass the same port into the container
			ExposedPorts: network.PortSet{hostPort: struct{}{}},
		},
		HostConfig: &container.HostConfig{
			PortBindings: network.PortMap{
				hostPort: []network.PortBinding{
					{HostPort: port},
				},
			},
		},
	})

	if _, err := apiClient.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		log.Fatalf("could not start container: %v", err)
	}
	log.Printf("container started: %s", resp.ID)

	// Inspect the ports then log them
	inspectResp, err := apiClient.ContainerInspect(ctx, resp.ID, client.ContainerInspectOptions{})
	if err != nil {
		log.Printf("could not inspect container for port mappings: %v", err)
	} else if len(inspectResp.Container.NetworkSettings.Ports) == 0 {
		log.Println("no ports exposed by this image")
	} else {
		for containerPort, bindings := range inspectResp.Container.NetworkSettings.Ports {
			for _, binding := range bindings {
				log.Printf("port %s -> host %s:%s", containerPort, binding.HostIP, binding.HostPort)
			}
		}
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("shutting down, stopping container...")
	timeout := 10
	if _, err := apiClient.ContainerStop(ctx, resp.ID, client.ContainerStopOptions{Timeout: &timeout}); err != nil {
		log.Printf("could not stop container: %v", err)
	}
	if _, err := apiClient.ContainerRemove(ctx, resp.ID, client.ContainerRemoveOptions{Force: true}); err != nil {
		log.Printf("could not remove container: %v", err)
	}

}
