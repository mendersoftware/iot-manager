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
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type ValidationError struct {
	Field  string
	Reason error
}

func (err *ValidationError) Error() string {
	if err == nil || err.Reason == nil {
		return ""
	}
	if err.Field == "" {
		return err.Reason.Error()
	}
	return err.Field + ": " + err.Reason.Error()
}

func (err *ValidationError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Reason
}

func (err *ValidationError) Is(target error) bool {
	if err == nil {
		return false
	}
	if verr, ok := target.(*ValidationError); ok && verr != nil {
		if err.Field != verr.Field {
			return false
		}
		return errors.Is(err.Reason, verr.Unwrap())
	}
	return errors.Is(err.Reason, target)
}

type ValidationErrors []*ValidationError

func (errs ValidationErrors) Error() string {
	if len(errs) <= 0 {
		return ""
	}
	var errMsgs []string = make([]string, len(errs))
	for i := range errs {
		errMsgs[i] = errs[i].Error()
	}
	return strings.Join(errMsgs, "; ")
}

func (errs ValidationErrors) Is(err error) bool {
	for _, verr := range errs {
		if errors.Is(verr, err) {
			return true
		}
	}
	return false
}

func IsValidationError(err error) bool {
	switch t := err.(type) {
	case nil:
		return false

	case *ValidationError:
		if t == nil {
			return false
		}

	case ValidationErrors:
		if t == nil {
			return false
		}

	case validation.Error:

	case validation.Errors:
		if len(t) == 0 {
			return false
		}
	default:
		return false
	}
	return true
}
