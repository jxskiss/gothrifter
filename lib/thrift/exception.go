package thrift

const (
	UNKNOWN_APPLICATION_EXCEPTION  = 0
	UNKNOWN_METHOD                 = 1
	INVALID_MESSAGE_TYPE_EXCEPTION = 2
	WRONG_METHOD_NAME              = 3
	BAD_SEQUENCE_ID                = 4
	MISSING_RESULT                 = 5
	INTERNAL_ERROR                 = 6
	PROTOCOL_ERROR                 = 7

	UNKNOWN_TRANSPORT_EXCEPTION = 30
	NOT_OPEN                    = 31
	ALREADY_OPEN                = 32
	TIMED_OUT                   = 33
	END_OF_FILE                 = 34
	INTERRUPTED                 = 35
	BAD_ARGS                    = 36
	CORRUPTED_DATA              = 37
	NOT_SUPPORTED               = 39
	INVALID_STATE               = 40
	INVALID_FRAME_SIZE          = 41
	SSL_ERROR                   = 42
	COULD_NOT_BIND              = 43
	SASL_HANDSHAKE_TIMEOUT      = 44
	NETWORK_ERROR               = 45

	UNKNOWN_PROTOCOL_EXCEPTION = 60
	INVALID_DATA               = 61
	NEGATIVE_SIZE              = 62
	SIZE_LIMIT                 = 63
	BAD_VERSION                = 64
	NOT_IMPLEMENTED            = 65
	DEPTH_LIMIT                = 66
)

type Exception interface {
	error
}

type ApplicationException struct {
	m string // 1
	t int32  // 2
}

func NewApplicationException(t int32, m string) error {
	return &ApplicationException{t: t, m: m}
}

func FromErr(err error) *ApplicationException {
	ex, ok := err.(*ApplicationException)
	if ok {
		return ex
	}
	return &ApplicationException{m: err.Error()}
}

// TypeID returns the exception type.
func (e *ApplicationException) TypeID() int32 {
	return e.t
}

// Error implements the error interface.
func (e *ApplicationException) Error() string {
	return e.m
}

//func (e *ApplicationException) Equal(o *ApplicationException) bool {
//	if e == nil || o == nil {
//		return false
//	}
//	return e.Message == o.Message && e.m == o.m
//}

func (e *ApplicationException) Read(r Reader) error {
	if _, err := r.ReadStructBegin(); err != nil {
		return err
	}
	for {
		_, ttype, fieldId, err := r.ReadFieldBegin()
		if err != nil {
			return err
		}
		if ttype == STOP {
			return nil
		}
		switch fieldId {
		case 1:
			if e.m, err = r.ReadString(); err != nil {
				return err
			}
		case 2:
			if e.t, err = r.ReadI32(); err != nil {
				return err
			}
		default:
			if err = SkipDefaultDepth(r, ttype); err != nil {
				return err
			}
		}
		if err = r.ReadFieldEnd(); err != nil {
			return err
		}
	}
}

func (e *ApplicationException) Write(w Writer) (err error) {
	if err = w.WriteStructBegin("ApplicationException"); err != nil {
		return err
	}
	if len(e.m) > 0 {
		if err = w.WriteFieldBegin("message", STRING, 1); err != nil {
			return err
		}
		if err = w.WriteString(e.Error()); err != nil {
			return err
		}
		if err = w.WriteFieldEnd(); err != nil {
			return err
		}
	}
	if err = w.WriteFieldBegin("type", I32, 2); err != nil {
		return err
	}
	if err = w.WriteI32(e.t); err != nil {
		return err
	}
	if err = w.WriteFieldEnd(); err != nil {
		return err
	}
	if err = w.WriteFieldStop(); err != nil {
		return err
	}
	return w.WriteStructEnd()
}
