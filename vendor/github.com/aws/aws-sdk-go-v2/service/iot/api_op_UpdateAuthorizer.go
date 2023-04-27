// Code generated by smithy-go-codegen DO NOT EDIT.

package iot

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Updates an authorizer. Requires permission to access the UpdateAuthorizer (https://docs.aws.amazon.com/service-authorization/latest/reference/list_awsiot.html#awsiot-actions-as-permissions)
// action.
func (c *Client) UpdateAuthorizer(ctx context.Context, params *UpdateAuthorizerInput, optFns ...func(*Options)) (*UpdateAuthorizerOutput, error) {
	if params == nil {
		params = &UpdateAuthorizerInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "UpdateAuthorizer", params, optFns, c.addOperationUpdateAuthorizerMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*UpdateAuthorizerOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type UpdateAuthorizerInput struct {

	// The authorizer name.
	//
	// This member is required.
	AuthorizerName *string

	// The ARN of the authorizer's Lambda function.
	AuthorizerFunctionArn *string

	// When true , the result from the authorizer’s Lambda function is cached for the
	// time specified in refreshAfterInSeconds . The cached result is used while the
	// device reuses the same HTTP connection.
	EnableCachingForHttp *bool

	// The status of the update authorizer request.
	Status types.AuthorizerStatus

	// The key used to extract the token from the HTTP headers.
	TokenKeyName *string

	// The public keys used to verify the token signature.
	TokenSigningPublicKeys map[string]string

	noSmithyDocumentSerde
}

type UpdateAuthorizerOutput struct {

	// The authorizer ARN.
	AuthorizerArn *string

	// The authorizer name.
	AuthorizerName *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationUpdateAuthorizerMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsRestjson1_serializeOpUpdateAuthorizer{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsRestjson1_deserializeOpUpdateAuthorizer{}, middleware.After)
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
	if err = addOpUpdateAuthorizerValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opUpdateAuthorizer(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opUpdateAuthorizer(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "execute-api",
		OperationName: "UpdateAuthorizer",
	}
}
