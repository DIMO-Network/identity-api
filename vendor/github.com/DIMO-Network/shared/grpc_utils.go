package shared

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GrpcErrorToFiber useful anywhere calling a grpc underlying service and wanting to augment the error for fiber from grpc status codes
// meant to play nicely with the ErrorHandler in api.go that this would hand off errors to.
// msgAppend appends to error string, to eg. help if this gets logged
func GrpcErrorToFiber(err error, msgAppend string) error {
	if err == nil {
		return nil
	}
	// pull out grpc error status to then convert to fiber http equivalent
	errStatus, _ := status.FromError(err)

	switch errStatus.Code() {
	case codes.InvalidArgument:
		return fiber.NewError(fiber.StatusBadRequest, errStatus.Message()+". "+msgAppend)
	case codes.NotFound:
		return fiber.NewError(fiber.StatusNotFound, errStatus.Message()+". "+msgAppend)
	case codes.Aborted:
		return fiber.NewError(fiber.StatusConflict, errStatus.Message()+". "+msgAppend)
	case codes.Internal:
		return fiber.NewError(fiber.StatusInternalServerError, errStatus.Message()+". "+msgAppend)
	}
	return errors.Wrap(err, msgAppend)
}
