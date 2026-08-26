package qcloud

import (
	"testing"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type fakeImageClient struct {
	responses []*cvm.DescribeImagesResponse
	requests  []*cvm.DescribeImagesRequest
}

func (f *fakeImageClient) DescribeImages(request *cvm.DescribeImagesRequest) (*cvm.DescribeImagesResponse, error) {
	f.requests = append(f.requests, request)
	response := f.responses[0]
	f.responses = f.responses[1:]
	return response, nil
}

func TestListImagesStopsOnEmptyPage(t *testing.T) {
	client := &fakeImageClient{
		responses: []*cvm.DescribeImagesResponse{describeImagesResponse(1)},
	}

	_, err := listImages(client, false)
	if err == nil {
		t.Fatal("listImages() error = nil, want empty page error")
	}
	if len(client.requests) != 1 {
		t.Fatalf("DescribeImages() calls = %d, want 1", len(client.requests))
	}
}

func TestListImagesRejectsInvalidResponse(t *testing.T) {
	tests := []struct {
		name     string
		response *cvm.DescribeImagesResponse
	}{
		{name: "nil response", response: nil},
		{name: "nil response body", response: &cvm.DescribeImagesResponse{}},
		{name: "nil total count", response: describeImagesResponseWithoutTotalCount()},
		{name: "negative total count", response: describeImagesResponse(-1)},
		{name: "nil image", response: describeImagesResponse(1, nil)},
		{name: "nil image type", response: describeImagesResponse(1, &cvm.Image{})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeImageClient{responses: []*cvm.DescribeImagesResponse{tt.response}}
			if _, err := listImages(client, false); err == nil {
				t.Fatal("listImages() error = nil, want invalid response error")
			}
		})
	}
}

func TestListImagesClassifiesKnownAndUnknownTypes(t *testing.T) {
	client := &fakeImageClient{
		responses: []*cvm.DescribeImagesResponse{describeImagesResponse(3,
			newImage("PUBLIC_IMAGE"),
			newImage("PRIVATE_IMAGE"),
			newImage("NEW_IMAGE_TYPE"),
		)},
	}

	images, err := listImages(client, false)
	if err != nil {
		t.Fatalf("listImages() error = %v", err)
	}
	wantTypes := []string{"官方", "自定义镜像", "未知"}
	if len(images) != len(wantTypes) {
		t.Fatalf("listImages() returned %d images, want %d", len(images), len(wantTypes))
	}
	for i, want := range wantTypes {
		if images[i].ImageType != want {
			t.Errorf("images[%d].ImageType = %q, want %q", i, images[i].ImageType, want)
		}
	}
}

func TestListImagesHandlesOptionalFields(t *testing.T) {
	image := newImage("PRIVATE_IMAGE")
	image.ImageName = nil
	image.ImageState = nil
	image.ImageDescription = nil
	image.OsName = nil
	client := &fakeImageClient{
		responses: []*cvm.DescribeImagesResponse{describeImagesResponse(1, image)},
	}

	images, err := listImages(client, false)
	if err != nil {
		t.Fatalf("listImages() error = %v", err)
	}
	if len(images) != 1 {
		t.Fatalf("listImages() returned %d images, want 1", len(images))
	}
	if images[0].ImageState != "未知" {
		t.Errorf("images[0].ImageState = %q, want %q", images[0].ImageState, "未知")
	}
}

func describeImagesResponse(total int64, images ...*cvm.Image) *cvm.DescribeImagesResponse {
	return &cvm.DescribeImagesResponse{
		Response: &cvm.DescribeImagesResponseParams{
			TotalCount: common.Int64Ptr(total),
			ImageSet:   images,
		},
	}
}

func describeImagesResponseWithoutTotalCount() *cvm.DescribeImagesResponse {
	return &cvm.DescribeImagesResponse{
		Response: &cvm.DescribeImagesResponseParams{},
	}
}

func newImage(imageType string) *cvm.Image {
	return &cvm.Image{
		ImageId:          common.StringPtr("img-1"),
		OsName:           common.StringPtr("Linux"),
		ImageType:        common.StringPtr(imageType),
		ImageName:        common.StringPtr("test-image"),
		ImageDescription: common.StringPtr("test image"),
		ImageState:       common.StringPtr("NORMAL"),
	}
}
