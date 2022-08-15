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

// Lists the findings (results) of a Device Defender audit or of the audits
// performed during a specified time period. (Findings are retained for 90 days.)
// Requires permission to access the ListAuditFindings
// (https://docs.aws.amazon.com/service-authorization/latest/reference/list_awsiot.html#awsiot-actions-as-permissions)
// action.
func (c *Client) ListAuditFindings(ctx context.Context, params *ListAuditFindingsInput, optFns ...func(*Options)) (*ListAuditFindingsOutput, error) {
	if params == nil {
		params = &ListAuditFindingsInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "ListAuditFindings", params, optFns, c.addOperationListAuditFindingsMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*ListAuditFindingsOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type ListAuditFindingsInput struct {

	// A filter to limit results to the findings for the specified audit check.
	CheckName *string

	// A filter to limit results to those found before the specified time. You must
	// specify either the startTime and endTime or the taskId, but not both.
	EndTime *time.Time

	// Boolean flag indicating whether only the suppressed findings or the unsuppressed
	// findings should be listed. If this parameter isn't provided, the response will
	// list both suppressed and unsuppressed findings.
	ListSuppressedFindings bool

	// The maximum number of results to return at one time. The default is 25.
	MaxResults *int32

	// The token for the next set of results.
	NextToken *string

	// Information identifying the noncompliant resource.
	ResourceIdentifier *types.ResourceIdentifier

	// A filter to limit results to those found after the specified time. You must
	// specify either the startTime and endTime or the taskId, but not both.
	StartTime *time.Time

	// A filter to limit results to the audit with the specified ID. You must specify
	// either the taskId or the startTime and endTime, but not both.
	TaskId *string

	noSmithyDocumentSerde
}

type ListAuditFindingsOutput struct {

	// The findings (results) of the audit.
	Findings []types.AuditFinding

	// A token that can be used to retrieve the next set of results, or null if there
	// are no additional results.
	NextToken *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationListAuditFindingsMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsRestjson1_serializeOpListAuditFindings{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsRestjson1_deserializeOpListAuditFindings{}, middleware.After)
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
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opListAuditFindings(options.Region), middleware.Before); err != nil {
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

// ListAuditFindingsAPIClient is a client that implements the ListAuditFindings
// operation.
type ListAuditFindingsAPIClient interface {
	ListAuditFindings(context.Context, *ListAuditFindingsInput, ...func(*Options)) (*ListAuditFindingsOutput, error)
}

var _ ListAuditFindingsAPIClient = (*Client)(nil)

// ListAuditFindingsPaginatorOptions is the paginator options for ListAuditFindings
type ListAuditFindingsPaginatorOptions struct {
	// The maximum number of results to return at one time. The default is 25.
	Limit int32

	// Set to true if pagination should stop if the service returns a pagination token
	// that matches the most recent token provided to the service.
	StopOnDuplicateToken bool
}

// ListAuditFindingsPaginator is a paginator for ListAuditFindings
type ListAuditFindingsPaginator struct {
	options   ListAuditFindingsPaginatorOptions
	client    ListAuditFindingsAPIClient
	params    *ListAuditFindingsInput
	nextToken *string
	firstPage bool
}

// NewListAuditFindingsPaginator returns a new ListAuditFindingsPaginator
func NewListAuditFindingsPaginator(client ListAuditFindingsAPIClient, params *ListAuditFindingsInput, optFns ...func(*ListAuditFindingsPaginatorOptions)) *ListAuditFindingsPaginator {
	if params == nil {
		params = &ListAuditFindingsInput{}
	}

	options := ListAuditFindingsPaginatorOptions{}
	if params.MaxResults != nil {
		options.Limit = *params.MaxResults
	}

	for _, fn := range optFns {
		fn(&options)
	}

	return &ListAuditFindingsPaginator{
		options:   options,
		client:    client,
		params:    params,
		firstPage: true,
		nextToken: params.NextToken,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *ListAuditFindingsPaginator) HasMorePages() bool {
	return p.firstPage || (p.nextToken != nil && len(*p.nextToken) != 0)
}

// NextPage retrieves the next ListAuditFindings page.
func (p *ListAuditFindingsPaginator) NextPage(ctx context.Context, optFns ...func(*Options)) (*ListAuditFindingsOutput, error) {
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

	result, err := p.client.ListAuditFindings(ctx, &params, optFns...)
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

func newServiceMetadataMiddleware_opListAuditFindings(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "execute-api",
		OperationName: "ListAuditFindings",
	}
}
