package response

// Общесистемные машиночитаемые коды ошибок платформы
const (
	CodeInvalidBody     = "INVALID_BODY"      // Невалидный JSON в теле запроса
	CodeValidationError = "VALIDATION_ERROR"  // Ошибка валидации DTO полей
	CodeInternalError   = "INTERNAL_ERROR"    // Непредвиденная внутренняя ошибка сервера (500)
	CodeUnauthorized    = "UNAUTHORIZED"      // Токен отсутствует или недействителен (401)
	CodeForbidden       = "FORBIDDEN"         // Недостаточно прав для операции (403)
	CodeNotFound        = "NOT_FOUND"         // Ресурс по ID/маршруту не найден (404)
	CodeTooManyRequests = "TOO_MANY_REQUESTS" // Превышен лимит запросов (429)
)
