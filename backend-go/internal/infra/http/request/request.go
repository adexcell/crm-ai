package request

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/adexcell/crm-ai/internal/infra/http/response"
	"github.com/adexcell/crm-ai/internal/infra/validation"
)

// defaultMaxBodyBytes — максимальный допустимый размер тела JSON (1 MB).
const defaultMaxBodyBytes = 1 << 20 // 1048576 байт

// Decode декодирует входящее JSON-тело в структуру v, проверяет размер тела
// и выполняет валидацию по тегам validate:"...".
//
// Если декодирование или валидация провалились, функция сама отправляет
// соответствующий JSON-ответ с ошибкой в w и возвращает false.
func Decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, defaultMaxBodyBytes)

	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			response.Error(
				w,
				http.StatusRequestEntityTooLarge,
				"BODY_TOO_LARGE",
				"Request body is too large (maximum 1 MB)",
				nil,
			)
			return false
		}
		if errors.Is(err, io.EOF) {
			response.Error(
				w,
				http.StatusBadRequest,
				response.CodeInvalidBody,
				"Request body cannot be empty",
				nil,
			)
			return false
		}

		response.InvalidJSON(w)
		return false
	}

	if valErr := validation.Validate(v); valErr != nil {
		response.ValidationError(w, valErr.Fields)
		return false
	}
	return true
}
