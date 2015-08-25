package requesttree

import (
	"golang.org/x/net/context"

	"github.com/mondough/mercury"
	terrors "github.com/mondough/typhon/errors"
)

const (
	parentIdHeader = "Parent-Request-ID"
	reqIdCtxKey    = "Request-ID"

	currentServiceHeader  = "Current-Service"
	currentEndpointHeader = "Current-Endpoint"
	originServiceHeader   = "Origin-Service"
	originEndpointHeader  = "Origin-Endpoint"
)

type requestTreeMiddleware struct{}

func (m requestTreeMiddleware) ProcessClientRequest(req mercury.Request) mercury.Request {
	if req.Headers()[parentIdHeader] == "" { // Don't overwrite an exiting header
		if parentId, ok := req.Context().Value(reqIdCtxKey).(string); ok && parentId != "" {
			req.SetHeader(parentIdHeader, parentId)
		}
	}

	// Pass through the current service and endpoint as the origin of this request
	if svc, ok := req.Value(currentServiceHeader).(string); ok {
		req.SetHeader(originServiceHeader, svc)
	}
	if ept, ok := req.Value(currentEndpointHeader).(string); ok {
		req.SetHeader(originEndpointHeader, ept)
	}

	return req
}

func (m requestTreeMiddleware) ProcessClientResponse(rsp mercury.Response, ctx context.Context) mercury.Response {
	return rsp
}

func (m requestTreeMiddleware) ProcessClientError(err *terrors.Error, ctx context.Context) {}

func (m requestTreeMiddleware) ProcessServerRequest(req mercury.Request) (mercury.Request, mercury.Response) {
	req.SetContext(context.WithValue(req.Context(), reqIdCtxKey, req.Id()))
	if v := req.Headers()[parentIdHeader]; v != "" {
		req.SetContext(context.WithValue(req.Context(), parentIdCtxKey, v))
	}

	// Set the current service and endpoint into the context
	req.SetContext(context.WithValue(req.Context(), currentServiceHeader, req.Service()))
	req.SetContext(context.WithValue(req.Context(), currentEndpointHeader, req.Endpoint()))

	// Set the originator into the context
	req.SetContext(context.WithValue(req.Context(), originServiceHeader, req.Headers()[originServiceHeader]))
	req.SetContext(context.WithValue(req.Context(), originEndpointHeader, req.Headers()[originEndpointHeader]))

	return req, nil
}

func (m requestTreeMiddleware) ProcessServerResponse(rsp mercury.Response, ctx context.Context) mercury.Response {
	if v, ok := ctx.Value(parentIdCtxKey).(string); ok && v != "" && rsp != nil {
		rsp.SetHeader(parentIdHeader, v)
	}
	return rsp
}

func Middleware() requestTreeMiddleware {
	return requestTreeMiddleware{}
}
