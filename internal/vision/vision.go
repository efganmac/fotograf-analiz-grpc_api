package vision

import (
	vision "cloud.google.com/go/vision/apiv1" // vision için apiv1 takma adı kullanılıyor
	"context"
	visionpb "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

// Vision Client
type VisionClient struct {
	Client *vision.ImageAnnotatorClient
}

// NewVisionClient, create ne
func NewVisionClient(ctx context.Context) (*VisionClient, error) {
	client, err := vision.NewImageAnnotatorClient(ctx) // vision takma adıyla fonksiyon çağrılıyor
	if err != nil {
		return nil, err
	}
	return &VisionClient{Client: client}, nil
}

// AnalyzeImage,
// analyzes the specified image and the face changes.
func (vc *VisionClient) DetectFaces(ctx context.Context, gcsURI string) ([]*visionpb.FaceAnnotation, error) {
	image := vision.NewImageFromURI(gcsURI) // vision takma adıyla fonksiyon çağrılıyor
	annotations, err := vc.Client.DetectFaces(ctx, image, nil, 10)
	if err != nil {
		return nil, err
	}
	return annotations, nil
}
