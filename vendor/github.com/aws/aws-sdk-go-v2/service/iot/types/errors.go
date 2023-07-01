// Code generated by smithy-go-codegen DO NOT EDIT.

package types

import (
	"fmt"
	smithy "github.com/aws/smithy-go"
)

// Unable to verify the CA certificate used to sign the device certificate you are
// attempting to register. This is happens when you have registered more than one
// CA certificate that has the same subject field and public key.
type CertificateConflictException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *CertificateConflictException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *CertificateConflictException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *CertificateConflictException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "CertificateConflictException"
	}
	return *e.ErrorCodeOverride
}
func (e *CertificateConflictException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The certificate operation is not allowed.
type CertificateStateException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *CertificateStateException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *CertificateStateException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *CertificateStateException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "CertificateStateException"
	}
	return *e.ErrorCodeOverride
}
func (e *CertificateStateException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The certificate is invalid.
type CertificateValidationException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *CertificateValidationException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *CertificateValidationException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *CertificateValidationException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "CertificateValidationException"
	}
	return *e.ErrorCodeOverride
}
func (e *CertificateValidationException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// A resource with the same name already exists.
type ConflictException struct {
	Message *string

	ErrorCodeOverride *string

	ResourceId *string

	noSmithyDocumentSerde
}

func (e *ConflictException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ConflictException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ConflictException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ConflictException"
	}
	return *e.ErrorCodeOverride
}
func (e *ConflictException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// A conflicting resource update exception. This exception is thrown when two
// pending updates cause a conflict.
type ConflictingResourceUpdateException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *ConflictingResourceUpdateException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ConflictingResourceUpdateException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ConflictingResourceUpdateException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ConflictingResourceUpdateException"
	}
	return *e.ErrorCodeOverride
}
func (e *ConflictingResourceUpdateException) ErrorFault() smithy.ErrorFault {
	return smithy.FaultClient
}

// You can't delete the resource because it is attached to one or more resources.
type DeleteConflictException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *DeleteConflictException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *DeleteConflictException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *DeleteConflictException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "DeleteConflictException"
	}
	return *e.ErrorCodeOverride
}
func (e *DeleteConflictException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The index is not ready.
type IndexNotReadyException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *IndexNotReadyException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *IndexNotReadyException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *IndexNotReadyException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "IndexNotReadyException"
	}
	return *e.ErrorCodeOverride
}
func (e *IndexNotReadyException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// An unexpected error has occurred.
type InternalException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *InternalException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *InternalException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *InternalException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "InternalException"
	}
	return *e.ErrorCodeOverride
}
func (e *InternalException) ErrorFault() smithy.ErrorFault { return smithy.FaultServer }

// An unexpected error has occurred.
type InternalFailureException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *InternalFailureException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *InternalFailureException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *InternalFailureException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "InternalFailureException"
	}
	return *e.ErrorCodeOverride
}
func (e *InternalFailureException) ErrorFault() smithy.ErrorFault { return smithy.FaultServer }

// Internal error from the service that indicates an unexpected error or that the
// service is unavailable.
type InternalServerException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *InternalServerException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *InternalServerException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *InternalServerException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "InternalServerException"
	}
	return *e.ErrorCodeOverride
}
func (e *InternalServerException) ErrorFault() smithy.ErrorFault { return smithy.FaultServer }

// The aggregation is invalid.
type InvalidAggregationException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *InvalidAggregationException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *InvalidAggregationException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *InvalidAggregationException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "InvalidAggregationException"
	}
	return *e.ErrorCodeOverride
}
func (e *InvalidAggregationException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The query is invalid.
type InvalidQueryException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *InvalidQueryException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *InvalidQueryException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *InvalidQueryException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "InvalidQueryException"
	}
	return *e.ErrorCodeOverride
}
func (e *InvalidQueryException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The request is not valid.
type InvalidRequestException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *InvalidRequestException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *InvalidRequestException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *InvalidRequestException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "InvalidRequestException"
	}
	return *e.ErrorCodeOverride
}
func (e *InvalidRequestException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The response is invalid.
type InvalidResponseException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *InvalidResponseException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *InvalidResponseException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *InvalidResponseException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "InvalidResponseException"
	}
	return *e.ErrorCodeOverride
}
func (e *InvalidResponseException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// An attempt was made to change to an invalid state, for example by deleting a
// job or a job execution which is "IN_PROGRESS" without setting the force
// parameter.
type InvalidStateTransitionException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *InvalidStateTransitionException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *InvalidStateTransitionException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *InvalidStateTransitionException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "InvalidStateTransitionException"
	}
	return *e.ErrorCodeOverride
}
func (e *InvalidStateTransitionException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// A limit has been exceeded.
type LimitExceededException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *LimitExceededException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *LimitExceededException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *LimitExceededException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "LimitExceededException"
	}
	return *e.ErrorCodeOverride
}
func (e *LimitExceededException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The policy documentation is not valid.
type MalformedPolicyException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *MalformedPolicyException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *MalformedPolicyException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *MalformedPolicyException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "MalformedPolicyException"
	}
	return *e.ErrorCodeOverride
}
func (e *MalformedPolicyException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The resource is not configured.
type NotConfiguredException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *NotConfiguredException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *NotConfiguredException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *NotConfiguredException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "NotConfiguredException"
	}
	return *e.ErrorCodeOverride
}
func (e *NotConfiguredException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The registration code is invalid.
type RegistrationCodeValidationException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *RegistrationCodeValidationException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *RegistrationCodeValidationException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *RegistrationCodeValidationException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "RegistrationCodeValidationException"
	}
	return *e.ErrorCodeOverride
}
func (e *RegistrationCodeValidationException) ErrorFault() smithy.ErrorFault {
	return smithy.FaultClient
}

