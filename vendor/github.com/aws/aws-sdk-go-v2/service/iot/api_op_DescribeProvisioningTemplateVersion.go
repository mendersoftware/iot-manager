// Code generated by smithy-go-codegen DO NOT EDIT.

package iot

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"time"
)

// Returns information about a provisioning template version. Requires permission
// to access the DescribeProvisioningTemplateVersion (https://docs.aws.amazon.com/service-authorization/latest/reference/list_awsiot.html#awsiot-actions-as-permissions)
// action.
func (c *Client) DescribeProvisioningTemplateVersion(ctx context.Context, params *DescribeProvisioningTemplateVersionInput, optFns ...func(*Options)) (*DescribeProvisioningTemplateVersionOutput, error) {
	if params == nil {
		params = &DescribeProvisioningTemplateVersionInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "DescribeProvisioningTemplateVersion", params, optFns, c.addOperationDescribeProvisioningTemplateVersionMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*DescribeProvisioningTemplateVersionOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type DescribeProvisioningTemplateVersionInput struct {

	// The template name.
	//
	// This member is required.
	TemplateName *string

	// The provisioning template version ID.
	//
	// This member is required.
	VersionId *int32

	noSmithyDocumentSerde
}

type DescribeProvisioningTemplateVersionOutput struct {

	// The date when the provisioning template version was created.
	CreationDate *time.Time

	// True if the provisioning template version is the default version.
	IsDefaultVersion bool

	// The JSON formatted contents of the provisioning template version.
	TemplateBody *string

	// The provisioning template version ID.
	VersionId *int32

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationDescribeProvisioningTemplateVersionMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsRestjson1_serializeOpDescribeProvisioningTemplateVersion{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsRestjson1_deserializeOpDescribeProvisioningTemplateVersion{}, middleware.After)
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
	if err = addOpDescribeProvisioningTemplateVersionValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opDescribeProvisioningTemplateVersion(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opDescribeProvisioningTemplateVersion(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "iot",
		OperationName: "DescribeProvisioningTemplateVersion",
	}
}
