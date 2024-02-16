package main

import (
	"context"
	"flag"
	"fmt"
	gen "gcs-imageUpload/api/gen"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	serverAddr := flag.String("addr", "localhost:50051", "the server address in the format of host:port")
	imageName := flag.String("image", "", "the name of the image file to upload within the /Users/efekanbicer/go/src/gcs-imageUpload/photos directory")
	flag.Parse()

	if *imageName == "" {
		log.Fatal("You must specify the image name using the -image flag.")
	}

	// Base Path
	basePath := "/Users/efekanbicer/go/src/gcs-imageUpload/photos"
	imagePath := filepath.Join(basePath, *imageName)
	// handle error
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		log.Fatalf("The file %s does not exist.", imagePath)
	}

	// connect GRPC server
	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	c := gen.NewImageManagementServiceClient(conn)

	// Upload Image
	uploadImage(c, imagePath)
}

func uploadImage(c gen.ImageManagementServiceClient, imagePath string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	imageData, err := ioutil.ReadFile(imagePath)
	if err != nil {
		log.Fatalf("Failed to read image file: %v", err)
	}

	// send request with GRPC
	resp, err := c.UploadImage(ctx, &gen.UploadImageRequest{
		Filename: filepath.Base(imagePath),
		Image:    imageData,
	})
	if err != nil {
		log.Fatalf("Failed to upload image: %v", err)
	} else {
		fmt.Printf("Image uploaded successfully: %s\n", resp.GetUrl())

	}

}
