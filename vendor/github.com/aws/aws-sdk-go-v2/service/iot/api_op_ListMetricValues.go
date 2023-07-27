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

// Lists the values reported for an IoT Device Defender metric (device-side
// metric, cloud-side metric, or custom metric) by the given thing during the
// specified time period.
func (c *Client) ListMetricValues(ctx context.Context, params *ListMetricValuesInput, optFns ...func(*Options)) (*ListMetricValuesOutput, error) {
	if params == nil {
		params = &ListMetricValuesInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "ListMetricValues", params, optFns, c.addOperationListMetricValuesMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*ListMetricValuesOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type ListMetricValuesInput struct {

	// The end of the time period for which metric values are returned.
	//
	// This member is required.
	EndTime *time.Time

	// The name of the security profile metric for which values are returned.
	//
	// This member is required.
	MetricName *string

	// The start of the time period for which metric values are returned.
	//
	// This member is required.
	StartTime *time.Time

	// The name of the thing for which security profile metric values are returned.
	//
	// This member is required.
	ThingName *string

	// The dimension name.
	DimensionName *string

	// The dimension value operator.
	DimensionValueOperator types.DimensionValueOperator

	// The maximum number of results to return at one time.
	MaxResults *int32

	// The token for the next set of results.
	NextToken *string

	noSmithyDocumentSerde
}

type ListMetricValuesOutput struct {

	// The data the thing reports for the metric during the specified time period.
	MetricDatumList []types.MetricDatum

	// A token that can be used to retrieve the next set of results, or null if there
	// are no additional results.
	NextToken *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationListMetricValuesMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsRestjson1_serializeOpListMetricValues{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsRestjson1_deserializeOpListMetricValues{}, middleware.After)
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
	if err = addOpListMetricValuesValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opListMetricValues(options.Region), middleware.Before); err != nil {
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

// ListMetricValuesAPIClient is a client that implements the ListMetricValues
// operation.
type ListMetricValuesAPIClient interface {
	ListMetricValues(context.Context, *ListMetricValuesInput, ...func(*Options)) (*ListMetricValuesOutput, error)
}

var _ ListMetricValuesAPIClient = (*Client)(nil)

// ListMetricValuesPaginatorOptions is the paginator options for ListMetricValues
type ListMetricValuesPaginatorOptions struct {
	// The maximum number of results to return at one time.
	Limit int32

	// Set to true if pagination should stop if the service returns a pagination token
	// that matches the most recent token provided to the service.
	StopOnDuplicateToken bool
}

// ListMetricValuesPaginator is a paginator for ListMetricValues
type ListMetricValuesPaginator struct {
	options   ListMetricValuesPaginatorOptions
	client    ListMetricValuesAPIClient
	params    *ListMetricValuesInput
	nextToken *string
	firstPage bool
}

// NewListMetricValuesPaginator returns a new ListMetricValuesPaginator
func NewListMetricValuesPaginator(client ListMetricValuesAPIClient, params *ListMetricValuesInput, optFns ...func(*ListMetricValuesPaginatorOptions)) *ListMetricValuesPaginator {
	if params == nil {
		params = &ListMetricValuesInput{}
	}

	options := ListMetricValuesPaginatorOptions{}
	if params.MaxResults != nil {
		options.Limit = *params.MaxResults
	}

	for _, fn := range optFns {
		fn(&options)
	}

	return &ListMetricValuesPaginator{
		options:   options,
		client:    client,
		params:    params,
		firstPage: true,
		nextToken: params.NextToken,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *ListMetricValuesPaginator) HasMorePages() bool {
	return p.firstPage || (p.nextToken != nil && len(*p.nextToken) != 0)
}

// NextPage retrieves the next ListMetricValues page.
func (p *ListMetricValuesPaginator) NextPage(ctx context.Context, optFns ...func(*Options)) (*ListMetricValuesOutput, error) {
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

	result, err := p.client.ListMetricValues(ctx, &params, optFns...)
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

func newServiceMetadataMiddleware_opListMetricValues(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "iot",
		OperationName: "ListMetricValues",
	}
}