// The resource already exists.
type ResourceAlreadyExistsException struct {
	Message *string

	ErrorCodeOverride *string

	ResourceId  *string
	ResourceArn *string

	noSmithyDocumentSerde
}

func (e *ResourceAlreadyExistsException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ResourceAlreadyExistsException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ResourceAlreadyExistsException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ResourceAlreadyExistsException"
	}
	return *e.ErrorCodeOverride
}
func (e *ResourceAlreadyExistsException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The specified resource does not exist.
type ResourceNotFoundException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *ResourceNotFoundException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ResourceNotFoundException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ResourceNotFoundException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ResourceNotFoundException"
	}
	return *e.ErrorCodeOverride
}
func (e *ResourceNotFoundException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The resource registration failed.
type ResourceRegistrationFailureException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *ResourceRegistrationFailureException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ResourceRegistrationFailureException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ResourceRegistrationFailureException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ResourceRegistrationFailureException"
	}
	return *e.ErrorCodeOverride
}
func (e *ResourceRegistrationFailureException) ErrorFault() smithy.ErrorFault {
	return smithy.FaultClient
}

// A limit has been exceeded.
type ServiceQuotaExceededException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *ServiceQuotaExceededException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ServiceQuotaExceededException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ServiceQuotaExceededException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ServiceQuotaExceededException"
	}
	return *e.ErrorCodeOverride
}
func (e *ServiceQuotaExceededException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The service is temporarily unavailable.
type ServiceUnavailableException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *ServiceUnavailableException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ServiceUnavailableException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ServiceUnavailableException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ServiceUnavailableException"
	}
	return *e.ErrorCodeOverride
}
func (e *ServiceUnavailableException) ErrorFault() smithy.ErrorFault { return smithy.FaultServer }

// The Rule-SQL expression can't be parsed correctly.
type SqlParseException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *SqlParseException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *SqlParseException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *SqlParseException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "SqlParseException"
	}
	return *e.ErrorCodeOverride
}
func (e *SqlParseException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// This exception occurs if you attempt to start a task with the same task-id as
// an existing task but with a different clientRequestToken.
type TaskAlreadyExistsException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *TaskAlreadyExistsException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *TaskAlreadyExistsException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *TaskAlreadyExistsException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "TaskAlreadyExistsException"
	}
	return *e.ErrorCodeOverride
}
func (e *TaskAlreadyExistsException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The rate exceeds the limit.
type ThrottlingException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *ThrottlingException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ThrottlingException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ThrottlingException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ThrottlingException"
	}
	return *e.ErrorCodeOverride
}
func (e *ThrottlingException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// You can't revert the certificate transfer because the transfer is already
// complete.
type TransferAlreadyCompletedException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *TransferAlreadyCompletedException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *TransferAlreadyCompletedException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *TransferAlreadyCompletedException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "TransferAlreadyCompletedException"
	}
	return *e.ErrorCodeOverride
}
func (e *TransferAlreadyCompletedException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// You can't transfer the certificate because authorization policies are still
// attached.
type TransferConflictException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *TransferConflictException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *TransferConflictException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *TransferConflictException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "TransferConflictException"
	}
	return *e.ErrorCodeOverride
}
func (e *TransferConflictException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// You are not authorized to perform this operation.
type UnauthorizedException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *UnauthorizedException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *UnauthorizedException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *UnauthorizedException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "UnauthorizedException"
	}
	return *e.ErrorCodeOverride
}
func (e *UnauthorizedException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The request is not valid.
type ValidationException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *ValidationException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ValidationException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ValidationException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "ValidationException"
	}
	return *e.ErrorCodeOverride
}
func (e *ValidationException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// An exception thrown when the version of an entity specified with the
// expectedVersion parameter does not match the latest version in the system.
type VersionConflictException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *VersionConflictException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *VersionConflictException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *VersionConflictException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "VersionConflictException"
	}
	return *e.ErrorCodeOverride
}
func (e *VersionConflictException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The number of policy versions exceeds the limit.
type VersionsLimitExceededException struct {
	Message *string

	ErrorCodeOverride *string

	noSmithyDocumentSerde
}

func (e *VersionsLimitExceededException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *VersionsLimitExceededException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *VersionsLimitExceededException) ErrorCode() string {
	if e == nil || e.ErrorCodeOverride == nil {
		return "VersionsLimitExceededException"
	}
	return *e.ErrorCodeOverride
}
func (e *VersionsLimitExceededException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }
