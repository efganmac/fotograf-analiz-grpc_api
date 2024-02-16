package main

import (
	"bytes"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	gen "gcs-imageUpload/api/gen"
	"gcs-imageUpload/internal/confluentkafka"
	"gcs-imageUpload/internal/vision"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/grpc/reflection"

	//"google.golang.org/genproto/googleapis/firestore/v1"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"os"
	"time"
)

type server struct {
	gen.UnimplementedImageManagementServiceServer
	storageClient *storage.Client
	bucketName    string
	producer      *confluentkafka.ProducerService
	KafkaConsumer *kafka.Consumer
	visionClient  *vision.VisionClient
}

func NewServer(ctx context.Context, storageClient *storage.Client, bucketName string, visionClient *vision.VisionClient, consumer *kafka.Consumer) *server {
	return &server{
		storageClient: storageClient,
		bucketName:    bucketName,
		visionClient:  visionClient,
		KafkaConsumer: consumer,
	}
}

func (s *server) UploadImage(ctx context.Context, req *gen.UploadImageRequest) (*gen.UploadImageResponse, error) {
	godotenv.Load()
	// upload file google cloud storage bucket
	object := s.storageClient.Bucket(s.bucketName).Object(req.Filename)
	writer := object.NewWriter(ctx)
	if _, err := io.Copy(writer, bytes.NewReader(req.Image)); err != nil {
		return nil, fmt.Errorf("failed to write image to GCS: %v", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close GCS writer: %v", err)
	}

	// generates signed private url
	url, err := generateSignedURL(ctx, s.storageClient, s.bucketName, req.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signed URL: %v", err)
	}

	bootstrap_servers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	group_id := os.Getenv("GROUP_ID")
	kafka_name := os.Getenv("KAFKA_API_NAME")
	kafka_passw := os.Getenv("KAFKA_API_PASSWORD")
	//KAFKA CONSUM
	configMap := &kafka.ConfigMap{
		"bootstrap.servers":  bootstrap_servers,
		"sasl.mechanisms":    "PLAIN",
		"security.protocol":  "SASL_SSL",
		"sasl.username":      kafka_name,
		"sasl.password":      kafka_passw,
		"group.id":           group_id,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "false",
	}
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	producer, err := kafka.NewProducer(configMap)
	if err != nil {
		log.Fatalf("Failed to create producer: %s", err)
	}

	defer producer.Close()

	topic := "caseStudy_topic"

	// convert json
	jsonData, err := json.Marshal(url)
	if err != nil {
		log.Fatalf("JSON marshal error: %v", err)
	}

	// send message kafka topic
	err = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          jsonData, // jsonData is json converted url
	}, nil)
	if err != nil {
		log.Fatalf("Failed to produce message: %v", err)
	}

	producer.Flush(15 * 1000)

	log.Println("Message sent to Kafka topic:", topic)

	return &gen.UploadImageResponse{Url: url}, nil
}

/*
	func (s *server) GetImageDetail(ctx context.Context, req *gen.GetImageDetailRequest) (*gen.GetImageDetailResponse, error) {
		log.Printf("Received GetImageDetail request for image_id: %s", req.GetImageId())
		docRef := s.firestoreClient.Collection("images").Doc(req.GetImageId())
		doc, err := docRef.Get(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get image details from Firestore: %v", err)
		}

		var annotations []*gen.FaceAnnotation
		if anns, ok := doc.Data()["annotations"].([]interface{}); ok {
			for _, ann := range anns {
				if m, ok := ann.(map[string]interface{}); ok {
					annotations = append(annotations, &gen.FaceAnnotation{
						Joy:      m["joy"].(string),
						Sorrow:   m["sorrow"].(string),
						Anger:    m["anger"].(string),
						Surprise: m["surprise"].(string),
					})
				}
			}
		}

		return &gen.GetImageDetailResponse{
			ImageId:     req.GetImageId(),
			Url:         doc.Data()["url"].(string),
			Annotations: annotations,
		}, nil
	}
*/
func (s *server) GetImageFeed(ctx context.Context, req *gen.GetImageFeedRequest) (*gen.GetImageFeedResponse, error) {
	log.Printf("Received GetImageFeed request for page: %d, page_size: %d", req.GetPage(), req.GetPageSize())

	return &gen.GetImageFeedResponse{
		Images: []*gen.ImageDetail{
			{ImageId: "1", Url: "http://example.com/image1.jpg", AverageEmotionScore: 0.8},
			{ImageId: "2", Url: "http://example.com/image2.jpg", AverageEmotionScore: 0.9},
		},
	}, nil
}

func (s *server) UpdateImageDetail(ctx context.Context, req *gen.UpdateImageDetailRequest) (*gen.UpdateImageDetailResponse, error) {
	log.Printf("Received UpdateImageDetail request for image_id: %s", req.GetImageId())
	return &gen.UpdateImageDetailResponse{
		Success: true,
	}, nil
}

func generateSignedURL(ctx context.Context, client *storage.Client, bucketName, objectName string) (string, error) {
	// generate signed url
	url, err := client.Bucket(bucketName).SignedURL(objectName, &storage.SignedURLOptions{

		Method:  "GET",
		Expires: time.Now().Add(48 * time.Hour),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create signed URL: %v", err)
	}
	return url, nil
}

func createFirestoreClient(ctx context.Context) (*firestore.Client, error) {
	projectID := os.Getenv("GCP_PROJECT_ID") // Ortam değişkenlerinden GCP proje ID'si alınır
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create Firestore client: %v", err)
	}
	return client, nil
}

type ImageData struct {
	URL string `json:"url"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	ctx := context.Background()

	storageClient, err := storage.NewClient(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		log.Fatalf("Failed to create storage client: %v", err)
	}

	firestoreClient, err := firestore.NewClient(ctx, os.Getenv("FIRESTORE_DB_ID"), option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}

	visionClient, err := vision.NewVisionClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Vision API client: %v", err)
	}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("bootstrap_servers"),
		"group.id":          "image-processing-service",
		"auto.offset.reset": "smallest"})
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	consumer.SubscribeTopics([]string{"caseStudy_topic"}, nil)

	// GRPC server conf
	port := os.Getenv("PORT")
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	bucketName := os.Getenv("GCP_STORAGE_BUCKET")
	grpcServer := grpc.NewServer()
	srv := NewServer(ctx, storageClient, bucketName, visionClient, consumer)
	gen.RegisterImageManagementServiceServer(grpcServer, srv)
	reflection.Register(grpcServer) //for curl

	log.Printf("Server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

	for {
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			fmt.Printf("Consumer error: %v\n", err)
			continue
		}

		var imageData ImageData
		if err := json.Unmarshal(msg.Value, &imageData); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		// analyze faces with visionapi
		annotations, err := visionClient.DetectFaces(ctx, imageData.URL)
		if err != nil {
			log.Printf("Failed to analyze image: %v", err)
			continue
		}

		// analyze log
		for _, annotation := range annotations {
			log.Printf("Joy: %v, Sorrow: %v, Anger: %v, Surprise: %v",
				annotation.JoyLikelihood, annotation.SorrowLikelihood,
				annotation.AngerLikelihood, annotation.SurpriseLikelihood)
		}
	}
}
