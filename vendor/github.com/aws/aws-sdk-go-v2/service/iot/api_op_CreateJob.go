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

// Creates a job. Requires permission to access the CreateJob (https://docs.aws.amazon.com/service-authorization/latest/reference/list_awsiot.html#awsiot-actions-as-permissions)
// action.
func (c *Client) CreateJob(ctx context.Context, params *CreateJobInput, optFns ...func(*Options)) (*CreateJobOutput, error) {
	if params == nil {
		params = &CreateJobInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "CreateJob", params, optFns, c.addOperationCreateJobMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*CreateJobOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type CreateJobInput struct {

	// A job identifier which must be unique for your Amazon Web Services account. We
	// recommend using a UUID. Alpha-numeric characters, "-" and "_" are valid for use
	// here.
	//
	// This member is required.
	JobId *string

	// A list of things and thing groups to which the job should be sent.
	//
	// This member is required.
	Targets []string

	// Allows you to create the criteria to abort a job.
	AbortConfig *types.AbortConfig

	// A short text description of the job.
	Description *string

	// The job document. Required if you don't specify a value for documentSource .
	Document *string

	// Parameters of an Amazon Web Services managed template that you can specify to
	// create the job document. documentParameters can only be used when creating jobs
	// from Amazon Web Services managed templates. This parameter can't be used with
	// custom job templates or to create jobs from them.
	DocumentParameters map[string]string

	// An S3 link, or S3 object URL, to the job document. The link is an Amazon S3
	// object URL and is required if you don't specify a value for document . For
	// example, --document-source
	// https://s3.region-code.amazonaws.com/example-firmware/device-firmware.1.0 . For
	// more information, see Methods for accessing a bucket (https://docs.aws.amazon.com/AmazonS3/latest/userguide/access-bucket-intro.html)
	// .
	DocumentSource *string

	// Allows you to create the criteria to retry a job.
	JobExecutionsRetryConfig *types.JobExecutionsRetryConfig

	// Allows you to create a staged rollout of the job.
	JobExecutionsRolloutConfig *types.JobExecutionsRolloutConfig

	// The ARN of the job template used to create the job.
	JobTemplateArn *string

	// The namespace used to indicate that a job is a customer-managed job. When you
	// specify a value for this parameter, Amazon Web Services IoT Core sends jobs
	// notifications to MQTT topics that contain the value in the following format.
	// $aws/things/THING_NAME/jobs/JOB_ID/notify-namespace-NAMESPACE_ID/ The
	// namespaceId feature is in public preview.
	NamespaceId *string

	// Configuration information for pre-signed S3 URLs.
	PresignedUrlConfig *types.PresignedUrlConfig

	// The configuration that allows you to schedule a job for a future date and time
	// in addition to specifying the end behavior for each job execution.
	SchedulingConfig *types.SchedulingConfig

	// Metadata which can be used to manage the job.
	Tags []types.Tag

	// Specifies whether the job will continue to run (CONTINUOUS), or will be
	// complete after all those things specified as targets have completed the job
	// (SNAPSHOT). If continuous, the job may also be run on a thing when a change is
	// detected in a target. For example, a job will run on a thing when the thing is
	// added to a target group, even after the job was completed by all things
	// originally in the group. We recommend that you use continuous jobs instead of
	// snapshot jobs for dynamic thing group targets. By using continuous jobs, devices
	// that join the group receive the job execution even after the job has been
	// created.
	TargetSelection types.TargetSelection

	// Specifies the amount of time each device has to finish its execution of the
	// job. The timer is started when the job execution status is set to IN_PROGRESS .
	// If the job execution status is not set to another terminal state before the time
	// expires, it will be automatically set to TIMED_OUT .
	TimeoutConfig *types.TimeoutConfig

	noSmithyDocumentSerde
}

type CreateJobOutput struct {

	// The job description.
	Description *string

	// The job ARN.
	JobArn *string

	// The unique identifier you assigned to this job.
	JobId *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationCreateJobMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsRestjson1_serializeOpCreateJob{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsRestjson1_deserializeOpCreateJob{}, middleware.After)
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
	if err = addOpCreateJobValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opCreateJob(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opCreateJob(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "iot",
		OperationName: "CreateJob",
	}
}
