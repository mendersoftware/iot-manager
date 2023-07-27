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
	"time"
)

// List of Device Defender ML Detect mitigation actions tasks. Requires permission
// to access the ListDetectMitigationActionsTasks (https://docs.aws.amazon.com/service-authorization/latest/reference/list_awsiot.html#awsiot-actions-as-permissions)
// action.
func (c *Client) ListDetectMitigationActionsTasks(ctx context.Context, params *ListDetectMitigationActionsTasksInput, optFns ...func(*Options)) (*ListDetectMitigationActionsTasksOutput, error) {
	if params == nil {
		params = &ListDetectMitigationActionsTasksInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "ListDetectMitigationActionsTasks", params, optFns, c.addOperationListDetectMitigationActionsTasksMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*ListDetectMitigationActionsTasksOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type ListDetectMitigationActionsTasksInput struct {

	// The end of the time period for which ML Detect mitigation actions tasks are
	// returned.
	//
	// This member is required.
	EndTime *time.Time

	// A filter to limit results to those found after the specified time. You must
	// specify either the startTime and endTime or the taskId, but not both.
	//
	// This member is required.
	StartTime *time.Time

	// The maximum number of results to return at one time. The default is 25.
	MaxResults *int32

	// The token for the next set of results.
	NextToken *string

	noSmithyDocumentSerde
}

type ListDetectMitigationActionsTasksOutput struct {

	// A token that can be used to retrieve the next set of results, or null if there
	// are no additional results.
	NextToken *string

	// The collection of ML Detect mitigation tasks that matched the filter criteria.
	Tasks []types.DetectMitigationActionsTaskSummary

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationListDetectMitigationActionsTasksMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsRestjson1_serializeOpListDetectMitigationActionsTasks{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsRestjson1_deserializeOpListDetectMitigationActionsTasks{}, middleware.After)
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
	if err = addClientUserAgent(stack, options); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = addOpListDetectMitigationActionsTasksValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opListDetectMitigationActionsTasks(options.Region), middleware.Before); err != nil {
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

// ListDetectMitigationActionsTasksAPIClient is a client that implements the
// ListDetectMitigationActionsTasks operation.
type ListDetectMitigationActionsTasksAPIClient interface {
	ListDetectMitigationActionsTasks(context.Context, *ListDetectMitigationActionsTasksInput, ...func(*Options)) (*ListDetectMitigationActionsTasksOutput, error)
}

var _ ListDetectMitigationActionsTasksAPIClient = (*Client)(nil)

// ListDetectMitigationActionsTasksPaginatorOptions is the paginator options for
// ListDetectMitigationActionsTasks
type ListDetectMitigationActionsTasksPaginatorOptions struct {
	// The maximum number of results to return at one time. The default is 25.
	Limit int32

	// Set to true if pagination should stop if the service returns a pagination token
	// that matches the most recent token provided to the service.
	StopOnDuplicateToken bool
}

// ListDetectMitigationActionsTasksPaginator is a paginator for
// ListDetectMitigationActionsTasks
type ListDetectMitigationActionsTasksPaginator struct {
	options   ListDetectMitigationActionsTasksPaginatorOptions
	client    ListDetectMitigationActionsTasksAPIClient
	params    *ListDetectMitigationActionsTasksInput
	nextToken *string
	firstPage bool
}

// NewListDetectMitigationActionsTasksPaginator returns a new
// ListDetectMitigationActionsTasksPaginator
func NewListDetectMitigationActionsTasksPaginator(client ListDetectMitigationActionsTasksAPIClient, params *ListDetectMitigationActionsTasksInput, optFns ...func(*ListDetectMitigationActionsTasksPaginatorOptions)) *ListDetectMitigationActionsTasksPaginator {
	if params == nil {
		params = &ListDetectMitigationActionsTasksInput{}
	}

	options := ListDetectMitigationActionsTasksPaginatorOptions{}
	if params.MaxResults != nil {
		options.Limit = *params.MaxResults
	}

	for _, fn := range optFns {
		fn(&options)
	}

	return &ListDetectMitigationActionsTasksPaginator{
		options:   options,
		client:    client,
		params:    params,
		firstPage: true,
		nextToken: params.NextToken,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *ListDetectMitigationActionsTasksPaginator) HasMorePages() bool {
	return p.firstPage || (p.nextToken != nil && len(*p.nextToken) != 0)
}

// NextPage retrieves the next ListDetectMitigationActionsTasks page.
func (p *ListDetectMitigationActionsTasksPaginator) NextPage(ctx context.Context, optFns ...func(*Options)) (*ListDetectMitigationActionsTasksOutput, error) {
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

	result, err := p.client.ListDetectMitigationActionsTasks(ctx, &params, optFns...)
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

func newServiceMetadataMiddleware_opListDetectMitigationActionsTasks(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "iot",
		OperationName: "ListDetectMitigationActionsTasks",
	}
}
