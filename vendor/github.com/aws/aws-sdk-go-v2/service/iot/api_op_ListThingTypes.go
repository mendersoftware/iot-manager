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

// Lists the existing thing types. Requires permission to access the ListThingTypes (https://docs.aws.amazon.com/service-authorization/latest/reference/list_awsiot.html#awsiot-actions-as-permissions)
// action.
func (c *Client) ListThingTypes(ctx context.Context, params *ListThingTypesInput, optFns ...func(*Options)) (*ListThingTypesOutput, error) {
	if params == nil {
		params = &ListThingTypesInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "ListThingTypes", params, optFns, c.addOperationListThingTypesMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*ListThingTypesOutput)
	out.ResultMetadata = metadata
	return out, nil
}

// The input for the ListThingTypes operation.
type ListThingTypesInput struct {

	// The maximum number of results to return in this operation.
	MaxResults *int32

	// To retrieve the next set of results, the nextToken value from a previous
	// response; otherwise null to receive the first set of results.
	NextToken *string

	// The name of the thing type.
	ThingTypeName *string

	noSmithyDocumentSerde
}

// The output for the ListThingTypes operation.
type ListThingTypesOutput struct {

	// The token for the next set of results. Will not be returned if operation has
	// returned all results.
	NextToken *string

	// The thing types.
	ThingTypes []types.ThingTypeDefinition

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationListThingTypesMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsRestjson1_serializeOpListThingTypes{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsRestjson1_deserializeOpListThingTypes{}, middleware.After)
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
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opListThingTypes(options.Region), middleware.Before); err != nil {
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

// ListThingTypesAPIClient is a client that implements the ListThingTypes
// operation.
type ListThingTypesAPIClient interface {
	ListThingTypes(context.Context, *ListThingTypesInput, ...func(*Options)) (*ListThingTypesOutput, error)
}

var _ ListThingTypesAPIClient = (*Client)(nil)

// ListThingTypesPaginatorOptions is the paginator options for ListThingTypes
type ListThingTypesPaginatorOptions struct {
	// The maximum number of results to return in this operation.
	Limit int32

	// Set to true if pagination should stop if the service returns a pagination token
	// that matches the most recent token provided to the service.
	StopOnDuplicateToken bool
}

// ListThingTypesPaginator is a paginator for ListThingTypes
type ListThingTypesPaginator struct {
	options   ListThingTypesPaginatorOptions
	client    ListThingTypesAPIClient
	params    *ListThingTypesInput
	nextToken *string
	firstPage bool
}

// NewListThingTypesPaginator returns a new ListThingTypesPaginator
func NewListThingTypesPaginator(client ListThingTypesAPIClient, params *ListThingTypesInput, optFns ...func(*ListThingTypesPaginatorOptions)) *ListThingTypesPaginator {
	if params == nil {
		params = &ListThingTypesInput{}
	}

	options := ListThingTypesPaginatorOptions{}
	if params.MaxResults != nil {
		options.Limit = *params.MaxResults
	}

	for _, fn := range optFns {
		fn(&options)
	}

	return &ListThingTypesPaginator{
		options:   options,
		client:    client,
		params:    params,
		firstPage: true,
		nextToken: params.NextToken,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *ListThingTypesPaginator) HasMorePages() bool {
	return p.firstPage || (p.nextToken != nil && len(*p.nextToken) != 0)
}

// NextPage retrieves the next ListThingTypes page.
func (p *ListThingTypesPaginator) NextPage(ctx context.Context, optFns ...func(*Options)) (*ListThingTypesOutput, error) {
	if !p.HasMorePages() {
		return nil, fmt.Errorf("no more pages available")
	}

	params := *p.params
	params.NextToken = p.nextToken

	var limit *int32
	if p.options.Limit > 0 {
		limit = &p.options.Limit
	}
	params.MaxResults = limit

	result, err := p.client.ListThingTypes(ctx, &params, optFns...)
	if err != nil {
		return nil, err
	}
	p.firstPage = false

	prevToken := p.nextToken
	p.nextToken = result.NextToken

	if p.options.StopOnDuplicateToken &&
		prevToken != nil &&
		p.nextToken != nil &&
		*prevToken == *p.nextToken {
		p.nextToken = nil
	}

	return result, nil
}

func newServiceMetadataMiddleware_opListThingTypes(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "execute-api",
		OperationName: "ListThingTypes",
	}
}
