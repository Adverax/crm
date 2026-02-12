package soqlHttp

import (
	"encoding/json"
	"errors"
	"net/http"

	platformApi "github.com/proxima-research/proxima.crm.platform/api/openapi/platform"
	authModel "github.com/proxima-research/proxima.crm.platform/internal/access/auth/domain"
	"github.com/proxima-research/proxima.crm.platform/internal/data/soql/application/engine"
	soqlService "github.com/proxima-research/proxima.crm.platform/internal/data/soql/application/service"
	soqlModel "github.com/proxima-research/proxima.crm.platform/internal/data/soql/domain"
)

// SOQLApi handles SOQL query HTTP endpoints.
type SOQLApi struct {
	queryService soqlService.QueryService
}

// New creates a new SOQLApi instance.
func New(queryService soqlService.QueryService) *SOQLApi {
	return &SOQLApi{
		queryService: queryService,
	}
}

// ExecuteSOQLQuery handles POST /data/query requests.
func (api *SOQLApi) ExecuteSOQLQuery(w http.ResponseWriter, r *http.Request) {
	var request platformApi.SOQLQueryRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeSOQLError(w, http.StatusBadRequest, &platformApi.SOQLErrorResponse{
			ErrorType: platformApi.ParseError,
			Message:   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Get user from context
	authUser := authModel.AuthUserFromContext(r.Context())
	var userID int64
	if authUser != nil && authUser.User != nil {
		userID = int64(authUser.User.Id)
	}

	// Get cursor from request
	var cursor string
	if request.Cursor != nil {
		cursor = *request.Cursor
	}

	// Execute query
	// PageSize = 0 means "use LIMIT from SOQL query"
	params := &soqlModel.QueryParams{
		UserID: userID,
	}

	result, err := api.queryService.Execute(r.Context(), request.Query, cursor, params)
	if err != nil {
		code, errResp := mapErrorToSOQLError(err)
		writeSOQLError(w, code, errResp)
		return
	}

	// Map result to response
	response := mapResultToResponse(result)
	writeJSON(w, http.StatusOK, response)
}

// ExecuteSOQLQueryGet handles GET /data/query requests.
func (api *SOQLApi) ExecuteSOQLQueryGet(w http.ResponseWriter, r *http.Request, params platformApi.ExecuteSOQLQueryGetParams) {
	// Get user from context
	authUser := authModel.AuthUserFromContext(r.Context())
	var userID int64
	if authUser != nil && authUser.User != nil {
		userID = int64(authUser.User.Id)
	}

	// Get cursor from params
	var cursor string
	if params.Cursor != nil {
		cursor = *params.Cursor
	}

	// Execute query
	// PageSize = 0 means "use LIMIT from SOQL query"
	queryParams := &soqlModel.QueryParams{
		UserID: userID,
	}

	result, err := api.queryService.Execute(r.Context(), params.Q, cursor, queryParams)
	if err != nil {
		code, errResp := mapErrorToSOQLError(err)
		writeSOQLError(w, code, errResp)
		return
	}

	// Map result to response
	response := mapResultToResponse(result)
	writeJSON(w, http.StatusOK, response)
}

// mapErrorToSOQLError maps engine errors to SOQLErrorResponse with position info.
func mapErrorToSOQLError(err error) (int, *platformApi.SOQLErrorResponse) {
	resp := &platformApi.SOQLErrorResponse{
		Message: err.Error(),
	}

	// Try to extract detailed error info
	var parseErr *engine.ParseError
	var validationErr *engine.ValidationError
	var accessErr *engine.AccessError
	var limitErr *engine.LimitError
	var executionErr *engine.ExecutionError

	switch {
	case errors.As(err, &parseErr):
		resp.ErrorType = platformApi.ParseError
		resp.Message = parseErr.Message
		if parseErr.Pos.Line > 0 {
			resp.Position = &platformApi.SOQLErrorPosition{
				Line:   &parseErr.Pos.Line,
				Column: &parseErr.Pos.Column,
				Offset: &parseErr.Pos.Offset,
			}
		}
		return http.StatusBadRequest, resp

	case errors.As(err, &validationErr):
		resp.ErrorType = platformApi.ValidationError
		resp.Message = validationErr.Message
		errorCode := validationErr.Code.String()
		resp.ErrorCode = &errorCode
		if validationErr.Object != "" {
			resp.Object = &validationErr.Object
		}
		if validationErr.Field != "" {
			resp.Field = &validationErr.Field
		}
		if validationErr.Pos.Line > 0 {
			resp.Position = &platformApi.SOQLErrorPosition{
				Line:   &validationErr.Pos.Line,
				Column: &validationErr.Pos.Column,
				Offset: &validationErr.Pos.Offset,
			}
		}
		return http.StatusUnprocessableEntity, resp

	case errors.As(err, &accessErr):
		resp.ErrorType = platformApi.AccessError
		resp.Message = accessErr.Message
		if accessErr.Object != "" {
			resp.Object = &accessErr.Object
		}
		if accessErr.Field != "" {
			resp.Field = &accessErr.Field
		}
		return http.StatusForbidden, resp

	case errors.As(err, &limitErr):
		resp.ErrorType = platformApi.LimitError
		resp.Message = limitErr.Message
		return http.StatusBadRequest, resp

	case errors.As(err, &executionErr):
		resp.ErrorType = platformApi.ExecutionError
		resp.Message = executionErr.Message
		return http.StatusInternalServerError, resp

	// Fallback to domain errors
	case errors.Is(err, soqlModel.ErrInvalidQuery):
		resp.ErrorType = platformApi.ParseError
		return http.StatusBadRequest, resp

	case errors.Is(err, soqlModel.ErrSemanticError):
		resp.ErrorType = platformApi.ValidationError
		return http.StatusUnprocessableEntity, resp

	case errors.Is(err, soqlModel.ErrQueryTooComplex):
		resp.ErrorType = platformApi.LimitError
		return http.StatusBadRequest, resp

	case errors.Is(err, soqlModel.ErrInvalidCursor):
		resp.ErrorType = platformApi.ParseError
		resp.Message = "Invalid pagination cursor"
		return http.StatusBadRequest, resp

	default:
		resp.ErrorType = platformApi.ExecutionError
		resp.Message = "Internal error executing query"
		return http.StatusInternalServerError, resp
	}
}

// mapResultToResponse maps domain.QueryResult to platformApi.SOQLQueryResponse.
func mapResultToResponse(result *soqlModel.QueryResult) platformApi.SOQLQueryResponse {
	records := make([]map[string]interface{}, len(result.Records))
	for i, rec := range result.Records {
		records[i] = rec
	}

	response := platformApi.SOQLQueryResponse{
		TotalSize: result.TotalSize,
		Done:      result.Done,
		Records:   records,
	}

	if result.NextCursor != "" {
		response.NextCursor = &result.NextCursor
	}

	return response
}

// writeSOQLError writes a SOQL error response.
func writeSOQLError(w http.ResponseWriter, code int, resp *platformApi.SOQLErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(resp)
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}
