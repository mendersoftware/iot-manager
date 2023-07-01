// Code generated by smithy-go-codegen DO NOT EDIT.

package iot

import (
	"context"
	"fmt"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Lists the authorizers registered in your account. Requires permission to access
// the ListAuthorizers (https://docs.aws.amazon.com/service-authorization/latest/reference/list_awsiot.html#awsiot-actions-as-permissions)
// action.
func (c *Client) ListAuthorizers(ctx context.Context, params *ListAuthorizersInput, optFns ...func(*Options)) (*ListAuthorizersOutput, error) {
	if params == nil {
		params = &ListAuthorizersInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "ListAuthorizers", params, optFns, c.addOperationListAuthorizersMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*ListAuthorizersOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type ListAuthorizersInput struct {

	// Return the list of authorizers in ascending alphabetical order.
	AscendingOrder bool

	// A marker used to get the next set of results.
	Marker *string

	// The maximum number of results to return at one time.
	PageSize *int32

	// The status of the list authorizers request.
	Status types.AuthorizerStatus

	noSmithyDocumentSerde
}

type ListAuthorizersOutput struct {

	// The authorizers.
	Authorizers []types.AuthorizerSummary

	// A marker used to get the next set of results.
	NextMarker *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationListAuthorizersMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsRestjson1_serializeOpListAuthorizers{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsRestjson1_deserializeOpListAuthorizers{}, middleware.After)
	if err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddClientRequestIDMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddComputeContentLengthMiddleware(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = v4.AddComputePayloadSHA256Middleware(stack); err != nil {
		return err
	}
	if err = addRetryMiddlewares(stack, options); err != nil {
		return err
	}
	if err = addHTTPSignerV4Middleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = addClientUserAgent(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opListAuthorizers(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecursionDetection(stack); err != nil {
		return err
	}
	if err = addRequestIDRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	return nil
}

// ListAuthorizersAPIClient is a client that implements the ListAuthorizers
// operation.
type ListAuthorizersAPIClient interface {
	ListAuthorizers(context.Context, *ListAuthorizersInput, ...func(*Options)) (*ListAuthorizersOutput, error)
}

var _ ListAuthorizersAPIClient = (*Client)(nil)

// ListAuthorizersPaginatorOptions is the paginator options for ListAuthorizers
type ListAuthorizersPaginatorOptions struct {
	// The maximum number of results to return at one time.
	Limit int32

	// Set to true if pagination should stop if the service returns a pagination token
	// that matches the most recent token provided to the service.
	StopOnDuplicateToken bool
}

// ListAuthorizersPaginator is a paginator for ListAuthorizers
type ListAuthorizersPaginator struct {
	options   ListAuthorizersPaginatorOptions
	client    ListAuthorizersAPIClient
	params    *ListAuthorizersInput
	nextToken *string
	firstPage bool
}

// NewListAuthorizersPaginator returns a new ListAuthorizersPaginator
func NewListAuthorizersPaginator(client ListAuthorizersAPIClient, params *ListAuthorizersInput, optFns ...func(*ListAuthorizersPaginatorOptions)) *ListAuthorizersPaginator {
	if params == nil {
		params = &ListAuthorizersInput{}
	}

	options := ListAuthorizersPaginatorOptions{}
	if params.PageSize != nil {
		options.Limit = *params.PageSize
	}

	for _, fn := range optFns {
		fn(&options)
	}

	return &ListAuthorizersPaginator{
		options:   options,
		client:    client,
		params:    params,
		firstPage: true,
		nextToken: params.Marker,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *ListAuthorizersPaginator) HasMorePages() bool {
	return p.firstPage || (p.nextToken != nil && len(*p.nextToken) != 0)
}

// NextPage retrieves the next ListAuthorizers page.
func (p *ListAuthorizersPaginator) NextPage(ctx context.Context, optFns ...func(*Options)) (*ListAuthorizersOutput, error) {
	if !p.HasMorePages() {
		return nil, fmt.Errorf("no more pages available")
	}

	params := *p.params
	params.Marker = p.nextToken

	var limit *int32
	if p.options.Limit > 0 {
		limit = &p.options.Limit
	}
	params.PageSize = limit

	result, err := p.client.ListAuthorizers(ctx, &params, optFns...)
	if err != nil {
		return nil, err
	}
	p.firstPage = false

	prevToken := p.nextToken
	p.nextToken = result.NextMarker

	if p.options.StopOnDuplicateToken &&
		prevToken != nil &&
		p.nextToken != nil &&
		*prevToken == *p.nextToken {
		p.nextToken = nil
	}

	return result, nil
}

func newServiceMetadataMiddleware_opListAuthorizers(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "iot",
		OperationName: "ListAuthorizers",
	}
}
