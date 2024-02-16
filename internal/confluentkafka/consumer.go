package confluentkafka

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/joho/godotenv"
	visionpb "google.golang.org/genproto/googleapis/cloud/vision/v1"
	"os"
)

type VisionAPI interface {
	AnalyzeImage(ctx context.Context, imageURL string) ([]*visionpb.FaceAnnotation, error)
}

type ConsumerService struct {
	consumer *kafka.Consumer
	vision   VisionAPI
}

type ImageData struct {
	URL string `json:"url"`
}

func (c *ConsumerService) saveAnalysisResultsToFirestore(ctx context.Context, imageURL string, annotations []*visionpb.FaceAnnotation) error {
	godotenv.Load("Users/efekanbicer/go/src/gcs-imageUpload/.env")
	doc_id := os.Getenv("DOC_ID")

	firestoreClient, err := firestore.NewClient(ctx, "sage-sylph-414022")
	if err != nil {
		return fmt.Errorf("failed to create Firestore client: %v", err)
	}
	defer firestoreClient.Close()

	// Analiz sonuçlarını Firestore'a kaydet
	docRef := firestoreClient.Collection("imageAnalyses").Doc(doc_id) // Benzersiz bir doküman ID'si için Doc() kullanılır
	_, err = docRef.Set(ctx, map[string]interface{}{
		"imageURL":    imageURL,
		"annotations": annotations, // Bu, doğrudan kaydedilemeyebilir, uygun bir yapıya dönüştürmeniz gerekebilir
	})
	if err != nil {
		return fmt.Errorf("failed to save analysis results to Firestore: %v", err)
	}
	return nil

}
