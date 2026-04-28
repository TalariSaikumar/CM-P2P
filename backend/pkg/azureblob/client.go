package azureblob

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

// Client wraps Azure Blob operations for car images and KYC documents.
type Client struct {
	account   string
	container string
	svc       *azblob.Client
}

// NewClient builds a blob client using the storage account name and account key.
func NewClient(accountName, accountKey, containerName string) (*Client, error) {
	if accountName == "" || accountKey == "" || containerName == "" {
		return nil, fmt.Errorf("azure blob: account, key, and container are required")
	}

	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, fmt.Errorf("azure blob: credential: %w", err)
	}

	svcURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)
	svc, err := azblob.NewClientWithSharedKeyCredential(svcURL, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("azure blob: client: %w", err)
	}

	return &Client{
		account:   accountName,
		container: containerName,
		svc:       svc,
	}, nil
}

func (c *Client) containerClient() *container.Client {
	return c.svc.ServiceClient().NewContainerClient(c.container)
}

// Upload puts an object into the configured container and returns its HTTPS URL (no SAS).
func (c *Client) Upload(ctx context.Context, blobName string, body io.Reader, contentType string) (string, error) {
	blobName = strings.TrimPrefix(blobName, "/")
	cc := c.containerClient()
	bc := cc.NewBlockBlobClient(blobName)

	opts := &blockblob.UploadStreamOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: to.Ptr(contentType),
		},
	}

	if _, err := bc.UploadStream(ctx, body, opts); err != nil {
		return "", fmt.Errorf("azure blob: upload %q: %w", blobName, err)
	}

	return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", c.account, c.container, blobName), nil
}

// Delete removes a blob by name (container-relative path).
func (c *Client) Delete(ctx context.Context, blobName string) error {
	blobName = strings.TrimPrefix(blobName, "/")
	bc := c.containerClient().NewBlobClient(blobName)
	if _, err := bc.Delete(ctx, nil); err != nil {
		return fmt.Errorf("azure blob: delete %q: %w", blobName, err)
	}
	return nil
}
