// Code generated by smithy-go-codegen DO NOT EDIT.

package iotdataplane

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Gets the details of a single retained message for the specified topic. This
// action returns the message payload of the retained message, which can incur
// messaging costs. To list only the topic names of the retained messages, call
// ListRetainedMessages (https://docs.aws.amazon.com/iot/latest/apireference/API_iotdata_ListRetainedMessages.html)
// . Requires permission to access the GetRetainedMessage (https://docs.aws.amazon.com/service-authorization/latest/reference/list_awsiotfleethubfordevicemanagement.html#awsiotfleethubfordevicemanagement-actions-as-permissions)
// action. For more information about messaging costs, see Amazon Web Services IoT
// Core pricing - Messaging (http://aws.amazon.com/iot-core/pricing/#Messaging) .
func (c *Client) GetRetainedMessage(ctx context.Context, params *GetRetainedMessageInput, optFns ...func(*Options)) (*GetRetainedMessageOutput, error) {
	if params == nil {
		params = &GetRetainedMessageInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "GetRetainedMessage", params, optFns, c.addOperationGetRetainedMessageMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*GetRetainedMessageOutput)
	out.ResultMetadata = metadata
	return out, nil
}

// The input for the GetRetainedMessage operation.
type GetRetainedMessageInput struct {

	// The topic name of the retained message to retrieve.
	//
	// This member is required.
	Topic *string

	noSmithyDocumentSerde
}

// The output from the GetRetainedMessage operation.
type GetRetainedMessageOutput struct {

	// The Epoch date and time, in milliseconds, when the retained message was stored
	// by IoT.
	LastModifiedTime int64

	// The Base64-encoded message payload of the retained message body.
	Payload []byte

	// The quality of service (QoS) level used to publish the retained message.
	Qos int32

	// The topic name to which the retained message was published.
	Topic *string

	// A base64-encoded JSON string that includes an array of JSON objects, or null if
	// the retained message doesn't include any user properties. The following example
	// userProperties parameter is a JSON string that represents two user properties.
	// Note that it will be base64-encoded: [{"deviceName": "alpha"}, {"deviceCnt":
	// "45"}]
	UserProperties []byte

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationGetRetainedMessageMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsRestjson1_serializeOpGetRetainedMessage{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsRestjson1_deserializeOpGetRetainedMessage{}, middleware.After)
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
	if err = addOpGetRetainedMessageValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opGetRetainedMessage(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opGetRetainedMessage(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "iotdata",
		OperationName: "GetRetainedMessage",
	}
}
