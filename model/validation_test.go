// Copyright 2022 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package model

import (
	"errors"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
)

var ErrCannotBeEmpty = errors.New("cannot be empty")

func TestValidationErrors(t *testing.T) {
	t.Parallel()

	assert.False(t, IsValidationError(errors.New("generic error")))

	var err *ValidationError
	assert.False(t, IsValidationError(err))
	assert.False(t, errors.Is(err, errors.New("generic error")))
	assert.EqualError(t, err, "")
	assert.Nil(t, err.Unwrap())

	// Empty ValidationError
	err = new(ValidationError)
	assert.EqualError(t, err, "")
	err.Field = "foo"
	assert.EqualError(t, err, "")

	// err is now a valid error
	err.Reason = ErrCannotBeEmpty
	assert.EqualError(t, err, err.Field+": "+ErrCannotBeEmpty.Error())
	assert.True(t, IsValidationError(err))

	// Compare underlying error
	assert.ErrorIs(t, err, ErrCannotBeEmpty)

	// Is compares both Reason and Field
	assert.ErrorIs(t, err,
		&ValidationError{Field: "foo", Reason: ErrCannotBeEmpty},
	)
	assert.False(t, errors.Is(err,
		&ValidationError{Field: "bar", Reason: ErrCannotBeEmpty},
	))

	// Verify nil check
	assert.False(t, err.Is((*ValidationError)(nil)))

	assert.EqualError(t,
		&ValidationError{Reason: ErrCannotBeEmpty},
		ErrCannotBeEmpty.Error(),
	)

	// Empty ValidationErrors
	var errs ValidationErrors
	assert.EqualError(t, errs, "")
	assert.False(t, IsValidationError(errs))

	// Make ValidationErrors valid
	errs = append(errs, err)
	assert.ErrorIs(t, errs, err)
	assert.ErrorIs(t, errs, ErrCannotBeEmpty)
	assert.False(t, errors.Is(errs, errors.New("generic error")))
	assert.EqualError(t, errs, err.Error())

	// Check ozzo validation errors evaluates to ValidationError
	var ozzoError validation.Error
	assert.False(t, IsValidationError(ozzoError))
	ozzoError = validation.NewError("foo", "bar")
	assert.True(t, IsValidationError(ozzoError))
	var ozzoErrors validation.Errors
	assert.False(t, IsValidationError(ozzoErrors))
	ozzoErrors = validation.Errors{"foo": ozzoError}
	assert.True(t, IsValidationError(ozzoErrors))
}
