package ffmpeg

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	ErrUnknown                       = "UNKNOWN_ERROR"
	ErrInvalidData                   = "INVALID_DATA"
	ErrUnsupportedInputFormat        = "UNSUPPORTED_INPUT_FORMAT"
	ErrCudaUnknown                   = "CUDA_ERROR_UNKNOWN"
	ErrFunctionNotImplemented        = "FUNCTION_NOT_IMPLEMENTED"
	ErrBitDepthNotSupported          = "BIT_DEPTH_NOT_SUPPORTED"
	ErrNonMonotonousDTS              = "NON_MONOTONOUS_DTS"
	ErrCudaInvalidHandle             = "CUDA_ERROR_INVALID_HANDLE"
	ErrCudaNotSupported              = "CUDA_ERROR_NOT_SUPPORTED"
	ErrGeneric                       = "GENERIC_ERROR_IN_EXTERNAL_LIBRARY"
	ErrInvalidNALUnitSize            = "INVALID_NAL_UNIT_SIZE"
	ErrACNotAllocated                = "CHANNEL_ELEMENT_NOT_ALLOCATED"
	ErrPredictionIsNotAllovedInAACLC = "PREDICTION_IS_NOT_ALLOWED_IN_AAC_LC"
	ErrCorruptInput                  = "CORRUPT_INPUT"
)

type regexpErr struct {
	re      *regexp.Regexp
	code    string
	message string
}

// order matters = parse order
var regexpErrors = []regexpErr{
	{
		re:      regexp.MustCompile(`Unsupported input format:(.*)`),
		code:    ErrUnsupportedInputFormat,
		message: "", // from regex match
	},
	{
		re:      regexp.MustCompile(`Bit depth (\d*) is not supported`),
		code:    ErrBitDepthNotSupported,
		message: "", // from regex match
	},
	{
		re:      regexp.MustCompile(`channel element (.*?) is not allocated`),
		code:    ErrACNotAllocated,
		message: "", // from regex match
	},
	{
		re:      regexp.MustCompile(`Non-monotonic DTS; previous: (.*?), current: (.*?);`),
		code:    ErrNonMonotonousDTS,
		message: "",
	},
	{
		re:      regexp.MustCompile(`Packet corrupt \(stream = (\d+), dts = (\d+)\)`),
		code:    ErrCorruptInput,
		message: "",
	},
	{
		re:      regexp.MustCompile(`corrupt input packet in stream (\d+)`),
		code:    ErrCorruptInput,
		message: "",
	},
}

var knownErrors map[string]string = map[string]string{
	"Prediction is not allowed in AAC-LC":                                  ErrPredictionIsNotAllovedInAACLC,
	"Invalid NAL unit size":                                                ErrInvalidNALUnitSize,
	"Non-monotonous DTS in output stream":                                  ErrNonMonotonousDTS,
	"CUDA_ERROR_UNKNOWN: unknown error":                                    ErrCudaUnknown,
	"Failed to inject frame into filter network: Function not implemented": ErrFunctionNotImplemented,
	"CUDA_ERROR_INVALID_HANDLE":                                            ErrCudaInvalidHandle,
	"CUDA_ERROR_NOT_SUPPORTED":                                             ErrCudaNotSupported,
	"Generic error in an external library":                                 ErrGeneric,
	"Invalid data found when processing input":                             ErrInvalidData,
	//"Non-monotonic DTS":                                                    ErrNonMonotonousDTS,
}

type Error struct {
	Cmd       string
	Code      string
	Path      string
	Message   string
	Timestamp time.Time
}

func (e *Error) GetTimestamp() string {
	var ts string = fmt.Sprintf("%v", e.Timestamp.Format("15:04:05.21"))
	if e.Timestamp.Equal(time.Time{}) {
		ts = "unknown"
	}
	return ts
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("[%s] (time=%s) %s ", e.Code, e.GetTimestamp(), e.Message)
}

func parseTranscodingError(msg, src string, lastTs time.Time) error {
	for _, regexpErr := range regexpErrors {
		if match := regexpErr.re.FindString(msg); match != "" {
			return &Error{
				Code:      regexpErr.code,
				Timestamp: lastTs,
				Path:      src,
				Message:   match,
			}
		}
	}

	for m, code := range knownErrors {
		if strings.Contains(msg, m) {
			return &Error{
				Code:      code,
				Timestamp: lastTs,
				Path:      src,
				Message:   m,
			}
		}
	}

	return nil
}

func parseError(msg, src string, lastTs time.Time) error {
	for _, regexpErr := range regexpErrors {
		if match := regexpErr.re.FindString(msg); match != "" {
			return &Error{
				Code:      regexpErr.code,
				Timestamp: lastTs,
				Path:      src,
				Message:   match,
			}
		}
	}

	for m, code := range knownErrors {
		if strings.Contains(msg, m) {
			return &Error{
				Code:      code,
				Timestamp: lastTs,
				Path:      src,
				Message:   m,
			}
		}
	}

	return &Error{
		Code:    ErrUnknown,
		Path:    src,
		Message: msg,
	}
}
